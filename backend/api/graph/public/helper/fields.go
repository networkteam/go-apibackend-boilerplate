package helper

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
)

type SelectedFieldsInfo struct {
	reqCtx    *graphql.OperationContext
	collected []graphql.CollectedField
}

func SelectedFields(ctx context.Context) *SelectedFieldsInfo {
	fieldContext := graphql.GetFieldContext(ctx)
	reqCtx := graphql.GetOperationContext(ctx)
	collected := graphql.CollectFields(reqCtx, fieldContext.Field.Selections, nil)

	return &SelectedFieldsInfo{
		reqCtx:    reqCtx,
		collected: collected,
	}
}

func (q *SelectedFieldsInfo) PathSelected(path ...string) bool {
	return isFieldPathSelected(q.reqCtx, q.collected, path)
}

func isFieldPathSelected(reqCtx *graphql.OperationContext, collected []graphql.CollectedField, fieldPathSelected []string) bool {
	if len(fieldPathSelected) == 0 {
		return false
	}

	for _, collectedField := range collected {
		if collectedField.Name == fieldPathSelected[0] {
			if len(fieldPathSelected) == 1 {
				return true
			}
			children := graphql.CollectFields(reqCtx, collectedField.Selections, nil)
			return isFieldPathSelected(reqCtx, children, fieldPathSelected[1:])
		}
	}

	return false
}
