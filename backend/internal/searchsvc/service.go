package searchsvc

import (
	"net/http"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/config"
	"github.com/bitik/backend/internal/middleware"
	platformsearch "github.com/bitik/backend/internal/platform/search"
	catalogstore "github.com/bitik/backend/internal/store/catalog"
	searchstore "github.com/bitik/backend/internal/store/search"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Service struct {
	cfg     config.Config
	log     *zap.Logger
	catalog *catalogstore.Queries
	search  *searchstore.Queries
	pool    *pgxpool.Pool

	os *platformsearch.Client
}

func NewService(cfg config.Config, logger *zap.Logger, pool *pgxpool.Pool, osClient *platformsearch.Client) *Service {
	return &Service{
		cfg:     cfg,
		log:     logger,
		catalog: catalogstore.New(pool),
		search:  searchstore.New(pool),
		pool:    pool,
		os:      osClient,
	}
}

func (s *Service) RegisterRoutes(rg *gin.RouterGroup) {
	public := rg.Group("/public")
	public.GET("/search", s.HandleSearch)
	public.GET("/search/suggestions", s.HandleSuggestions)
	public.GET("/search/trending", s.HandleTrending)
	public.GET("/search/recent", s.HandleRecent)
	public.DELETE("/search/recent", s.HandleClearRecent)
	public.POST("/search/click", s.HandleClick)

	internal := rg.Group("/internal/jobs", middleware.RequireInternalAPI(s.cfg))
	internal.POST("/index-product", s.HandleIndexProduct)
	internal.POST("/reindex-products", s.HandleReindexProducts)
}

func currentUserID(c *gin.Context) (uuid.UUID, bool) {
	raw, ok := c.Get(middleware.AuthUserIDKey)
	if !ok {
		return uuid.Nil, false
	}
	id, ok := raw.(uuid.UUID)
	return id, ok
}

func (s *Service) requireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, _ := c.Get(middleware.AuthRolesKey)
		got, _ := raw.([]string)
		for _, r := range got {
			for _, want := range roles {
				if r == want {
					c.Next()
					return
				}
			}
		}
		apiresponse.Error(c, http.StatusForbidden, "forbidden", "Required role is missing.")
		c.Abort()
	}
}
