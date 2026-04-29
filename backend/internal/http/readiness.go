package httpapi

import (
	"context"
	"errors"
	"sync"
	"time"
)

type CheckFunc func(context.Context) error

type Readiness struct {
	mu     sync.RWMutex
	checks map[string]CheckFunc
}

func NewReadiness() *Readiness {
	return &Readiness{checks: map[string]CheckFunc{}}
}

func (r *Readiness) Register(name string, check CheckFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checks[name] = check
}

func (r *Readiness) Check(ctx context.Context) (map[string]string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	status := map[string]string{}
	ready := true

	for name, check := range r.checks {
		checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		err := check(checkCtx)
		cancel()

		if err != nil {
			status[name] = err.Error()
			ready = false
			continue
		}
		status[name] = "ok"
	}

	if len(status) == 0 {
		status["application"] = "ok"
	}

	return status, ready
}

func Unavailable(name string) CheckFunc {
	return func(context.Context) error {
		return errors.New(name + " unavailable")
	}
}
