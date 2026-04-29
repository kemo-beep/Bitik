package catalogsvc

import (
	"net/http/httptest"
	"testing"

	catalogstore "github.com/bitik/backend/internal/store/catalog"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestProductParamsFiltersSortingAndPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest("GET", "/?page=2&per_page=10&sort=price_asc&q=shirt&min_price_cents=100&max_price_cents=5000", nil)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req

	p := parsePagination(c)
	params, ok := productParams(c, p)
	if !ok {
		t.Fatal("expected valid product params")
	}

	if p.Page != 2 || p.PerPage != 10 || p.Offset != 10 {
		t.Fatalf("unexpected pagination: %+v", p)
	}
	if params.Sort != "price_asc" {
		t.Fatalf("unexpected sort: %s", params.Sort)
	}
	if !params.Query.Valid || params.Query.String != "shirt" {
		t.Fatalf("unexpected query: %+v", params.Query)
	}
	if !params.MinPriceCents.Valid || params.MinPriceCents.Int64 != 100 {
		t.Fatalf("unexpected min price: %+v", params.MinPriceCents)
	}
	if !params.MaxPriceCents.Valid || params.MaxPriceCents.Int64 != 5000 {
		t.Fatalf("unexpected max price: %+v", params.MaxPriceCents)
	}
}

func TestProductParamsFallsBackForUnknownSortAndLargePageSize(t *testing.T) {
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest("GET", "/?page=0&per_page=1000&sort=random", nil)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req

	p := parsePagination(c)
	params, ok := productParams(c, p)
	if !ok {
		t.Fatal("expected valid product params")
	}

	if p.Page != defaultPage || p.PerPage != maxPerPage {
		t.Fatalf("unexpected pagination fallback: %+v", p)
	}
	if params.Sort != "newest" {
		t.Fatalf("expected newest fallback, got %q", params.Sort)
	}
}

func TestProductParamsRejectsInvalidFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest("GET", "/?category_id=bad&min_price_cents=5000&max_price_cents=100", nil)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req

	if _, ok := productParams(c, parsePagination(c)); ok {
		t.Fatal("expected invalid filters")
	}
}

func TestRegisterRoutesDoesNotConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	svc := &Service{}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("RegisterRoutes panicked: %v", r)
		}
	}()
	svc.RegisterRoutes(v1)
}

func BenchmarkProductListMapping(b *testing.B) {
	rows := make([]catalogstore.ListProductsRow, 100)
	for i := range rows {
		rows[i] = catalogstore.ListProductsRow{
			Name:          "Benchmark Product",
			Slug:          "benchmark-product",
			Description:   pgtype.Text{String: "Fast list mapping benchmark", Valid: true},
			MinPriceCents: 1000,
			MaxPriceCents: 2000,
			Currency:      "USD",
			ReviewCount:   5,
		}
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = mapSlice(rows, productRowJSON)
	}
}
