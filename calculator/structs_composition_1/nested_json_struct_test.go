package structscomposition1

import (
	"testing"
	"time"
)

func TestPromotedFieldAccess(t *testing.T) {
	createdAt := time.Date(2026, time.May, 10, 12, 0, 0, 0, time.UTC)

	user := User{
		BaseEntity: BaseEntity{
			ID:        "u-100",
			CreatedAt: createdAt,
		},
		Name: "Khizer",
	}

	if user.ID != "u-100" {
		t.Fatalf("expected user.ID to be u-100, got %s", user.ID)
	}

	if !user.CreatedAt.Equal(createdAt) {
		t.Fatalf("expected user.CreatedAt to be %v, got %v", createdAt, user.CreatedAt)
	}

	if user.BaseEntity.ID != user.ID {
		t.Fatalf("expected embedded ID and promoted ID to match, got %s and %s", user.BaseEntity.ID, user.ID)
	}
}
