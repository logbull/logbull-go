package core

import (
	"sync"
	"time"
)

var (
	timestampMu     sync.Mutex
	lastTimestampNs int64
)

func GenerateUniqueTimestamp() string {
	timestampMu.Lock()
	defer timestampMu.Unlock()

	currentNs := time.Now().UnixNano()
	if currentNs <= lastTimestampNs {
		currentNs = lastTimestampNs + 1
	}
	lastTimestampNs = currentNs

	return formatTimestamp(currentNs)
}

func formatTimestamp(timestampNs int64) string {
	seconds := timestampNs / 1_000_000_000
	nanos := timestampNs % 1_000_000_000

	t := time.Unix(seconds, nanos).UTC()
	return t.Format("2006-01-02T15:04:05.000000000Z")
}
