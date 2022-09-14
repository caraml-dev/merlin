package hystrix

import (
	"time"
)

const (
	maxUint = ^uint(0)
	maxInt  = int(maxUint >> 1)
)

func DurationToInt(duration, unit time.Duration) int {
	durationAsNumber := duration / unit

	if int64(durationAsNumber) > int64(maxInt) {
		// Returning max possible value seems like best possible solution here
		// the alternative is to panic as there is no way of returning an error
		// without changing the NewClient API
		return maxInt
	}
	return int(durationAsNumber)
}
