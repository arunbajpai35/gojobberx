package main

import (
	"math"
	"testing"
	"time"
)

func TestExponentialBackoff(t *testing.T) {
	for attempt := 1; attempt <= 5; attempt++ {
		got := exponentialBackoff(attempt)
		base := time.Duration(math.Pow(2, float64(attempt))) * time.Second
		minD := base
		maxD := base + 500*time.Millisecond
		if got < minD || got > maxD {
			t.Errorf("attempt %d: got %v, want between %v and %v", attempt, got, minD, maxD)
		}
	}
}

func TestQueueByPriority(t *testing.T) {
	drain := func() {
		for {
			select {
			case <-highQueue:
			case <-mediumQueue:
			case <-lowQueue:
			default:
				return
			}
		}
	}
	drain()
	defer drain()

	cases := []struct {
		priority string
		want     chan *Job
	}{
		{"high", highQueue},
		{"medium", mediumQueue},
		{"low", lowQueue},
		{"", mediumQueue},
		{"garbage", mediumQueue},
	}

	for _, tc := range cases {
		t.Run(tc.priority, func(t *testing.T) {
			drain()
			job := &Job{ID: "x", Priority: tc.priority}
			queueByPriority(job)
			select {
			case got := <-tc.want:
				if got != job {
					t.Fatalf("priority %q: wrong job in channel", tc.priority)
				}
			case <-time.After(100 * time.Millisecond):
				t.Fatalf("priority %q: nothing in expected channel", tc.priority)
			}
		})
	}
}
