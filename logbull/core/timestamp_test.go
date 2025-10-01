package core

import (
	"sync"
	"testing"
	"time"
)

func TestGenerateUniqueTimestamp(t *testing.T) {
	t.Run("generates RFC3339Nano format", func(t *testing.T) {
		ts := GenerateUniqueTimestamp()

		_, err := time.Parse(time.RFC3339Nano, ts)
		if err != nil {
			t.Errorf("GenerateUniqueTimestamp() returned invalid RFC3339Nano format: %s, error: %v", ts, err)
		}
	})

	t.Run("generates unique timestamps", func(t *testing.T) {
		ts1 := GenerateUniqueTimestamp()
		ts2 := GenerateUniqueTimestamp()

		if ts1 == ts2 {
			t.Errorf("GenerateUniqueTimestamp() generated duplicate timestamps: %s", ts1)
		}
	})

	t.Run("generates monotonically increasing timestamps", func(t *testing.T) {
		timestamps := make([]string, 100)
		for i := 0; i < 100; i++ {
			timestamps[i] = GenerateUniqueTimestamp()
		}

		for i := 1; i < len(timestamps); i++ {
			if timestamps[i] <= timestamps[i-1] {
				t.Errorf("Timestamps not monotonically increasing: %s <= %s", timestamps[i], timestamps[i-1])
			}
		}
	})
}

func TestGenerateUniqueTimestampConcurrent(t *testing.T) {
	t.Run("thread-safe concurrent generation", func(t *testing.T) {
		const numGoroutines = 10
		const numTimestamps = 100

		timestamps := make([]string, numGoroutines*numTimestamps)
		var wg sync.WaitGroup

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()
				for j := 0; j < numTimestamps; j++ {
					timestamps[goroutineID*numTimestamps+j] = GenerateUniqueTimestamp()
				}
			}(i)
		}

		wg.Wait()

		seen := make(map[string]bool)
		for _, ts := range timestamps {
			if seen[ts] {
				t.Errorf("Duplicate timestamp generated: %s", ts)
			}
			seen[ts] = true
		}
	})
}

func TestFormatTimestamp(t *testing.T) {
	tests := []struct {
		name        string
		timestampNs int64
	}{
		{
			name:        "epoch time",
			timestampNs: 0,
		},
		{
			name:        "current time",
			timestampNs: time.Now().UnixNano(),
		},
		{
			name:        "specific time",
			timestampNs: 1700000000000000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted := formatTimestamp(tt.timestampNs)

			_, err := time.Parse(time.RFC3339Nano, formatted)
			if err != nil {
				t.Errorf("formatTimestamp() returned invalid RFC3339Nano format: %s, error: %v", formatted, err)
			}
		})
	}
}

func BenchmarkGenerateUniqueTimestamp(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateUniqueTimestamp()
	}
}

func BenchmarkGenerateUniqueTimestampParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			GenerateUniqueTimestamp()
		}
	})
}
