package promotionsvc

import (
	"testing"

	"github.com/bitik/backend/internal/middleware"
	orderstore "github.com/bitik/backend/internal/store/orders"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestCurrentUserID(t *testing.T) {
	c, _ := gin.CreateTestContext(nil)
	id := uuid.New()
	c.Set(middleware.AuthUserIDKey, id)
	got, ok := currentUserID(c)
	if !ok || got != id {
		t.Fatalf("currentUserID mismatch, ok=%v got=%v", ok, got)
	}
}

func TestSellerFromContext(t *testing.T) {
	c, _ := gin.CreateTestContext(nil)
	seller := orderstore.Seller{ShopName: "x"}
	c.Set("seller", seller)
	got, ok := sellerFromContext(c)
	if !ok || got.ShopName != "x" {
		t.Fatalf("sellerFromContext mismatch, ok=%v got=%v", ok, got.ShopName)
	}
}
