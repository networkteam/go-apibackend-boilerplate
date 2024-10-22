package model_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"myvendor.mytld/myproject/backend/api/graph/model"
	"myvendor.mytld/myproject/backend/domain/types"
)

func TestUnmarshalDateScalar(t *testing.T) {
	tests := []struct {
		name    string
		v       any
		want    types.Date
		wantErr bool
	}{
		{
			name:    "empty string",
			v:       "",
			wantErr: true,
		},
		{
			name: "ISO date",
			v:    "2020-10-09",
			want: types.Date{Year: 2020, Month: time.October, Day: 9},
		},
		{
			name:    "RFC3339 date with time",
			v:       "2020-10-09T00:00:00.00000Z",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := model.UnmarshalDateScalar(tt.v)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			assert.Equal(t, tt.want, got)
		})
	}
}
