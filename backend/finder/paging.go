package finder

import "myvendor.mytld/myproject/backend/persistence/repository"

type Paging struct {
	Page      int
	PerPage   *int
	SortField *string
	SortOrder *string
}

const defaultSortOrder = repository.SortOrderAsc

func (p Paging) options() []repository.PagingOption {
	var opts []repository.PagingOption
	if p.PerPage != nil {
		perPage := *p.PerPage
		opts = append(opts, repository.WithLimit(perPage))
		if p.Page > 0 {
			opts = append(opts, repository.WithOffset(p.Page*perPage))
		}
	}
	if p.SortField != nil {
		sortOrder := defaultSortOrder
		if p.SortOrder != nil {
			sortOrder = *p.SortOrder
		}
		opts = append(opts, repository.WithSort(*p.SortField, sortOrder))
	}
	return opts
}
