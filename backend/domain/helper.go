package domain

import (
	"strings"

	"github.com/gofrs/uuid"
)

func IsBlank(s string) bool {
	return strings.Trim(s, " ") == ""
}

type UUIDSet map[uuid.UUID]struct{}

func NewUUIDSet(size ...int) UUIDSet {
	n := 0
	if len(size) > 0 {
		n = size[0]
	}
	return make(UUIDSet, n)
}

func (s UUIDSet) Add(id uuid.UUID) {
	s[id] = struct{}{}
}

func (s UUIDSet) Has(id uuid.UUID) bool {
	_, ok := s[id]
	return ok
}
