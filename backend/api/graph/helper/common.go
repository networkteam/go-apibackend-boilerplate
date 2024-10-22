package helper

import (
	"errors"

	"github.com/gofrs/uuid"

	"myvendor.mytld/myproject/backend/domain/types"
	"myvendor.mytld/myproject/backend/finder"
)

const (
	defaultPerPage = 50
	maxPerPage     = 1000
)

var ErrMaxPerPageExceeded = errors.New("perPage exceeds maximum")

func MapToPaging(page *int, perPage *int, sortField *string, sortOrder *string) (finder.Paging, error) {
	paging := finder.Paging{
		// Always apply a default per page for API listings
		PerPage:   ToPtr(defaultPerPage),
		SortField: sortField,
		SortOrder: sortOrder,
	}
	if page != nil {
		paging.Page = *page
	}
	if perPage != nil {
		paging.PerPage = perPage
	}
	if paging.PerPage != nil && *paging.PerPage > maxPerPage {
		return paging, ErrMaxPerPageExceeded
	}

	return paging, nil
}

func ToPtr[T any](value T) *T {
	return &value
}

func ToVal[T any](ptr *T) T {
	if ptr == nil {
		var value T
		return value
	}
	return *ptr
}

func uuidOrNil(id uuid.NullUUID) *uuid.UUID {
	if id.Valid {
		return &id.UUID
	}
	return nil
}

func ToNullUUID(id *uuid.UUID) uuid.NullUUID {
	if id == nil {
		return uuid.NullUUID{Valid: false}
	}
	return uuid.NullUUID{UUID: *id, Valid: true}
}

func ToNullDate(date *types.Date) types.NullDate {
	if date == nil {
		return types.NullDate{Valid: false}
	}
	return types.NullDate{Date: *date, Valid: true}
}
