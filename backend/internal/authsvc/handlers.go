package authsvc

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	authstore "github.com/bitik/backend/internal/store/auth"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/middleware"
	"github.com/bitik/backend/internal/pgxutil"
	"github.com/bitik/backend/internal/store/users"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"
)

func (s *Service) setRefreshCookie(c *gin.Context, refreshToken string) {
	ttl := s.cfg.Auth.RefreshTokenTTL
	maxAge := int(ttl.Seconds())
	if maxAge < 0 {
		maxAge = 0
	}
	sameSite := http.SameSiteLaxMode
	switch strings.ToLower(strings.TrimSpace(s.cfg.Auth.RefreshCookieSameSite)) {
	case "strict":
		sameSite = http.SameSiteStrictMode
	case "none":
		sameSite = http.SameSiteNoneMode
	}
	if strings.EqualFold(s.cfg.App.Environment, "production") {
		sameSite = http.SameSiteStrictMode
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     s.cfg.Auth.RefreshCookieName,
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.cfg.Auth.RefreshCookieSecure,
		SameSite: sameSite,
		MaxAge:   maxAge,
	})
}

func (s *Service) clearRefreshCookie(c *gin.Context) {
	sameSite := http.SameSiteLaxMode
	if strings.EqualFold(s.cfg.App.Environment, "production") {
		sameSite = http.SameSiteStrictMode
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     s.cfg.Auth.RefreshCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   s.cfg.Auth.RefreshCookieSecure,
		SameSite: sameSite,
		MaxAge:   -1,
	})
}

func (s *Service) refreshTokenFromCookie(c *gin.Context) string {
	cookie, err := c.Request.Cookie(s.cfg.Auth.RefreshCookieName)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(cookie.Value)
}

func (s *Service) RegisterRoutes(v1 *gin.RouterGroup) {
	jwt := middleware.RequireBearerJWT(s.cfg)
	active := s.RequireActiveUser()

	auth := v1.Group("/auth")
	{
		auth.POST("/register", s.HandleRegister)
		auth.POST("/login", s.HandleLogin)
		auth.POST("/logout", s.HandleLogout)
		auth.POST("/refresh-token", s.HandleRefresh)
		auth.POST("/forgot-password", s.HandleForgotPassword)
		auth.POST("/reset-password", s.HandleResetPassword)
		auth.POST("/verify-email", s.HandleVerifyEmail)
		auth.POST("/resend-email-verification", jwt, active, s.HandleResendEmailVerification)
		auth.POST("/send-phone-otp", jwt, active, s.HandleSendPhoneOTP)
		auth.POST("/verify-phone-otp", jwt, active, s.HandleVerifyPhoneOTP)

		auth.GET("/oauth/google", s.HandleGoogleOAuthStart)
		auth.GET("/oauth/google/callback", s.HandleGoogleOAuthCallback)
		auth.GET("/oauth/facebook", s.HandleFacebookOAuthStart)
		auth.GET("/oauth/facebook/callback", s.HandleFacebookOAuthCallback)
		auth.GET("/oauth/apple", s.HandleAppleOAuthStart)
		auth.GET("/oauth/apple/callback", s.HandleAppleOAuthCallback)
	}

	users := v1.Group("/users", jwt, active)
	{
		users.GET("/me", s.HandleGetMe)
		users.PATCH("/me", s.HandlePatchMe)
		users.DELETE("/me", s.HandleDeleteMe)
		users.GET("/me/profile", s.HandleGetProfile)
		users.PATCH("/me/profile", s.HandlePatchProfile)
		users.GET("/me/sessions", s.HandleListSessions)
		users.DELETE("/me/sessions/:session_id", s.HandleRevokeSession)
		users.GET("/me/devices", s.HandleListDevices)
		users.DELETE("/me/devices/:device_id", s.HandleRevokeDevice)
	}

	if s.Casbin != nil {
		admin := v1.Group("/admin", jwt, active, middleware.RequireCasbinHTTP(s.Casbin))
		admin.GET("/dashboard", s.HandleAdminDashboard)
		seller := v1.Group("/seller", jwt, active, middleware.RequireCasbinHTTP(s.Casbin))
		seller.GET("/health", s.HandleSellerHealth)
	}
}

func (s *Service) RequireActiveUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := currentUserUUID(c)
		if !ok {
			apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing user context.")
			c.Abort()
			return
		}
		u, err := s.users.GetUserByID(c.Request.Context(), pgxutil.UUID(uid))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "User no longer exists.")
				c.Abort()
				return
			}
			apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load user.")
			c.Abort()
			return
		}
		if err := s.userActive(u); err != nil {
			apiresponse.Error(c, http.StatusForbidden, "account_locked", "This account cannot access the API.")
			c.Abort()
			return
		}
		roles, err := s.roleNames(c.Request.Context(), uid)
		if err != nil {
			apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load roles.")
			c.Abort()
			return
		}
		c.Set(middleware.AuthRolesKey, roles)
		c.Next()
	}
}

func (s *Service) HandleAdminDashboard(c *gin.Context) {
	apiresponse.OK(c, gin.H{"module": "admin", "status": "ok"})
}

func (s *Service) HandleSellerHealth(c *gin.Context) {
	apiresponse.OK(c, gin.H{"module": "seller", "status": "ok"})
}

func (s *Service) HandleRegister(c *gin.Context) {
	var req struct {
		Email       string `json:"email" binding:"required,email"`
		Password    string `json:"password" binding:"required,min=8"`
		DisplayName string `json:"display_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_body", "Invalid JSON body.")
		return
	}
	var dn *string
	if strings.TrimSpace(req.DisplayName) != "" {
		v := strings.TrimSpace(req.DisplayName)
		dn = &v
	}
	pair, _, err := s.Register(c.Request.Context(), req.Email, req.Password, dn, clientMetaFromGin(c))
	if err != nil {
		s.writeAuthError(c, err)
		return
	}
	s.setRefreshCookie(c, pair.RefreshToken)
	apiresponse.OK(c, pair)
}

func (s *Service) HandleLogin(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_body", "Invalid JSON body.")
		return
	}
	pair, _, err := s.Login(c.Request.Context(), req.Email, req.Password, clientMetaFromGin(c))
	if err != nil {
		s.writeAuthError(c, err)
		return
	}
	s.setRefreshCookie(c, pair.RefreshToken)
	apiresponse.OK(c, pair)
}

func (s *Service) HandleLogout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = c.ShouldBindJSON(&req)
	if strings.TrimSpace(req.RefreshToken) == "" {
		req.RefreshToken = s.refreshTokenFromCookie(c)
	}
	if strings.TrimSpace(req.RefreshToken) == "" {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_body", "refresh_token is required.")
		return
	}
	if err := s.Logout(c.Request.Context(), req.RefreshToken, clientMetaFromGin(c)); err != nil {
		s.writeAuthError(c, err)
		return
	}
	s.clearRefreshCookie(c)
	apiresponse.OK(c, gin.H{"logged_out": true})
}

func (s *Service) HandleRefresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = c.ShouldBindJSON(&req)
	if strings.TrimSpace(req.RefreshToken) == "" {
		req.RefreshToken = s.refreshTokenFromCookie(c)
	}
	if strings.TrimSpace(req.RefreshToken) == "" {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_body", "refresh_token is required.")
		return
	}
	pair, err := s.Refresh(c.Request.Context(), req.RefreshToken, clientMetaFromGin(c))
	if err != nil {
		s.writeAuthError(c, err)
		return
	}
	s.setRefreshCookie(c, pair.RefreshToken)
	apiresponse.OK(c, pair)
}

func (s *Service) HandleForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_body", "Invalid JSON body.")
		return
	}
	_ = s.ForgotPassword(c.Request.Context(), req.Email, clientMetaFromGin(c))
	apiresponse.OK(c, gin.H{"ok": true})
}

func (s *Service) HandleResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_body", "Invalid JSON body.")
		return
	}
	if err := s.ResetPassword(c.Request.Context(), req.Token, req.NewPassword, clientMetaFromGin(c)); err != nil {
		s.writeAuthError(c, err)
		return
	}
	apiresponse.OK(c, gin.H{"ok": true})
}

func (s *Service) HandleVerifyEmail(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_body", "token is required.")
		return
	}
	if err := s.VerifyEmail(c.Request.Context(), req.Token, clientMetaFromGin(c)); err != nil {
		s.writeAuthError(c, err)
		return
	}
	apiresponse.OK(c, gin.H{"verified": true})
}

func (s *Service) HandleResendEmailVerification(c *gin.Context) {
	uid, ok := currentUserUUID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing user context.")
		return
	}
	if err := s.ResendEmailVerification(c.Request.Context(), uid, clientMetaFromGin(c)); err != nil {
		s.writeAuthError(c, err)
		return
	}
	apiresponse.OK(c, gin.H{"ok": true})
}

func (s *Service) HandleSendPhoneOTP(c *gin.Context) {
	uid, ok := currentUserUUID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing user context.")
		return
	}
	var req struct {
		Phone string `json:"phone" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_body", "phone is required.")
		return
	}
	if err := s.SendPhoneOTP(c.Request.Context(), uid, req.Phone, clientMetaFromGin(c)); err != nil {
		s.writeAuthError(c, err)
		return
	}
	apiresponse.OK(c, gin.H{"ok": true})
}

func (s *Service) HandleVerifyPhoneOTP(c *gin.Context) {
	uid, ok := currentUserUUID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing user context.")
		return
	}
	var req struct {
		Phone string `json:"phone" binding:"required"`
		Code  string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_body", "phone and code are required.")
		return
	}
	if err := s.VerifyPhoneOTP(c.Request.Context(), uid, req.Phone, req.Code, clientMetaFromGin(c)); err != nil {
		s.writeAuthError(c, err)
		return
	}
	apiresponse.OK(c, gin.H{"verified": true})
}

func currentUserUUID(c *gin.Context) (uuid.UUID, bool) {
	raw, ok := c.Get(middleware.AuthUserIDKey)
	if !ok {
		return uuid.Nil, false
	}
	s, ok := raw.(string)
	if !ok {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

func (s *Service) writeAuthError(c *gin.Context, err error) {
	switch err.Error() {
	case "email_taken":
		apiresponse.Error(c, http.StatusConflict, "email_taken", "That email is already registered.")
	case "weak_password":
		apiresponse.Error(c, http.StatusUnprocessableEntity, "weak_password", "Password must be at least 8 characters.")
	case "email_required":
		apiresponse.Error(c, http.StatusBadRequest, "validation_error", "Email is required.")
	case "invalid_credentials":
		apiresponse.Error(c, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password.")
	case "banned", "inactive", "deleted":
		apiresponse.Error(c, http.StatusForbidden, "account_locked", "This account cannot sign in.")
	case "missing_refresh", "invalid_refresh", "expired_refresh", "refresh_reuse":
		apiresponse.Error(c, http.StatusUnauthorized, "invalid_refresh", "Refresh token is invalid or expired.")
	case "invalid_token":
		apiresponse.Error(c, http.StatusBadRequest, "invalid_token", "The token is invalid or has expired.")
	case "no_email":
		apiresponse.Error(c, http.StatusBadRequest, "no_email", "No email on file.")
	case "already_verified":
		apiresponse.Error(c, http.StatusConflict, "already_verified", "Email is already verified.")
	case "rate_limited":
		apiresponse.Error(c, http.StatusTooManyRequests, "rate_limited", "Too many OTP requests. Try again later.")
	case "login_locked":
		apiresponse.Error(c, http.StatusTooManyRequests, "login_locked", "Too many failed sign-in attempts. Please try again later.")
	case "otp_locked":
		apiresponse.Error(c, http.StatusTooManyRequests, "otp_locked", "Too many invalid OTP attempts. Please request a new OTP later.")
	case "phone_required", "phone_mismatch", "no_otp", "invalid_code":
		apiresponse.Error(c, http.StatusBadRequest, err.Error(), "Unable to verify phone OTP.")
	case "oauth_email_required":
		apiresponse.Error(c, http.StatusBadRequest, "oauth_email_required", "OAuth provider did not return an email.")
	case "oauth_email_unverified":
		apiresponse.Error(c, http.StatusBadRequest, "oauth_email_unverified", "OAuth provider did not verify this email.")
	default:
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			apiresponse.Error(c, http.StatusConflict, "conflict", "Email or phone is already in use.")
			return
		}
		if s.log != nil {
			s.log.Debug("auth_error", zap.Error(err))
		}
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Something went wrong.")
	}
}

func (s *Service) HandleGetMe(c *gin.Context) {
	uid, ok := currentUserUUID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing user context.")
		return
	}
	u, err := s.users.GetUserByID(c.Request.Context(), pgxutil.UUID(uid))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			apiresponse.Error(c, http.StatusNotFound, "not_found", "User not found.")
			return
		}
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load user.")
		return
	}
	apiresponse.OK(c, publicUser(u))
}

func publicUser(u userstore.User) gin.H {
	uid, ok := pgxutil.ToUUID(u.ID)
	if !ok {
		return gin.H{"error": "invalid_user_id"}
	}
	out := gin.H{
		"id":             uid.String(),
		"status":         u.Status,
		"email_verified": u.EmailVerified,
		"phone_verified": u.PhoneVerified,
		"created_at":     u.CreatedAt.Time,
	}
	if u.Email.Valid {
		out["email"] = u.Email.String
	}
	if u.Phone.Valid {
		out["phone"] = u.Phone.String
	}
	return out
}

func (s *Service) HandlePatchMe(c *gin.Context) {
	uid, ok := currentUserUUID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing user context.")
		return
	}
	var req struct {
		Email *string `json:"email"`
		Phone *string `json:"phone"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_body", "Invalid JSON body.")
		return
	}
	email := pgtype.Text{}
	if req.Email != nil {
		email = pgText(strings.TrimSpace(strings.ToLower(*req.Email)))
	}
	phone := pgtype.Text{}
	if req.Phone != nil {
		phone = pgText(strings.TrimSpace(*req.Phone))
	}
	if !email.Valid && !phone.Valid {
		apiresponse.Error(c, http.StatusBadRequest, "validation_error", "email or phone is required.")
		return
	}
	u, err := s.users.UpdateUserContact(c.Request.Context(), userstore.UpdateUserContactParams{
		ID:    pgxutil.UUID(uid),
		Email: email,
		Phone: phone,
	})
	if err != nil {
		s.writeAuthError(c, err)
		return
	}
	s.audit(c.Request.Context(), uid, "user.update_contact", "user", &uid, clientMetaFromGin(c).IP, clientMetaFromGin(c).UserAgent)
	apiresponse.OK(c, publicUser(u))
}

func (s *Service) HandleDeleteMe(c *gin.Context) {
	uid, ok := currentUserUUID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing user context.")
		return
	}
	if err := s.users.SoftDeleteUser(c.Request.Context(), pgxutil.UUID(uid)); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not delete account.")
		return
	}
	_ = s.auth.RevokeAllRefreshTokensForUser(c.Request.Context(), pgxutil.UUID(uid))
	_ = s.auth.RevokeAllUserSessions(c.Request.Context(), pgxutil.UUID(uid))
	s.audit(c.Request.Context(), uid, "user.delete_self", "user", &uid, clientMetaFromGin(c).IP, clientMetaFromGin(c).UserAgent)
	apiresponse.OK(c, gin.H{"deleted": true})
}

func (s *Service) HandleGetProfile(c *gin.Context) {
	uid, ok := currentUserUUID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing user context.")
		return
	}
	p, err := s.users.GetUserProfileByUserID(c.Request.Context(), pgxutil.UUID(uid))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			apiresponse.Error(c, http.StatusNotFound, "not_found", "Profile not found.")
			return
		}
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load profile.")
		return
	}
	apiresponse.OK(c, profileJSON(p))
}

func profileJSON(p userstore.UserProfile) gin.H {
	h := gin.H{
		"language": p.Language,
	}
	if uid, ok := pgxutil.ToUUID(p.UserID); ok {
		h["user_id"] = uid.String()
	}
	if p.FirstName.Valid {
		h["first_name"] = p.FirstName.String
	}
	if p.LastName.Valid {
		h["last_name"] = p.LastName.String
	}
	if p.DisplayName.Valid {
		h["display_name"] = p.DisplayName.String
	}
	if p.AvatarUrl.Valid {
		h["avatar_url"] = p.AvatarUrl.String
	}
	if p.CountryCode.Valid {
		h["country_code"] = p.CountryCode.String
	}
	if p.Timezone.Valid {
		h["timezone"] = p.Timezone.String
	}
	return h
}

func (s *Service) HandlePatchProfile(c *gin.Context) {
	uid, ok := currentUserUUID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing user context.")
		return
	}
	var req struct {
		FirstName   *string `json:"first_name"`
		LastName    *string `json:"last_name"`
		DisplayName *string `json:"display_name"`
		AvatarURL   *string `json:"avatar_url"`
		Language    *string `json:"language"`
		CountryCode *string `json:"country_code"`
		Timezone    *string `json:"timezone"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_body", "Invalid JSON body.")
		return
	}
	ctx := c.Request.Context()
	cur, err := s.users.GetUserProfileByUserID(ctx, pgxutil.UUID(uid))
	if errors.Is(err, pgx.ErrNoRows) {
		cur = userstore.UserProfile{UserID: pgxutil.UUID(uid), Language: "en"}
	} else if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load profile.")
		return
	}
	p := mergeProfile(cur, req)
	if _, err := s.users.UpdateUserProfile(ctx, p); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not update profile.")
		return
	}
	apiresponse.OK(c, gin.H{"updated": true})
}

func mergeProfile(cur userstore.UserProfile, req struct {
	FirstName   *string `json:"first_name"`
	LastName    *string `json:"last_name"`
	DisplayName *string `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
	Language    *string `json:"language"`
	CountryCode *string `json:"country_code"`
	Timezone    *string `json:"timezone"`
}) userstore.UpdateUserProfileParams {
	pickText := func(ptr *string, existing pgtype.Text) pgtype.Text {
		if ptr != nil {
			return pgText(strings.TrimSpace(*ptr))
		}
		return existing
	}
	lang := cur.Language
	if req.Language != nil && strings.TrimSpace(*req.Language) != "" {
		lang = strings.TrimSpace(*req.Language)
	}
	return userstore.UpdateUserProfileParams{
		UserID:      cur.UserID,
		FirstName:   pickText(req.FirstName, cur.FirstName),
		LastName:    pickText(req.LastName, cur.LastName),
		DisplayName: pickText(req.DisplayName, cur.DisplayName),
		AvatarUrl:   pickText(req.AvatarURL, cur.AvatarUrl),
		Gender:      cur.Gender,
		Birthdate:   cur.Birthdate,
		Column8:     lang,
		CountryCode: pickText(req.CountryCode, cur.CountryCode),
		Timezone:    pickText(req.Timezone, cur.Timezone),
	}
}

func (s *Service) HandleListSessions(c *gin.Context) {
	uid, ok := currentUserUUID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing user context.")
		return
	}
	list, err := s.auth.ListUserSessionsForUser(c.Request.Context(), pgxutil.UUID(uid))
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list sessions.")
		return
	}
	out := make([]gin.H, 0, len(list))
	for _, x := range list {
		h := gin.H{
			"last_seen_at": x.LastSeenAt.Time,
			"created_at":   x.CreatedAt.Time,
			"revoked":      x.RevokedAt.Valid,
		}
		if sid, ok := pgxutil.ToUUID(x.ID); ok {
			h["id"] = sid.String()
		}
		if x.DeviceID.Valid {
			h["device_id"] = x.DeviceID.String
		}
		if x.Platform.Valid {
			h["platform"] = x.Platform.String
		}
		out = append(out, h)
	}
	apiresponse.OK(c, gin.H{"sessions": out})
}

func (s *Service) HandleRevokeSession(c *gin.Context) {
	uid, ok := currentUserUUID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing user context.")
		return
	}
	sid, err := uuid.Parse(c.Param("session_id"))
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_session", "Invalid session id.")
		return
	}
	if err := s.auth.RevokeUserSession(c.Request.Context(), authstore.RevokeUserSessionParams{
		ID:     pgxutil.UUID(sid),
		UserID: pgxutil.UUID(uid),
	}); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not revoke session.")
		return
	}
	if err := s.auth.RevokeRefreshTokensForSession(c.Request.Context(), authstore.RevokeRefreshTokensForSessionParams{
		SessionID: pgxutil.UUID(sid),
		UserID:    pgxutil.UUID(uid),
	}); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not revoke session tokens.")
		return
	}
	apiresponse.OK(c, gin.H{"revoked": true})
}

func (s *Service) HandleListDevices(c *gin.Context) {
	uid, ok := currentUserUUID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing user context.")
		return
	}
	list, err := s.users.ListUserDevicesForUser(c.Request.Context(), pgxutil.UUID(uid))
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list devices.")
		return
	}
	out := make([]gin.H, 0, len(list))
	for _, d := range list {
		h := gin.H{
			"device_id":    d.DeviceID,
			"platform":     d.Platform,
			"last_seen_at": d.LastSeenAt.Time,
			"revoked":      d.RevokedAt.Valid,
		}
		if did, ok := pgxutil.ToUUID(d.ID); ok {
			h["id"] = did.String()
		}
		out = append(out, h)
	}
	apiresponse.OK(c, gin.H{"devices": out})
}

func (s *Service) HandleRevokeDevice(c *gin.Context) {
	uid, ok := currentUserUUID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing user context.")
		return
	}
	did := strings.TrimSpace(c.Param("device_id"))
	if did == "" {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_device", "device_id is required.")
		return
	}
	if err := s.users.RevokeUserDevice(c.Request.Context(), userstore.RevokeUserDeviceParams{
		UserID:   pgxutil.UUID(uid),
		DeviceID: did,
	}); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not revoke device.")
		return
	}
	apiresponse.OK(c, gin.H{"revoked": true})
}

// --- OAuth (Google / Facebook) ---

func (s *Service) oauthBase() string {
	return strings.TrimRight(strings.TrimSpace(s.cfg.Auth.PublicBaseURL), "/")
}

func (s *Service) oauthRedirectBase() string {
	return strings.TrimRight(strings.TrimSpace(s.cfg.Auth.OAuthRedirectBaseURL), "/")
}

func (s *Service) requireRedisOAuth(c *gin.Context) bool {
	if s.redis == nil {
		apiresponse.Error(c, http.StatusServiceUnavailable, "oauth_unavailable", "OAuth requires Redis for CSRF state storage.")
		return false
	}
	return true
}

func (s *Service) HandleGoogleOAuthStart(c *gin.Context) {
	if strings.TrimSpace(s.cfg.Auth.OAuthGoogleClientID) == "" {
		apiresponse.Error(c, http.StatusNotImplemented, "oauth_not_configured", "Google OAuth is not configured.")
		return
	}
	if !s.requireRedisOAuth(c) {
		return
	}
	state, err := newOpaqueToken()
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not start OAuth.")
		return
	}
	ctx := c.Request.Context()
	if err := s.redis.Set(ctx, "oauth:state:"+state, "google", 10*time.Minute).Err(); err != nil {
		apiresponse.Error(c, http.StatusServiceUnavailable, "oauth_store_error", "Could not persist OAuth state.")
		return
	}
	conf := &oauth2.Config{
		ClientID:     s.cfg.Auth.OAuthGoogleClientID,
		ClientSecret: s.cfg.Auth.OAuthGoogleClientSecret,
		RedirectURL:  s.oauthRedirectBase() + "/auth/callback/google",
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
	url := conf.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (s *Service) HandleGoogleOAuthCallback(c *gin.Context) {
	if errMsg := c.Query("error"); errMsg != "" {
		apiresponse.Error(c, http.StatusBadRequest, "oauth_denied", c.Query("error_description"))
		return
	}
	state := c.Query("state")
	code := c.Query("code")
	if state == "" || code == "" {
		apiresponse.Error(c, http.StatusBadRequest, "oauth_invalid", "Missing state or code.")
		return
	}
	ctx := c.Request.Context()
	if s.redis != nil {
		v, err := s.redis.Get(ctx, "oauth:state:"+state).Result()
		if err != nil || v != "google" {
			apiresponse.Error(c, http.StatusBadRequest, "oauth_state_invalid", "Invalid or expired OAuth state.")
			return
		}
		_ = s.redis.Del(ctx, "oauth:state:"+state)
	}
	conf := &oauth2.Config{
		ClientID:     s.cfg.Auth.OAuthGoogleClientID,
		ClientSecret: s.cfg.Auth.OAuthGoogleClientSecret,
		RedirectURL:  s.oauthRedirectBase() + "/auth/callback/google",
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "oauth_exchange_failed", "Could not exchange authorization code.")
		return
	}
	client := conf.Client(ctx, tok)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		apiresponse.Error(c, http.StatusBadGateway, "oauth_profile_failed", "Could not load Google profile.")
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		apiresponse.Error(c, http.StatusBadGateway, "oauth_profile_failed", "Google userinfo error.")
		return
	}
	var gu struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		Name          string `json:"name"`
		VerifiedEmail bool   `json:"verified_email"`
	}
	if err := json.Unmarshal(body, &gu); err != nil || gu.Email == "" {
		apiresponse.Error(c, http.StatusBadGateway, "oauth_profile_invalid", "Google profile did not include an email.")
		return
	}
	if !gu.VerifiedEmail {
		s.writeAuthError(c, errors.New("oauth_email_unverified"))
		return
	}
	prof := map[string]any{"name": gu.Name, "email_verified": gu.VerifiedEmail}
	pair, _, err := s.OAuthUpsertUser(ctx, "google", gu.ID, gu.Email, prof, clientMetaFromGin(c))
	if err != nil {
		s.writeAuthError(c, err)
		return
	}
	s.setRefreshCookie(c, pair.RefreshToken)
	apiresponse.OK(c, pair)
}

func (s *Service) HandleFacebookOAuthStart(c *gin.Context) {
	if strings.TrimSpace(s.cfg.Auth.OAuthFacebookAppID) == "" {
		apiresponse.Error(c, http.StatusNotImplemented, "oauth_not_configured", "Facebook OAuth is not configured.")
		return
	}
	if !s.requireRedisOAuth(c) {
		return
	}
	state, err := newOpaqueToken()
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not start OAuth.")
		return
	}
	ctx := c.Request.Context()
	if err := s.redis.Set(ctx, "oauth:state:"+state, "facebook", 10*time.Minute).Err(); err != nil {
		apiresponse.Error(c, http.StatusServiceUnavailable, "oauth_store_error", "Could not persist OAuth state.")
		return
	}
	conf := &oauth2.Config{
		ClientID:     s.cfg.Auth.OAuthFacebookAppID,
		ClientSecret: s.cfg.Auth.OAuthFacebookAppSecret,
		RedirectURL:  s.oauthRedirectBase() + "/auth/callback/facebook",
		Scopes:       []string{"email", "public_profile"},
		Endpoint:     facebook.Endpoint,
	}
	c.Redirect(http.StatusTemporaryRedirect, conf.AuthCodeURL(state, oauth2.AccessTypeOffline))
}

func (s *Service) HandleFacebookOAuthCallback(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")
	if state == "" || code == "" {
		apiresponse.Error(c, http.StatusBadRequest, "oauth_invalid", "Missing state or code.")
		return
	}
	ctx := c.Request.Context()
	if s.redis != nil {
		v, err := s.redis.Get(ctx, "oauth:state:"+state).Result()
		if err != nil || v != "facebook" {
			apiresponse.Error(c, http.StatusBadRequest, "oauth_state_invalid", "Invalid or expired OAuth state.")
			return
		}
		_ = s.redis.Del(ctx, "oauth:state:"+state)
	}
	conf := &oauth2.Config{
		ClientID:     s.cfg.Auth.OAuthFacebookAppID,
		ClientSecret: s.cfg.Auth.OAuthFacebookAppSecret,
		RedirectURL:  s.oauthRedirectBase() + "/auth/callback/facebook",
		Scopes:       []string{"email", "public_profile"},
		Endpoint:     facebook.Endpoint,
	}
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "oauth_exchange_failed", "Could not exchange authorization code.")
		return
	}
	u := url.Values{}
	u.Set("fields", "id,email,name,verified,is_verified")
	u.Set("access_token", tok.AccessToken)
	graphURL := "https://graph.facebook.com/me?" + u.Encode()
	resp, err := http.Get(graphURL)
	if err != nil {
		apiresponse.Error(c, http.StatusBadGateway, "oauth_profile_failed", "Could not load Facebook profile.")
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var fu struct {
		ID         string `json:"id"`
		Email      string `json:"email"`
		Name       string `json:"name"`
		Verified   bool   `json:"verified"`
		IsVerified bool   `json:"is_verified"`
	}
	if err := json.Unmarshal(body, &fu); err != nil || fu.Email == "" {
		apiresponse.Error(c, http.StatusBadGateway, "oauth_profile_invalid", "Facebook profile did not include an email.")
		return
	}
	emailVerified := fu.Verified || fu.IsVerified
	if !emailVerified {
		s.writeAuthError(c, errors.New("oauth_email_unverified"))
		return
	}
	prof := map[string]any{"name": fu.Name, "email_verified": emailVerified}
	pair, _, err := s.OAuthUpsertUser(ctx, "facebook", fu.ID, fu.Email, prof, clientMetaFromGin(c))
	if err != nil {
		s.writeAuthError(c, err)
		return
	}
	s.setRefreshCookie(c, pair.RefreshToken)
	apiresponse.OK(c, pair)
}

func (s *Service) HandleAppleOAuthStart(c *gin.Context) {
	conf, err := s.appleOAuthConfig()
	if err != nil {
		apiresponse.Error(c, http.StatusNotImplemented, "oauth_not_configured", "Apple Sign In is not configured (set auth.oauth_apple_client_id and related keys).")
		return
	}
	conf.RedirectURL = s.oauthRedirectBase() + "/auth/callback/apple"
	if !s.requireRedisOAuth(c) {
		return
	}
	state, err := newOpaqueToken()
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not start OAuth.")
		return
	}
	ctx := c.Request.Context()
	if err := s.redis.Set(ctx, "oauth:state:"+state, "apple", 10*time.Minute).Err(); err != nil {
		apiresponse.Error(c, http.StatusServiceUnavailable, "oauth_store_error", "Could not persist OAuth state.")
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, conf.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("response_mode", "query"),
	))
}

func (s *Service) HandleAppleOAuthCallback(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")
	if state == "" || code == "" {
		apiresponse.Error(c, http.StatusBadRequest, "oauth_invalid", "Missing state or code.")
		return
	}
	ctx := c.Request.Context()
	if s.redis != nil {
		v, err := s.redis.Get(ctx, "oauth:state:"+state).Result()
		if err != nil || v != "apple" {
			apiresponse.Error(c, http.StatusBadRequest, "oauth_state_invalid", "Invalid or expired OAuth state.")
			return
		}
		_ = s.redis.Del(ctx, "oauth:state:"+state)
	}
	conf, err := s.appleOAuthConfig()
	if err != nil {
		apiresponse.Error(c, http.StatusNotImplemented, "oauth_not_configured", "Apple Sign In is not configured.")
		return
	}
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "oauth_exchange_failed", "Could not exchange authorization code.")
		return
	}
	rawIDToken, _ := tok.Extra("id_token").(string)
	claims, err := parseAppleIDTokenUnverified(rawIDToken)
	if err != nil {
		apiresponse.Error(c, http.StatusBadGateway, "oauth_profile_invalid", "Apple id_token was invalid.")
		return
	}
	emailVerified := appleEmailVerified(claims["email_verified"])
	if !emailVerified {
		s.writeAuthError(c, errors.New("oauth_email_unverified"))
		return
	}
	sub, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)
	if sub == "" || email == "" {
		apiresponse.Error(c, http.StatusBadGateway, "oauth_profile_invalid", "Apple profile did not include required identity fields.")
		return
	}
	prof := map[string]any{"email_verified": emailVerified}
	pair, _, err := s.OAuthUpsertUser(ctx, "apple", sub, email, prof, clientMetaFromGin(c))
	if err != nil {
		s.writeAuthError(c, err)
		return
	}
	s.setRefreshCookie(c, pair.RefreshToken)
	apiresponse.OK(c, pair)
}

func (s *Service) appleOAuthConfig() (*oauth2.Config, error) {
	clientSecret, err := s.appleClientSecret()
	if err != nil {
		return nil, err
	}
	return &oauth2.Config{
		ClientID:     s.cfg.Auth.OAuthAppleClientID,
		ClientSecret: clientSecret,
		RedirectURL:  s.oauthRedirectBase() + "/auth/callback/apple",
		Scopes:       []string{"name", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://appleid.apple.com/auth/authorize",
			TokenURL: "https://appleid.apple.com/auth/token",
		},
	}, nil
}

func (s *Service) appleClientSecret() (string, error) {
	if strings.TrimSpace(s.cfg.Auth.OAuthAppleClientID) == "" ||
		strings.TrimSpace(s.cfg.Auth.OAuthAppleTeamID) == "" ||
		strings.TrimSpace(s.cfg.Auth.OAuthAppleKeyID) == "" ||
		strings.TrimSpace(s.cfg.Auth.OAuthApplePrivateKeyPEM) == "" {
		return "", errors.New("apple_not_configured")
	}
	key, err := parseApplePrivateKey(s.cfg.Auth.OAuthApplePrivateKeyPEM)
	if err != nil {
		return "", err
	}
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:    s.cfg.Auth.OAuthAppleTeamID,
		Subject:   s.cfg.Auth.OAuthAppleClientID,
		Audience:  jwt.ClaimStrings{"https://appleid.apple.com"},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(30 * 24 * time.Hour)),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tok.Header["kid"] = s.cfg.Auth.OAuthAppleKeyID
	return tok.SignedString(key)
}

func parseApplePrivateKey(raw string) (*ecdsa.PrivateKey, error) {
	raw = strings.ReplaceAll(raw, `\n`, "\n")
	block, _ := pem.Decode([]byte(raw))
	if block == nil {
		return nil, errors.New("apple_private_key_invalid")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		key, ok := parsed.(*ecdsa.PrivateKey)
		if !ok {
			return nil, errors.New("apple_private_key_not_ec")
		}
		return key, nil
	}
	return x509.ParseECPrivateKey(block.Bytes)
}

func parseAppleIDTokenUnverified(raw string) (jwt.MapClaims, error) {
	if raw == "" {
		return nil, errors.New("missing_id_token")
	}
	claims := jwt.MapClaims{}
	_, _, err := new(jwt.Parser).ParseUnverified(raw, claims)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

func appleEmailVerified(value any) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return strings.EqualFold(v, "true")
	default:
		return false
	}
}
