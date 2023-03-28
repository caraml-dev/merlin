package redis

import (
	"context"
	"errors"
	"time"

	"github.com/caraml-dev/merlin/pkg/transformer"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	hitConnStats     = "hits"
	missConnStats    = "misses"
	timeoutConnStats = "timeouts"
	idleConnStats    = "idle_conns"
	staleConnStats   = "stale_conns"
	totalConnStats   = "total_conns"

	success = "success"
	fail    = "fail"
)

var (
	redisNewConn = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: transformer.PromNamespace,
		Name:      "redis_new_connection_count",
		Help:      "New connection established to redis server",
	}, []string{})

	redisPipeline = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: transformer.PromNamespace,
		Name:      "redis_pipelined_count",
		Help:      "Number of pipelined redis command",
	}, []string{"status", "command"})

	redisCommandLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: transformer.PromNamespace,
		Name:      "redis_command_latency_ms",
		Help:      "Latency of redis command",
		Buckets:   prometheus.ExponentialBuckets(1, 2, 10), // 1,2,4,8,16,32,64,128,256,512,+Inf
	}, []string{"command"})

	redisConnPoolStats = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: transformer.PromNamespace,
		Name:      "redis_conn_stats",
		Help:      "Redis connection stats",
	}, []string{"stat"})
)

type redisHook struct{}

type requestStartKey struct{}

func (h *redisHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	return context.WithValue(ctx, requestStartKey{}, time.Now()), nil
}

func (h *redisHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	startTime, ok := ctx.Value(requestStartKey{}).(time.Time)
	if !ok {
		return nil
	}

	duration := time.Since(startTime).Milliseconds()
	redisCommandLatency.WithLabelValues(cmd.Name()).Observe(float64(duration))
	return nil
}

func (h *redisHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	return context.WithValue(ctx, requestStartKey{}, time.Now()), nil
}

func (h *redisHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	if err := h.AfterProcess(ctx, redis.NewCmd(ctx, "pipeline")); err != nil {
		return err
	}

	for _, cmd := range cmds {
		if isActualError(cmd.Err()) {
			redisPipeline.WithLabelValues(fail, cmd.Name()).Inc()
			continue
		}
		redisPipeline.WithLabelValues(success, cmd.Name()).Inc()
	}
	return nil
}

func isActualError(err error) bool {
	return err != nil && !errors.Is(err, redis.Nil)
}
