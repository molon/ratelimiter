# ratelimiter

**Move to https://github.com/theplant/ratelimiter**

Currently only supports GORM as the driver, with plans to integrate Redis in the future.

```
package ratelimiter_test

import (
	"context"
	"fmt"
	"time"

	"github.com/molon/ratelimiter"
)

func runExample(limiter *ratelimiter.RateLimiter) {
	// every 10 min , burst 5
	durationPerToken := 10 * time.Minute
	burst := int64(5)
	now := time.Now()
	key := "test_allow"

	ctx := context.Background()

	try := func(delta time.Duration) bool {
		reserveReq := &ratelimiter.ReserveRequest{
			Key:              key,
			DurationPerToken: durationPerToken,
			Burst:            burst,
			Now:              now.Add(delta),
			Tokens:           1,
			MaxFutureReserve: 0,
		}
		r, err := limiter.Reserve(ctx, reserveReq)
		if err != nil {
			panic(err)
		}

		if r.OK {
			fmt.Printf("%v: allowed: %t\n", delta, r.OK)
			return true
		}

		fmt.Printf("%v: allowed: %t , you can retry after %v\n", delta, false, r.RetryAfterFrom(reserveReq.Now))
		return false
	}

	for i := 0; i < int(25); i++ {
		delta := time.Duration(i) * time.Minute
		try(delta)
	}

	fmt.Printf("--- Sleep 20 minutes ---\n")

	for i := 45; i < int(55); i++ {
		delta := time.Duration(i) * time.Minute
		try(delta)
	}

	fmt.Printf("--- Sleep 100 minutes ---\n")

	for i := 155; i < int(165); i++ {
		delta := time.Duration(i) * time.Minute
		try(delta)
	}
}

func ExampleDriverGORM() {
	resetDB()

	limiter := ratelimiter.New(
		ratelimiter.DriverGORM(db),
	)
	runExample(limiter)
	// Output:
	// 0s: allowed: true
	// 1m0s: allowed: true
	// 2m0s: allowed: true
	// 3m0s: allowed: true
	// 4m0s: allowed: true
	// 5m0s: allowed: false , you can retry after 5m0s
	// 6m0s: allowed: false , you can retry after 4m0s
	// 7m0s: allowed: false , you can retry after 3m0s
	// 8m0s: allowed: false , you can retry after 2m0s
	// 9m0s: allowed: false , you can retry after 1m0s
	// 10m0s: allowed: true
	// 11m0s: allowed: false , you can retry after 9m0s
	// 12m0s: allowed: false , you can retry after 8m0s
	// 13m0s: allowed: false , you can retry after 7m0s
	// 14m0s: allowed: false , you can retry after 6m0s
	// 15m0s: allowed: false , you can retry after 5m0s
	// 16m0s: allowed: false , you can retry after 4m0s
	// 17m0s: allowed: false , you can retry after 3m0s
	// 18m0s: allowed: false , you can retry after 2m0s
	// 19m0s: allowed: false , you can retry after 1m0s
	// 20m0s: allowed: true
	// 21m0s: allowed: false , you can retry after 9m0s
	// 22m0s: allowed: false , you can retry after 8m0s
	// 23m0s: allowed: false , you can retry after 7m0s
	// 24m0s: allowed: false , you can retry after 6m0s
	// --- Sleep 20 minutes ---
	// 45m0s: allowed: true
	// 46m0s: allowed: true
	// 47m0s: allowed: false , you can retry after 3m0s
	// 48m0s: allowed: false , you can retry after 2m0s
	// 49m0s: allowed: false , you can retry after 1m0s
	// 50m0s: allowed: true
	// 51m0s: allowed: false , you can retry after 9m0s
	// 52m0s: allowed: false , you can retry after 8m0s
	// 53m0s: allowed: false , you can retry after 7m0s
	// 54m0s: allowed: false , you can retry after 6m0s
	// --- Sleep 100 minutes ---
	// 2h35m0s: allowed: true
	// 2h36m0s: allowed: true
	// 2h37m0s: allowed: true
	// 2h38m0s: allowed: true
	// 2h39m0s: allowed: true
	// 2h40m0s: allowed: false , you can retry after 5m0s
	// 2h41m0s: allowed: false , you can retry after 4m0s
	// 2h42m0s: allowed: false , you can retry after 3m0s
	// 2h43m0s: allowed: false , you can retry after 2m0s
	// 2h44m0s: allowed: false , you can retry after 1m0s
}

```
