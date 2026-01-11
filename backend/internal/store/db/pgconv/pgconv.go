// Package pgconv provides small conversions used by storage and sync code.
package pgconv

import (
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// IntToInt32 converts an int to int32 and returns 0 when the value is out of range.
func IntToInt32(v int) int32 {
	if v > math.MaxInt32 || v < math.MinInt32 {
		return 0
	}
	return int32(v)
}

// Uint32ToInt32 converts a uint32 to int32 and returns 0 when the value is out of range.
func Uint32ToInt32(v uint32) int32 {
	if v > math.MaxInt32 {
		return 0
	}
	return int32(v)
}

func Int32ToUint32(v int32) uint32 {
	if v < 0 {
		return 0
	}
	return uint32(v)
}

// StrPtr returns a pointer to s unless it is empty.
func StrPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// TextOrNull converts an optional string to a pgtype.Text value.
func TextOrNull(v *string) pgtype.Text {
	if v == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *v, Valid: true}
}

// TextVal converts a pgtype.Text value into a string pointer.
func TextVal(v pgtype.Text) *string {
	if !v.Valid {
		return nil
	}
	s := v.String
	return &s
}

// Int32OrNull converts an optional int32 to pgtype.Int4.
func Int32OrNull(v *int32) pgtype.Int4 {
	if v == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *v, Valid: true}
}

// Int32Val converts a pgtype.Int4 to an int32 pointer.
func Int32Val(v pgtype.Int4) *int32 {
	if !v.Valid {
		return nil
	}
	val := v.Int32
	return &val
}

// TimeOrNull converts an optional time to pgtype.Timestamptz.
func TimeOrNull(v *time.Time) pgtype.Timestamptz {
	if v == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *v, Valid: true}
}

// TimeVal converts a pgtype.Timestamptz to a time pointer.
func TimeVal(v pgtype.Timestamptz) *time.Time {
	if !v.Valid {
		return nil
	}
	t := v.Time
	return &t
}

// TextArray returns a non nil slice for TEXT[] columns.
func TextArray(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

// LimitOffset converts limit and offset to int32.
func LimitOffset(limit, offset int) (int32, int32) {
	return IntToInt32(limit), IntToInt32(offset)
}
