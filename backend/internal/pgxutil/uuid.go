package pgxutil

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func UUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func ToUUID(p pgtype.UUID) (uuid.UUID, bool) {
	if !p.Valid {
		return uuid.Nil, false
	}
	return uuid.UUID(p.Bytes), true
}
