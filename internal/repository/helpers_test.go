package repository

import (
	"errors"
	"testing"
	"time"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	"github.com/google/uuid"
)

type fakeRow struct {
	values []any
	err    error
}

func (r fakeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	for i := range dest {
		switch d := dest[i].(type) {
		case *uuid.UUID:
			*d = r.values[i].(uuid.UUID)
		case *string:
			*d = r.values[i].(string)
		case **string:
			*d = r.values[i].(*string)
		case *time.Time:
			*d = r.values[i].(time.Time)
		case *int16:
			*d = r.values[i].(int16)
		case **int16:
			*d = r.values[i].(*int16)
		case *int64:
			*d = r.values[i].(int64)
		case **int64:
			*d = r.values[i].(*int64)
		case *domain.Sex:
			*d = r.values[i].(domain.Sex)
		default:
			return errors.New("unsupported dest type")
		}
	}
	return nil
}

type fakeRows struct {
	idx  int
	rows []fakeRow
	err  error
}

func (r *fakeRows) Next() bool {
	return r.idx < len(r.rows)
}

func (r *fakeRows) Scan(dest ...any) error {
	row := r.rows[r.idx]
	r.idx++
	return row.Scan(dest...)
}

func (r *fakeRows) Err() error { return r.err }
func (r *fakeRows) Close()     {}

func TestScanUser(t *testing.T) {
	id := uuid.Must(uuid.NewV7())
	now := time.Now()
	var (
		bio         *string
		alcohol     *string
		smoking     *string
		heightCM    *int16
		cityID      *int64
		primaryKey  = "k"
		primaryURL  = "u"
		firstName   = "A"
		lastName    = "B"
		email       = "a@b.com"
		birthDate   = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		toilerScore = int16(7)
		sex         = domain.SexMale
	)

	u, err := ScanUser(fakeRow{values: []any{
		id, firstName, lastName, email, birthDate,
		bio, toilerScore, alcohol, smoking, sex,
		heightCM, cityID, primaryKey, primaryURL, now, now,
	}})
	if err != nil {
		t.Fatalf("ScanUser: %v", err)
	}
	if u.ID != id || u.Email != email {
		t.Fatalf("unexpected user")
	}
}

func TestCollectExtraPhotos(t *testing.T) {
	uid := uuid.Must(uuid.NewV7())
	now := time.Now()

	rows := &fakeRows{rows: []fakeRow{
		{values: []any{int64(1), uid, "k1", "u1", int16(1), now}},
		{values: []any{int64(2), uid, "k2", "u2", int16(2), now}},
	}}
	photos, err := CollectExtraPhotos(rows)
	if err != nil {
		t.Fatalf("CollectExtraPhotos: %v", err)
	}
	if len(photos) != 2 || photos[0].ID != 1 {
		t.Fatalf("unexpected photos")
	}
}
