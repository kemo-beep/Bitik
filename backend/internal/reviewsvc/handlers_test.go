package reviewsvc

import (
	"testing"
	"time"

	reviewstore "github.com/bitik/backend/internal/store/reviews"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestReviewJSON_IncludesSellerReplyFields(t *testing.T) {
	now := time.Now().UTC()
	id := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	pid := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	uid := pgtype.UUID{Bytes: uuid.New(), Valid: true}

	row := reviewstore.ProductReview{
		ID:                 id,
		ProductID:          pid,
		UserID:             uid,
		Rating:             5,
		Title:              pgtype.Text{String: "Great", Valid: true},
		Body:               pgtype.Text{String: "Nice product", Valid: true},
		IsVerifiedPurchase: true,
		IsHidden:           false,
		CreatedAt:          pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:          pgtype.Timestamptz{Time: now, Valid: true},
		SellerReply:        pgtype.Text{String: "Thanks!", Valid: true},
		SellerReplyAt:      pgtype.Timestamptz{Time: now, Valid: true},
	}

	out := reviewJSON(row)
	if out["seller_reply"] != "Thanks!" {
		t.Fatalf("expected seller_reply to be set, got %#v", out["seller_reply"])
	}
	if out["seller_reply_at"] == nil {
		t.Fatalf("expected seller_reply_at to be non-nil")
	}
}

func TestReviewJSON_OmitsInvalidSellerReplyAt(t *testing.T) {
	id := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	pid := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	uid := pgtype.UUID{Bytes: uuid.New(), Valid: true}

	row := reviewstore.ProductReview{
		ID:            id,
		ProductID:     pid,
		UserID:        uid,
		Rating:        4,
		CreatedAt:     pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		UpdatedAt:     pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		SellerReplyAt: pgtype.Timestamptz{},
	}

	out := reviewJSON(row)
	if out["seller_reply_at"] != nil {
		t.Fatalf("expected seller_reply_at to be nil, got %#v", out["seller_reply_at"])
	}
}

