package authsvc

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"net"
	"net/netip"
	"strings"
	"time"

	authstore "github.com/bitik/backend/internal/store/auth"
	rbacstore "github.com/bitik/backend/internal/store/rbac"
	systemstore "github.com/bitik/backend/internal/store/system"
	userstore "github.com/bitik/backend/internal/store/users"

	"github.com/bitik/backend/internal/config"
	"github.com/bitik/backend/internal/jwtutil"
	"github.com/bitik/backend/internal/pgxutil"
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	cfg         config.Config
	log         *zap.Logger
	pool        *pgxpool.Pool
	redis       *redis.Client
	auth        *authstore.Queries
	users       *userstore.Queries
	rbacQ       *rbacstore.Queries
	sys         *systemstore.Queries
	emailSender EmailSender
	otpSender   OTPSender
	Casbin      *casbin.Enforcer
}

func NewService(cfg config.Config, log *zap.Logger, pool *pgxpool.Pool, r *redis.Client, enf *casbin.Enforcer, opts ...Option) *Service {
	s := &Service{
		cfg:    cfg,
		log:    log,
		pool:   pool,
		redis:  r,
		auth:   authstore.New(pool),
		users:  userstore.New(pool),
		rbacQ:  rbacstore.New(pool),
		sys:    systemstore.New(pool),
		Casbin: enf,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

type TokenPair struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresInSeconds int64  `json:"expires_in"`
	TokenType        string `json:"token_type"`
}

func (s *Service) hashOpaque(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func pgText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func (s *Service) audit(ctx context.Context, actor uuid.UUID, action, entityType string, entityID *uuid.UUID, ip *netip.Addr, ua string) {
	var eid pgtype.UUID
	if entityID != nil {
		eid = pgxutil.UUID(*entityID)
	}
	_, err := s.sys.CreateAuditLog(ctx, systemstore.CreateAuditLogParams{
		ActorUserID: pgxutil.UUID(actor),
		Action:      action,
		EntityType:  pgText(entityType),
		EntityID:    eid,
		OldValues:   nil,
		NewValues:   nil,
		IpAddress:   ip,
		UserAgent:   pgText(ua),
	})
	if err != nil && s.log != nil {
		s.log.Warn("audit_log_failed", zap.Error(err))
	}
}

type clientMeta struct {
	IP        *netip.Addr
	UserAgent string
	DeviceID  string
	Platform  string
	PushToken string
}

func clientMetaFromGin(c *gin.Context) clientMeta {
	ipStr := c.ClientIP()
	var ipPtr *netip.Addr
	if ipStr != "" {
		if a, err := netip.ParseAddr(ipStr); err == nil {
			ipPtr = &a
		} else if host, _, err := net.SplitHostPort(c.Request.RemoteAddr); err == nil {
			if a, err := netip.ParseAddr(host); err == nil {
				ipPtr = &a
			}
		}
	}
	return clientMeta{
		IP:        ipPtr,
		UserAgent: c.Request.UserAgent(),
		DeviceID:  strings.TrimSpace(c.GetHeader("X-Device-Id")),
		Platform:  strings.TrimSpace(c.GetHeader("X-Platform")),
		PushToken: strings.TrimSpace(c.GetHeader("X-Push-Token")),
	}
}

func (s *Service) roleNames(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return s.rbacQ.ListRoleNamesForUser(ctx, pgxutil.UUID(userID))
}

func (s *Service) issueTokenPair(ctx context.Context, userID uuid.UUID, roles []string) (*TokenPair, error) {
	ttl := s.cfg.Auth.AccessTokenTTL
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	at, err := jwtutil.Sign(s.cfg.Auth.JWTSecret, s.cfg.Auth.JWTIssuer, userID, roles, ttl)
	if err != nil {
		return nil, err
	}
	rawRefresh, err := newOpaqueToken()
	if err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken:      at,
		RefreshToken:     rawRefresh,
		ExpiresInSeconds: int64(ttl.Seconds()),
		TokenType:        "Bearer",
	}, nil
}

func newOpaqueToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *Service) createSessionAndRefresh(ctx context.Context, userID uuid.UUID, meta clientMeta, rawRefresh string) (pgtype.UUID, error) {
	sess, err := s.auth.CreateUserSession(ctx, authstore.CreateUserSessionParams{
		UserID:    pgxutil.UUID(userID),
		DeviceID:  pgText(meta.DeviceID),
		UserAgent: pgText(meta.UserAgent),
		Platform:  pgText(meta.Platform),
		PushToken: pgText(meta.PushToken),
		IpText:    pgText(clientIPString(meta.IP)),
	})
	if err != nil {
		return pgtype.UUID{}, err
	}
	rtTTL := s.cfg.Auth.RefreshTokenTTL
	if rtTTL <= 0 {
		rtTTL = 720 * time.Hour
	}
	exp := pgtype.Timestamptz{Time: time.Now().Add(rtTTL), Valid: true}
	_, err = s.auth.CreateRefreshToken(ctx, authstore.CreateRefreshTokenParams{
		UserID:    pgxutil.UUID(userID),
		TokenHash: s.hashOpaque(rawRefresh),
		ExpiresAt: exp,
		SessionID: sess.ID,
	})
	if err != nil {
		return pgtype.UUID{}, err
	}
	if meta.DeviceID != "" {
		_, _ = s.users.CreateUserDevice(ctx, userstore.CreateUserDeviceParams{
			UserID:     pgxutil.UUID(userID),
			DeviceID:   meta.DeviceID,
			Platform:   firstNonEmpty(meta.Platform, "unknown"),
			AppVersion: pgtype.Text{},
			PushToken:  pgText(meta.PushToken),
		})
	}
	return sess.ID, nil
}

func clientIPString(ip *netip.Addr) string {
	if ip == nil {
		return ""
	}
	return ip.String()
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func (s *Service) userActive(u userstore.User) error {
	if u.DeletedAt.Valid {
		return errors.New("deleted")
	}
	switch userStatusString(u.Status) {
	case "active":
		return nil
	case "banned":
		return errors.New("banned")
	case "inactive":
		return errors.New("inactive")
	default:
		return errors.New("inactive")
	}
}

func userStatusString(status any) string {
	switch v := status.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return ""
	}
}

func (s *Service) logSecret(name string, fields ...zap.Field) {
	if s.log == nil || !s.cfg.Auth.LogSecrets {
		return
	}
	s.log.Info(name, fields...)
}

// --- Register / login / refresh ---

func (s *Service) Register(ctx context.Context, email, password string, displayName *string, meta clientMeta) (*TokenPair, uuid.UUID, error) {
	if len(password) < 8 {
		return nil, uuid.Nil, errors.New("weak_password")
	}
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return nil, uuid.Nil, errors.New("email_required")
	}
	if _, err := s.users.GetUserByEmail(ctx, pgText(email)); err == nil {
		return nil, uuid.Nil, errors.New("email_taken")
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, uuid.Nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, uuid.Nil, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, uuid.Nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	uq := userstore.New(tx)
	rq := rbacstore.New(tx)
	u, err := uq.CreateUser(ctx, userstore.CreateUserParams{
		Email:        pgText(email),
		Phone:        pgtype.Text{},
		PasswordHash: pgText(string(hash)),
		Column4:      nil,
	})
	if err != nil {
		return nil, uuid.Nil, err
	}
	uid, ok := pgxutil.ToUUID(u.ID)
	if !ok {
		return nil, uuid.Nil, errors.New("user_id")
	}
	dn := pgtype.Text{}
	if displayName != nil && strings.TrimSpace(*displayName) != "" {
		dn = pgText(strings.TrimSpace(*displayName))
	}
	if _, err := uq.UpdateUserProfile(ctx, userstore.UpdateUserProfileParams{
		UserID:      u.ID,
		FirstName:   pgtype.Text{},
		LastName:    pgtype.Text{},
		DisplayName: dn,
		AvatarUrl:   pgtype.Text{},
		Gender:      pgtype.Text{},
		Birthdate:   pgtype.Date{},
		Column8:     nil,
		CountryCode: pgtype.Text{},
		Timezone:    pgtype.Text{},
	}); err != nil {
		return nil, uuid.Nil, err
	}
	if err := rq.AssignUserRoleByRoleName(ctx, rbacstore.AssignUserRoleByRoleNameParams{
		UserID:   u.ID,
		RoleName: "buyer",
	}); err != nil {
		return nil, uuid.Nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, uuid.Nil, err
	}
	rawTok, err := s.createEmailVerifyToken(ctx, uid, email)
	if err == nil {
		if s.emailSender != nil {
			if sendErr := s.emailSender.SendEmailVerification(ctx, email, rawTok); sendErr != nil {
				return nil, uuid.Nil, sendErr
			}
		} else {
			s.logSecret("email_verification_token", zap.String("email", email), zap.String("token", rawTok))
		}
	}
	roles, err := s.roleNames(ctx, uid)
	if err != nil {
		return nil, uuid.Nil, err
	}
	pair, err := s.issueTokenPair(ctx, uid, roles)
	if err != nil {
		return nil, uuid.Nil, err
	}
	if _, err := s.createSessionAndRefresh(ctx, uid, meta, pair.RefreshToken); err != nil {
		return nil, uuid.Nil, err
	}
	s.audit(ctx, uid, "auth.register", "user", &uid, meta.IP, meta.UserAgent)
	return pair, uid, nil
}

func (s *Service) createEmailVerifyToken(ctx context.Context, userID uuid.UUID, email string) (string, error) {
	raw, err := newOpaqueToken()
	if err != nil {
		return "", err
	}
	ttl := s.cfg.Auth.EmailVerifyTokenTTL
	if ttl <= 0 {
		ttl = 48 * time.Hour
	}
	exp := pgtype.Timestamptz{Time: time.Now().Add(ttl), Valid: true}
	_, err = s.auth.CreateEmailVerificationToken(ctx, authstore.CreateEmailVerificationTokenParams{
		UserID:    pgxutil.UUID(userID),
		Email:     email,
		TokenHash: s.hashOpaque(raw),
		ExpiresAt: exp,
	})
	if err != nil {
		return "", err
	}
	return raw, nil
}

func (s *Service) Login(ctx context.Context, email, password string, meta clientMeta) (*TokenPair, uuid.UUID, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if s.isLoginLocked(ctx, email) {
		return nil, uuid.Nil, errors.New("login_locked")
	}
	u, err := s.users.GetUserByEmail(ctx, pgText(email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.recordLoginFailure(ctx, email)
			return nil, uuid.Nil, errors.New("invalid_credentials")
		}
		return nil, uuid.Nil, err
	}
	if err := s.userActive(u); err != nil {
		return nil, uuid.Nil, err
	}
	if !u.PasswordHash.Valid {
		s.recordLoginFailure(ctx, email)
		return nil, uuid.Nil, errors.New("invalid_credentials")
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash.String), []byte(password)) != nil {
		s.recordLoginFailure(ctx, email)
		return nil, uuid.Nil, errors.New("invalid_credentials")
	}
	s.clearLoginFailures(ctx, email)
	uid, ok := pgxutil.ToUUID(u.ID)
	if !ok {
		return nil, uuid.Nil, errors.New("user_id")
	}
	roles, err := s.roleNames(ctx, uid)
	if err != nil {
		return nil, uuid.Nil, err
	}
	pair, err := s.issueTokenPair(ctx, uid, roles)
	if err != nil {
		return nil, uuid.Nil, err
	}
	if _, err := s.createSessionAndRefresh(ctx, uid, meta, pair.RefreshToken); err != nil {
		return nil, uuid.Nil, err
	}
	_ = s.users.UpdateUserLastLogin(ctx, u.ID)
	s.audit(ctx, uid, "auth.login", "user", &uid, meta.IP, meta.UserAgent)
	return pair, uid, nil
}

func (s *Service) Refresh(ctx context.Context, rawRefresh string, meta clientMeta) (*TokenPair, error) {
	rawRefresh = strings.TrimSpace(rawRefresh)
	if rawRefresh == "" {
		return nil, errors.New("missing_refresh")
	}
	h := s.hashOpaque(rawRefresh)
	row, err := s.auth.GetRefreshTokenByHash(ctx, h)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("invalid_refresh")
		}
		return nil, err
	}
	if row.RevokedAt.Valid {
		_ = s.auth.RevokeAllRefreshTokensForUser(ctx, row.UserID)
		_ = s.auth.RevokeAllUserSessions(ctx, row.UserID)
		uid, _ := pgxutil.ToUUID(row.UserID)
		s.audit(ctx, uid, "auth.refresh_reuse", "user", &uid, meta.IP, meta.UserAgent)
		return nil, errors.New("refresh_reuse")
	}
	if row.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("expired_refresh")
	}
	uid, ok := pgxutil.ToUUID(row.UserID)
	if !ok {
		return nil, errors.New("user_id")
	}
	u, err := s.users.GetUserByID(ctx, row.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("invalid_refresh")
		}
		return nil, err
	}
	if err := s.userActive(u); err != nil {
		return nil, err
	}
	if row.SessionID.Valid {
		sess, err := s.auth.GetUserSessionByID(ctx, authstore.GetUserSessionByIDParams{
			ID:     row.SessionID,
			UserID: row.UserID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, errors.New("invalid_refresh")
			}
			return nil, err
		}
		if sess.RevokedAt.Valid {
			return nil, errors.New("invalid_refresh")
		}
	}
	if err := s.auth.RevokeRefreshToken(ctx, h); err != nil {
		return nil, err
	}
	roles, err := s.roleNames(ctx, uid)
	if err != nil {
		return nil, err
	}
	pair, err := s.issueTokenPair(ctx, uid, roles)
	if err != nil {
		return nil, err
	}
	rtTTL := s.cfg.Auth.RefreshTokenTTL
	if rtTTL <= 0 {
		rtTTL = 720 * time.Hour
	}
	exp := pgtype.Timestamptz{Time: time.Now().Add(rtTTL), Valid: true}
	_, err = s.auth.CreateRefreshToken(ctx, authstore.CreateRefreshTokenParams{
		UserID:    row.UserID,
		TokenHash: s.hashOpaque(pair.RefreshToken),
		ExpiresAt: exp,
		SessionID: row.SessionID,
	})
	if err != nil {
		return nil, err
	}
	if row.SessionID.Valid {
		_ = s.auth.TouchUserSession(ctx, authstore.TouchUserSessionParams{
			ID:     row.SessionID,
			UserID: row.UserID,
		})
	}
	s.audit(ctx, uid, "auth.refresh", "user", &uid, meta.IP, meta.UserAgent)
	return pair, nil
}

func (s *Service) Logout(ctx context.Context, rawRefresh string, meta clientMeta) error {
	rawRefresh = strings.TrimSpace(rawRefresh)
	if rawRefresh == "" {
		return errors.New("missing_refresh")
	}
	h := s.hashOpaque(rawRefresh)
	row, err := s.auth.GetRefreshTokenByHash(ctx, h)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	if err := s.auth.RevokeRefreshToken(ctx, h); err != nil {
		return err
	}
	uid, ok := pgxutil.ToUUID(row.UserID)
	if ok {
		s.audit(ctx, uid, "auth.logout", "user", &uid, meta.IP, meta.UserAgent)
	}
	return nil
}

func (s *Service) ForgotPassword(ctx context.Context, email string, meta clientMeta) error {
	email = strings.TrimSpace(strings.ToLower(email))
	u, err := s.users.GetUserByEmail(ctx, pgText(email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	raw, err := newOpaqueToken()
	if err != nil {
		return err
	}
	ttl := s.cfg.Auth.PasswordResetTokenTTL
	if ttl <= 0 {
		ttl = time.Hour
	}
	exp := pgtype.Timestamptz{Time: time.Now().Add(ttl), Valid: true}
	if _, err := s.auth.CreatePasswordResetToken(ctx, authstore.CreatePasswordResetTokenParams{
		UserID:    u.ID,
		TokenHash: s.hashOpaque(raw),
		ExpiresAt: exp,
	}); err != nil {
		return err
	}
	uid, _ := pgxutil.ToUUID(u.ID)
	s.audit(ctx, uid, "auth.forgot_password", "user", &uid, meta.IP, meta.UserAgent)
	if s.emailSender != nil {
		if err := s.emailSender.SendPasswordReset(ctx, email, raw); err != nil {
			return err
		}
	} else {
		s.logSecret("password_reset_token", zap.String("email", email), zap.String("token", raw))
	}
	return nil
}

func (s *Service) ResetPassword(ctx context.Context, token, newPassword string, meta clientMeta) error {
	if len(newPassword) < 8 {
		return errors.New("weak_password")
	}
	h := s.hashOpaque(strings.TrimSpace(token))
	row, err := s.auth.ConsumePasswordResetToken(ctx, h)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("invalid_token")
		}
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.users.UpdateUserPasswordHash(ctx, userstore.UpdateUserPasswordHashParams{
		ID:           row.UserID,
		PasswordHash: pgText(string(hash)),
	}); err != nil {
		return err
	}
	_ = s.auth.RevokeAllRefreshTokensForUser(ctx, row.UserID)
	_ = s.auth.RevokeAllUserSessions(ctx, row.UserID)
	uid, _ := pgxutil.ToUUID(row.UserID)
	s.audit(ctx, uid, "auth.reset_password", "user", &uid, meta.IP, meta.UserAgent)
	return nil
}

func (s *Service) VerifyEmail(ctx context.Context, token string, meta clientMeta) error {
	h := s.hashOpaque(strings.TrimSpace(token))
	row, err := s.auth.ConsumeEmailVerificationToken(ctx, h)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("invalid_token")
		}
		return err
	}
	if err := s.users.SetUserEmailVerified(ctx, row.UserID); err != nil {
		return err
	}
	uid, _ := pgxutil.ToUUID(row.UserID)
	s.audit(ctx, uid, "auth.verify_email", "user", &uid, meta.IP, meta.UserAgent)
	return nil
}

func (s *Service) ResendEmailVerification(ctx context.Context, userID uuid.UUID, meta clientMeta) error {
	u, err := s.users.GetUserByID(ctx, pgxutil.UUID(userID))
	if err != nil {
		return err
	}
	if !u.Email.Valid {
		return errors.New("no_email")
	}
	if u.EmailVerified {
		return errors.New("already_verified")
	}
	raw, err := s.createEmailVerifyToken(ctx, userID, u.Email.String)
	if err != nil {
		return err
	}
	if s.emailSender != nil {
		if err := s.emailSender.SendEmailVerification(ctx, u.Email.String, raw); err != nil {
			return err
		}
	} else {
		s.logSecret("email_verification_token_resend", zap.String("email", u.Email.String), zap.String("token", raw))
	}
	s.audit(ctx, userID, "auth.resend_email_verification", "user", &userID, meta.IP, meta.UserAgent)
	return nil
}

const phonePurposeVerify = "verify_phone"

func (s *Service) phoneOTPRateKey(phone string) string {
	return "otp:hour:" + phone
}

func (s *Service) SendPhoneOTP(ctx context.Context, userID uuid.UUID, phone string, meta clientMeta) error {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return errors.New("phone_required")
	}
	if s.redis != nil {
		key := s.phoneOTPRateKey(phone)
		n, err := s.redis.Incr(ctx, key).Result()
		if err != nil {
			return err
		}
		if n == 1 {
			_ = s.redis.Expire(ctx, key, time.Hour)
		}
		max := s.cfg.Auth.OTPMaxPerHour
		if max <= 0 {
			max = 5
		}
		if int(n) > max {
			return errors.New("rate_limited")
		}
	}
	code, err := randomDigits(6)
	if err != nil {
		return err
	}
	otpHash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.MinCost)
	if err != nil {
		return err
	}
	ttl := s.cfg.Auth.OTPCodeTTL
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	exp := pgtype.Timestamptz{Time: time.Now().Add(ttl), Valid: true}
	if _, err := s.auth.CreatePhoneOTPAttempt(ctx, authstore.CreatePhoneOTPAttemptParams{
		UserID:    pgxutil.UUID(userID),
		Phone:     phone,
		OtpHash:   string(otpHash),
		Purpose:   phonePurposeVerify,
		ExpiresAt: exp,
	}); err != nil {
		return err
	}
	if s.otpSender != nil {
		if err := s.otpSender.SendOTP(ctx, phone, code); err != nil {
			return err
		}
	} else {
		s.logSecret("phone_otp", zap.String("phone", phone), zap.String("code", code))
	}
	s.audit(ctx, userID, "auth.send_phone_otp", "user", &userID, meta.IP, meta.UserAgent)
	return nil
}

func randomDigits(n int) (string, error) {
	const digits = "0123456789"
	b := make([]byte, n)
	for i := range b {
		v, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		b[i] = digits[v.Int64()]
	}
	return string(b), nil
}

func (s *Service) VerifyPhoneOTP(ctx context.Context, userID uuid.UUID, phone, code string, meta clientMeta) error {
	phone = strings.TrimSpace(phone)
	code = strings.TrimSpace(code)
	row, err := s.auth.GetLatestPendingPhoneOTPForUser(ctx, authstore.GetLatestPendingPhoneOTPForUserParams{
		UserID:  pgxutil.UUID(userID),
		Purpose: phonePurposeVerify,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("no_otp")
		}
		return err
	}
	if row.Phone != phone {
		return errors.New("phone_mismatch")
	}
	if bcrypt.CompareHashAndPassword([]byte(row.OtpHash), []byte(code)) != nil {
		_ = s.auth.IncrementPhoneOTPAttempts(ctx, row.ID)
		if s.recordOTPVerifyFailure(ctx, userID, phone) {
			return errors.New("otp_locked")
		}
		return errors.New("invalid_code")
	}
	s.clearOTPVerifyFailures(ctx, userID, phone)
	if _, err := s.auth.VerifyPhoneOTPAttempt(ctx, row.ID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("invalid_code")
		}
		return err
	}
	if err := s.users.SetUserPhoneVerified(ctx, pgxutil.UUID(userID)); err != nil {
		return err
	}
	s.audit(ctx, userID, "auth.verify_phone_otp", "user", &userID, meta.IP, meta.UserAgent)
	return nil
}

func (s *Service) loginFailKey(email string) string {
	return "auth:login:fail:" + email
}

func (s *Service) loginLockKey(email string) string {
	return "auth:login:lock:" + email
}

func (s *Service) recordLoginFailure(ctx context.Context, email string) {
	if s.redis == nil || email == "" {
		return
	}
	failures, err := s.redis.Incr(ctx, s.loginFailKey(email)).Result()
	if err != nil {
		return
	}
	_ = s.redis.Expire(ctx, s.loginFailKey(email), time.Hour)
	max := s.cfg.Auth.MaxLoginFailures
	if max <= 0 {
		max = 8
	}
	if int(failures) >= max {
		d := s.cfg.Auth.LoginLockoutDuration
		if d <= 0 {
			d = 20 * time.Minute
		}
		_ = s.redis.Set(ctx, s.loginLockKey(email), "1", d).Err()
	}
}

func (s *Service) clearLoginFailures(ctx context.Context, email string) {
	if s.redis == nil || email == "" {
		return
	}
	_ = s.redis.Del(ctx, s.loginFailKey(email), s.loginLockKey(email)).Err()
}

func (s *Service) isLoginLocked(ctx context.Context, email string) bool {
	if s.redis == nil || email == "" {
		return false
	}
	locked, err := s.redis.Exists(ctx, s.loginLockKey(email)).Result()
	return err == nil && locked > 0
}

func (s *Service) otpVerifyFailKey(userID uuid.UUID, phone string) string {
	return "auth:otp:verify:fail:" + userID.String() + ":" + phone
}

func (s *Service) recordOTPVerifyFailure(ctx context.Context, userID uuid.UUID, phone string) bool {
	if s.redis == nil {
		return false
	}
	key := s.otpVerifyFailKey(userID, phone)
	n, err := s.redis.Incr(ctx, key).Result()
	if err != nil {
		return false
	}
	_ = s.redis.Expire(ctx, key, time.Hour)
	max := s.cfg.Auth.OTPMaxVerifyFailures
	if max <= 0 {
		max = 5
	}
	return int(n) >= max
}

func (s *Service) clearOTPVerifyFailures(ctx context.Context, userID uuid.UUID, phone string) {
	if s.redis == nil {
		return
	}
	_ = s.redis.Del(ctx, s.otpVerifyFailKey(userID, phone)).Err()
}

// OAuthGoogleCallback completes Google OAuth (requires exchanged access token + profile).
func (s *Service) OAuthUpsertUser(ctx context.Context, provider, providerUserID, email string, profile map[string]any, meta clientMeta) (*TokenPair, uuid.UUID, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return nil, uuid.Nil, errors.New("oauth_email_required")
	}
	if verified, ok := profile["email_verified"].(bool); ok && !verified {
		return nil, uuid.Nil, errors.New("oauth_email_unverified")
	}
	rawProfile, _ := json.Marshal(profile)
	// Find existing oauth identity
	// Simplified: lookup by email for merge; else create user
	var uid uuid.UUID
	u, err := s.users.GetUserByEmail(ctx, pgText(email))
	if err == nil {
		id, ok := pgxutil.ToUUID(u.ID)
		if !ok {
			return nil, uuid.Nil, errors.New("user_id")
		}
		uid = id
	} else if errors.Is(err, pgx.ErrNoRows) {
		tx, err := s.pool.Begin(ctx)
		if err != nil {
			return nil, uuid.Nil, err
		}
		defer func() { _ = tx.Rollback(ctx) }()
		uq := userstore.New(tx)
		rq := rbacstore.New(tx)
		au, err := uq.CreateUser(ctx, userstore.CreateUserParams{
			Email:        pgText(email),
			Phone:        pgtype.Text{},
			PasswordHash: pgtype.Text{},
			Column4:      nil,
		})
		if err != nil {
			return nil, uuid.Nil, err
		}
		id, ok := pgxutil.ToUUID(au.ID)
		if !ok {
			return nil, uuid.Nil, errors.New("user_id")
		}
		uid = id
		disp := email
		if i := strings.IndexByte(email, '@'); i > 0 {
			disp = email[:i]
		}
		if _, err := uq.UpdateUserProfile(ctx, userstore.UpdateUserProfileParams{
			UserID:      au.ID,
			FirstName:   pgtype.Text{},
			LastName:    pgtype.Text{},
			DisplayName: pgText(disp),
			AvatarUrl:   pgtype.Text{},
			Gender:      pgtype.Text{},
			Birthdate:   pgtype.Date{},
			Column8:     nil,
			CountryCode: pgtype.Text{},
			Timezone:    pgtype.Text{},
		}); err != nil {
			return nil, uuid.Nil, err
		}
		if err := rq.AssignUserRoleByRoleName(ctx, rbacstore.AssignUserRoleByRoleNameParams{
			UserID:   au.ID,
			RoleName: "buyer",
		}); err != nil {
			return nil, uuid.Nil, err
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, uuid.Nil, err
		}
	} else {
		return nil, uuid.Nil, err
	}
	if _, err := s.auth.UpsertOAuthIdentity(ctx, authstore.UpsertOAuthIdentityParams{
		UserID:         pgxutil.UUID(uid),
		Provider:       provider,
		ProviderUserID: providerUserID,
		Email:          pgText(email),
		RawProfile:     rawProfile,
	}); err != nil {
		return nil, uuid.Nil, err
	}
	roles, err := s.roleNames(ctx, uid)
	if err != nil {
		return nil, uuid.Nil, err
	}
	pair, err := s.issueTokenPair(ctx, uid, roles)
	if err != nil {
		return nil, uuid.Nil, err
	}
	if _, err := s.createSessionAndRefresh(ctx, uid, meta, pair.RefreshToken); err != nil {
		return nil, uuid.Nil, err
	}
	s.audit(ctx, uid, "auth.oauth_login", "user", &uid, meta.IP, meta.UserAgent)
	return pair, uid, nil
}
