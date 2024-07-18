package helper

import (
	"context"
	"time"

	"github.com/apex/log"
)

type ConstantTimeWaiter struct {
	started time.Time
	delay   time.Duration
}

func ConstantTime(delay time.Duration) *ConstantTimeWaiter {
	started := time.Now()
	return &ConstantTimeWaiter{started, delay}
}

func (w *ConstantTimeWaiter) Wait(ctx context.Context) {
	if w.delay == 0 {
		return
	}

	waitTime := w.delay - time.Since(w.started)
	if waitTime <= 0 {
		log.Warnf("Constant time operation exceeded delay by %s", -waitTime)
		return
	}

	log.Debugf("Waiting %s to ensure constant time for operation", waitTime)

	// Wait until delay since start has passed (if needed to wait) or the context is done
	select {
	case <-time.After(waitTime):
	case <-ctx.Done():
	}
}
