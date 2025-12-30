package helper

import (
	"myvendor.mytld/myproject/backend/api/graph/public/model"
)

func SingleFieldsError(fieldName string, errorCode string) *model.FieldsError {
	return &model.FieldsError{
		Errors: []*model.FieldError{
			{
				Path: []string{fieldName},
				Code: errorCode,
			},
		},
	}
}
