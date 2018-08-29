// Package utils contains various common utils separate by utility types
package utils

import (
	"time"
)

// SecsFromEpochToTime converts an int64 of seconds from epoch to Time struct
func SecsFromEpochToTime(ts int64) time.Time {
	return time.Unix(ts, 0)
}

// NanoSecsFromEpochToTime converts an int64 of nanoseconds from epoch to Time struct
func NanoSecsFromEpochToTime(ts int64) time.Time {
	return time.Unix(0, ts)
}

// TimeToSecsFromEpoch converts a time.Time struct to nanoseconds from epoch.
func TimeToSecsFromEpoch(t *time.Time) int64 {
	return t.Unix()
}

// TimeToNanoSecsFromEpoch converts a time.Time struct to nanoseconds from epoch.
func TimeToNanoSecsFromEpoch(t *time.Time) int64 {
	return t.UnixNano()
}
