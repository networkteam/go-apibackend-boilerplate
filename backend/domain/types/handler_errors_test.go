package types_test

import (
	"testing"

	"github.com/friendsofgo/errors"
	"github.com/stretchr/testify/assert"

	"myvendor.mytld/myproject/backend/domain/types"
)

func TestHandlerLockedError(t *testing.T) {
	t.Run("error message", func(t *testing.T) {
		err := types.ErrHandlerLocked
		assert.Equal(t, "handler is locked by another process", err.Error())
	})

	t.Run("errors.Is works with direct error", func(t *testing.T) {
		err := types.ErrHandlerLocked
		assert.True(t, errors.Is(err, types.ErrHandlerLocked))
	})

	t.Run("errors.Is works with wrapped error", func(t *testing.T) {
		err := errors.Wrap(types.ErrHandlerLocked, "failed to process")
		assert.True(t, errors.Is(err, types.ErrHandlerLocked))
	})

	t.Run("errors.Is works with multiple wrappings", func(t *testing.T) {
		err := errors.Wrap(types.ErrHandlerLocked, "inner error")
		err = errors.Wrap(err, "middle error")
		err = errors.Wrap(err, "outer error")
		assert.True(t, errors.Is(err, types.ErrHandlerLocked))
	})

	t.Run("IsHandlerLockedError helper works", func(t *testing.T) {
		err := errors.Wrap(types.ErrHandlerLocked, "wrapped error")
		assert.True(t, types.IsHandlerLockedError(err))
	})

	t.Run("errors.Is returns false for different error", func(t *testing.T) {
		err := errors.New("some other error")
		assert.False(t, errors.Is(err, types.ErrHandlerLocked))
		assert.False(t, types.IsHandlerLockedError(err))
	})

	t.Run("errors.Is returns false for nil", func(t *testing.T) {
		assert.False(t, errors.Is(nil, types.ErrHandlerLocked))
		assert.False(t, types.IsHandlerLockedError(nil))
	})
}
