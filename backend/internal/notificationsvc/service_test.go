package notificationsvc

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParsePageBounds(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/?page=2&limit=500", nil)
	p := parsePage(c)
	if p.Page != 2 {
		t.Fatalf("page expected 2 got %d", p.Page)
	}
	if p.Limit != 100 {
		t.Fatalf("limit expected capped 100 got %d", p.Limit)
	}
	if p.Offset != 100 {
		t.Fatalf("offset expected 100 got %d", p.Offset)
	}
}

func TestPageMeta(t *testing.T) {
	meta := pageMeta(pageParams{Page: 2, Limit: 20}, 100)
	if meta["has_next"] != true {
		t.Fatalf("expected has_next true, got %v", meta["has_next"])
	}
}
