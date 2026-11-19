package imagebuilder

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func withFastKubeRetry(t *testing.T, timeout time.Duration, retries uint64) {
	oldTimeout, oldRetries, oldInterval := kubeCallTimeout, kubeCallMaxRetries, kubeCallInitialInterval
	kubeCallTimeout, kubeCallMaxRetries, kubeCallInitialInterval = timeout, retries, time.Millisecond
	t.Cleanup(func() {
		kubeCallTimeout, kubeCallMaxRetries, kubeCallInitialInterval = oldTimeout, oldRetries, oldInterval
	})
}

func TestRetryKubeCall_StalledCallIsRetriedWithoutDeadlineOnParent(t *testing.T) {
	withFastKubeRetry(t, 20*time.Millisecond, 3)

	ctx := context.Background()
	calls := 0
	err := retryKubeCall(ctx, "test", func(callCtx context.Context) error {
		calls++
		_, hasDeadline := callCtx.Deadline()
		assert.True(t, hasDeadline, "sub-context must have a deadline")
		if calls < 3 {
			<-callCtx.Done() // stalled call
			return callCtx.Err()
		}
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 3, calls)
	_, parentHasDeadline := ctx.Deadline()
	assert.False(t, parentHasDeadline)
}

func TestRetryKubeCall_GivesUpAfterMaxRetries(t *testing.T) {
	withFastKubeRetry(t, time.Second, 2)

	calls := 0
	err := retryKubeCall(context.Background(), "test", func(context.Context) error {
		calls++
		return errors.New("boom")
	})

	assert.EqualError(t, err, "boom")
	assert.Equal(t, 3, calls) // first attempt + 2 retries
}

func TestRetryKubeCall_PermanentErrorNotRetried(t *testing.T) {
	withFastKubeRetry(t, time.Second, 3)

	calls := 0
	notFound := kerrors.NewNotFound(schema.ParseGroupResource("jobs.batch"), "job")
	err := retryKubeCall(context.Background(), "test", func(context.Context) error {
		calls++
		return notFound
	})

	assert.True(t, kerrors.IsNotFound(err))
	assert.Equal(t, 1, calls)
}

func TestRetryKubeCall_ParentCancellationStopsRetries(t *testing.T) {
	withFastKubeRetry(t, time.Second, 5)

	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	err := retryKubeCall(ctx, "test", func(context.Context) error {
		calls++
		cancel()
		return errors.New("boom")
	})

	assert.Error(t, err)
	assert.Equal(t, 1, calls)
}
