package internal

import (
	"math/rand"
	"time"

	ulid "github.com/oklog/ulid/v2"
)

func generateId() (ulid.ULID, error) {
	entropy := rand.New(rand.NewSource(time.Now().UnixNano()))
	ms := ulid.Timestamp(time.Now())
	return ulid.New(ms, entropy)
}
