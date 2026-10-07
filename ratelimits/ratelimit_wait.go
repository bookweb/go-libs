package ratelimits

import (
	"context"
	"fmt"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestRateLimitWait(t *testing.T) {
	limiter := rate.NewLimiter(rate.Every(time.Second/5), 5)
	ctx := context.Background()

	for i := 1; i <= 15; i++ {
		if err := limiter.Wait(ctx); err != nil {
			fmt.Println("Error:", err)
			return
		}

		go func(id int) {
			fmt.Printf("Executing task %d at %s\n", id, time.Now().Format("15:04:05.000"))
		}(i)
	}

	time.Sleep(2 * time.Second)
}

func processTask(id int) {
	fmt.Printf("Executing task %d at %s\n", id, time.Now().Format("15:04:05.000"))
}
