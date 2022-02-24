package helper

import (
	"github.com/gofrs/uuid"

	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/persistence/repository"
	"myvendor.mytld/myproject/backend/persistence/types"
)

func MapToPaging(page *int, perPage *int, sortField *string, sortOrder *string) repository.Paging {
	paging := repository.Paging{
		// Always apply a default per page for API listings (max is checked by repository)
		PerPage:   intPtr(repository.DefaultPerPage),
		SortField: sortField,
		SortOrder: sortOrder,
	}
	if page != nil {
		paging.Page = *page
	}
	if perPage != nil {
		paging.PerPage = perPage
	}

	return paging
}

func intPtr(v int) *int {
	return &v
}

func intPtrOrNil(v types.NullInt64) *int {
	if !v.Valid {
		return nil
	}
	i := int(v.Int64)
	return &i
}

func IntPtrToNullInt64(v *int) (result types.NullInt64) {
	if v != nil {
		result.Valid = true
		result.Int64 = int64(*v)
	}
	return result
}

func uuidOrNil(id uuid.NullUUID) *uuid.UUID {
	if id.Valid {
		return &id.UUID
	}
	return nil
}

func StrVal(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}

func BoolVal(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func NullUUIDVal(id *uuid.UUID) uuid.NullUUID {
	if id == nil {
		return uuid.NullUUID{Valid: false}
	}
	return uuid.NullUUID{UUID: *id, Valid: true}
}

func dateOrNil(date domain.NullDate) *domain.Date {
	if date.Valid {
		return &date.Date
	}
	return nil
}

func NullDateVal(date *domain.Date) domain.NullDate {
	if date == nil {
		return domain.NullDate{Valid: false}
	}
	return domain.NullDate{Date: *date, Valid: true}
}
