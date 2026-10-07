package ratelimits

import (
	"fmt"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestRateLimitDrop(t *testing.T) {
	limiter := rate.NewLimiter(rate.Every(time.Second/5), 5)

	for i := 1; i <= 15; i++ {
		if limiter.Allow() {
			go processTask(i)
		} else {
			fmt.Printf("Task %d dropped: rate limit exceeded\n", i)
		}
	}

	time.Sleep(2 * time.Second)
}
