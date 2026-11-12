package imagebuilder

import (
	"context"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	kerrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/caraml-dev/merlin/log"
)

var (
	// kubeCallTimeout is the deadline of a single Kubernetes API call (Get/Create/Delete job).
	// It is applied to a sub-context only, the caller's context is never given a deadline.
	kubeCallTimeout = 10 * time.Second
	// kubeCallMaxRetries is the maximum number of retries (on top of the first attempt) of a Kubernetes API call.
	kubeCallMaxRetries uint64 = 3
	// kubeCallInitialInterval is the initial backoff interval between retries of a Kubernetes API call.
	kubeCallInitialInterval = 500 * time.Millisecond
)

// retryKubeCall runs fn with exponential backoff retry. Every attempt runs with its own sub-context of ctx
// bounded by kubeCallTimeout, so a stalled Kubernetes API call is abandoned and retried instead of blocking forever.
// ctx itself is never given a deadline, but its cancellation stops the retries.
func retryKubeCall(ctx context.Context, operation string, fn func(callCtx context.Context) error) error {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = kubeCallInitialInterval
	policy := backoff.WithContext(backoff.WithMaxRetries(b, kubeCallMaxRetries), ctx)

	attempt := 0
	return backoff.Retry(func() error {
		attempt++
		callCtx, cancel := context.WithTimeout(ctx, kubeCallTimeout)
		defer cancel()

		err := fn(callCtx)
		if err == nil {
			return nil
		}
		if ctx.Err() != nil || isPermanentKubeError(err) {
			return backoff.Permanent(err)
		}
		log.Warnf("%s failed (attempt %d), will retry: %v", operation, attempt, err)
		return err
	}, policy)
}

// isPermanentKubeError returns true for errors that will not go away by retrying the same request.
func isPermanentKubeError(err error) bool {
	return kerrors.IsNotFound(err) ||
		kerrors.IsAlreadyExists(err) ||
		kerrors.IsForbidden(err) ||
		kerrors.IsUnauthorized(err) ||
		kerrors.IsInvalid(err) ||
		kerrors.IsBadRequest(err)
}
