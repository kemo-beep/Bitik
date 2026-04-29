package sellersvc

import (
	"encoding/json"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"github.com/bitik/backend/internal/config"
	"github.com/bitik/backend/internal/middleware"
	"github.com/bitik/backend/internal/pgxutil"
	"github.com/bitik/backend/internal/platform/queue"
	sellerstore "github.com/bitik/backend/internal/store/sellers"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const (
	defaultPage    = int32(1)
	defaultPerPage = int32(25)
	maxPerPage     = int32(100)
)

var slugCleaner = regexp.MustCompile(`[^a-z0-9]+`)

type Service struct {
	cfg     config.Config
	log     *zap.Logger
	pool    *pgxpool.Pool
	queries *sellerstore.Queries
	queue   *queue.Producer
}

func NewService(cfg config.Config, logger *zap.Logger, pool *pgxpool.Pool) *Service {
	return &Service{
		cfg:     cfg,
		log:     logger,
		pool:    pool,
		queries: sellerstore.New(pool),
	}
}

func (s *Service) SetQueueProducer(p *queue.Producer) {
	s.queue = p
}

type pageParams struct {
	Page    int32
	PerPage int32
	Limit   int32
	Offset  int32
}

func pagination(c *gin.Context) pageParams {
	page := parsePositiveInt32(c.DefaultQuery("page", "1"), defaultPage)
	perPage := parsePositiveInt32(c.DefaultQuery("per_page", strconv.Itoa(int(defaultPerPage))), defaultPerPage)
	if perPage > maxPerPage {
		perPage = maxPerPage
	}
	return pageParams{Page: page, PerPage: perPage, Limit: perPage, Offset: (page - 1) * perPage}
}

func pageMeta(p pageParams, total int64) gin.H {
	pages := int64(0)
	if p.PerPage > 0 {
		pages = (total + int64(p.PerPage) - 1) / int64(p.PerPage)
	}
	return gin.H{"page": p.Page, "per_page": p.PerPage, "total": total, "total_pages": pages}
}

func parsePositiveInt32(raw string, fallback int32) int32 {
	v, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 32)
	if err != nil || v < 1 {
		return fallback
	}
	return int32(v)
}

func currentUserID(c *gin.Context) (uuid.UUID, bool) {
	raw, ok := c.Get(middleware.AuthUserIDKey)
	if !ok {
		return uuid.Nil, false
	}
	id, ok := raw.(uuid.UUID)
	return id, ok
}

func hasRole(c *gin.Context, allowed ...string) bool {
	raw, _ := c.Get(middleware.AuthRolesKey)
	roles, _ := raw.([]string)
	for _, role := range roles {
		for _, want := range allowed {
			if role == want {
				return true
			}
		}
	}
	return false
}

func toSlug(value string) string {
	slug := strings.ToLower(strings.TrimSpace(value))
	slug = slugCleaner.ReplaceAllString(slug, "-")
	return strings.Trim(slug, "-")
}

func text(v string) pgtype.Text {
	v = strings.TrimSpace(v)
	return pgtype.Text{String: v, Valid: v != ""}
}

func optionalUUID(raw string) (pgtype.UUID, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return pgtype.UUID{}, true
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return pgtype.UUID{}, false
	}
	return pgxutil.UUID(id), true
}

func uuidParam(c *gin.Context, name string) (pgtype.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		return pgtype.UUID{}, false
	}
	return pgxutil.UUID(id), true
}

func int8Ptr(v *int64) pgtype.Int8 {
	if v == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *v, Valid: true}
}

func int4Ptr(v *int32) pgtype.Int4 {
	if v == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *v, Valid: true}
}

func boolPtr(v *bool) pgtype.Bool {
	if v == nil {
		return pgtype.Bool{}
	}
	return pgtype.Bool{Bool: *v, Valid: true}
}

func jsonOrEmpty(v any) []byte {
	if v == nil {
		return []byte(`{}`)
	}
	b, err := json.Marshal(v)
	if err != nil {
		return []byte(`{}`)
	}
	return b
}

func jsonOptional(v any) []byte {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}

func uuidString(id pgtype.UUID) string {
	if v, ok := pgxutil.ToUUID(id); ok {
		return v.String()
	}
	return ""
}

func nullableUUID(id pgtype.UUID) any {
	if !id.Valid {
		return nil
	}
	return uuidString(id)
}

func textValue(t pgtype.Text) any {
	if !t.Valid {
		return nil
	}
	return t.String
}

func int8Value(v pgtype.Int8) any {
	if !v.Valid {
		return nil
	}
	return v.Int64
}

func int4Value(v pgtype.Int4) any {
	if !v.Valid {
		return nil
	}
	return v.Int32
}

func numericString(n pgtype.Numeric) string {
	if !n.Valid || n.Int == nil {
		return "0"
	}
	r := new(big.Rat).SetInt(n.Int)
	if n.Exp > 0 {
		r.Mul(r, new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n.Exp)), nil)))
	} else if n.Exp < 0 {
		r.Quo(r, new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-n.Exp)), nil)))
	}
	return r.FloatString(2)
}

func statusString(v any) string {
	if v == nil {
		return ""
	}
	return strings.Trim(v.(string), "\"")
}
