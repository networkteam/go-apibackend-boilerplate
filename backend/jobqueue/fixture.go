package jobqueue

import (
	"context"
	"sync"

	"myvendor.mytld/myproject/backend/domain/command"
)

type FixtureQueue struct {
	mx sync.Mutex

	AccountSendWelcomeEmailCallCount           int
	AccountSendWelcomeEmailLastCmd             command.AccountSendWelcomeEmail
	AccountSendWelcomeEmailLastDeferProcessing bool
}

var _ Queue = &FixtureQueue{}

func NewFixture() *FixtureQueue {
	return &FixtureQueue{}
}

func (i *FixtureQueue) AccountSendWelcomeEmail(_ context.Context, cmd command.AccountSendWelcomeEmail, deferProcessing bool) error {
	i.mx.Lock()
	defer i.mx.Unlock()

	i.AccountSendWelcomeEmailCallCount++
	i.AccountSendWelcomeEmailLastCmd = cmd
	i.AccountSendWelcomeEmailLastDeferProcessing = deferProcessing

	return nil
}

func (i *FixtureQueue) Close() error {
	return nil
}
