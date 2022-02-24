package types

import (
	"bytes"
	"database/sql"
	"encoding/json"
)

type NullInt64 struct {
	sql.NullInt64
}

// MarshalJSON marshals the NullInt64 as null or the nested int64
func (u NullInt64) MarshalJSON() ([]byte, error) {
	if !u.Valid {
		return json.Marshal(nil)
	}

	return json.Marshal(u.Int64)
}

// UnmarshalJSON unmarshals a NullInt64
func (u *NullInt64) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) {
		u.Int64, u.Valid = 0, false
		return nil
	}

	if err := json.Unmarshal(b, &u.Int64); err != nil {
		return err
	}

	u.Valid = true

	return nil
}
