package deterministic

import "time"

// NowFunc returns a function that reports the current time deterministically,
// incrementing by one second with each call. The initial time is fixed at
// 2006-01-02T15:04:05Z. Each call returns the current time and advances the
// internal clock by one second for the next call. Each NowFunc instance
// maintains its own independent clock; advancing one instance does not affect
// other instances.
func NowFunc() func() time.Time {
	now, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")

	return func() time.Time {
		out := now
		now = now.Add(time.Second)

		return out
	}
}
