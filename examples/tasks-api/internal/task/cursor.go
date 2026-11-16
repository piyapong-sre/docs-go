package task

import (
	"encoding/base64"
	"encoding/json"
	"errors"
)

var errBadCursor = errors.New("invalid cursor")

// EncodeCursor turns a cursor into an opaque string for clients.
// Clients must treat it as a black box, which lets us change its
// contents later without breaking them.
func EncodeCursor(c Cursor) string {
	b, _ := json.Marshal(c) // marshalling this struct cannot fail
	return base64.RawURLEncoding.EncodeToString(b)
}

// DecodeCursor parses a string produced by EncodeCursor.
func DecodeCursor(s string) (Cursor, error) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return Cursor{}, errBadCursor
	}
	var c Cursor
	if err := json.Unmarshal(b, &c); err != nil || c.ID == "" || c.CreatedAt.IsZero() {
		return Cursor{}, errBadCursor
	}
	return c, nil
}
