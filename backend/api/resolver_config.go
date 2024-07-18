package api

import "time"

type ResolverConfig struct {
	// Constant minimal time duration for sensitive operations (e.g. login / request password reset / perform password reset / registration)
	SensitiveOperationConstantTime time.Duration
}
