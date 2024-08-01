package ratelimiter

import (
	"context"
	"time"
)

type AllowRequest struct {
	Key              string
	DurationPerToken time.Duration
	Burst            int64
	Now              time.Time
	Tokens           int64
}

type ReserveRequest struct {
	Key              string
	DurationPerToken time.Duration
	Burst            int64
	Now              time.Time
	Tokens           int64
	MaxFutureReserve time.Duration
}

type Reservation struct {
	*ReserveRequest
	OK        bool
	TimeToAct time.Time
}

func (r *Reservation) DelayFrom(t time.Time) time.Duration {
	if !r.OK {
		panic("ratelimiter: cannot get delay from non-OK reservation")
	}

	delay := r.TimeToAct.Sub(t)
	if delay < 0 {
		return 0
	}
	return delay
}

func (r *Reservation) Delay() time.Duration {
	return r.DelayFrom(time.Now())
}

func (r *Reservation) RetryAfterFrom(t time.Time) time.Duration {
	if r.OK {
		panic("ratelimiter: cannot get retry after from OK reservation")
	}

	delay := r.TimeToAct.Sub(t) - r.MaxFutureReserve
	if delay < 0 {
		return 0
	}
	return delay
}

func (r *Reservation) RetryAfter() time.Duration {
	return r.RetryAfterFrom(time.Now())
}

type Driver interface {
	Reserve(ctx context.Context, req *ReserveRequest) (*Reservation, error)
}

type DriverFunc func(ctx context.Context, req *ReserveRequest) (*Reservation, error)

func (f DriverFunc) Reserve(ctx context.Context, req *ReserveRequest) (*Reservation, error) {
	return f(ctx, req)
}

type RateLimiter struct {
	driver Driver
}

func New(driver Driver) *RateLimiter {
	return &RateLimiter{driver: driver}
}

func (lim *RateLimiter) Allow(ctx context.Context, req *AllowRequest) (bool, error) {
	reserveReq := &ReserveRequest{
		Key:              req.Key,
		DurationPerToken: req.DurationPerToken,
		Burst:            req.Burst,
		Now:              req.Now,
		Tokens:           req.Tokens,
		MaxFutureReserve: 0,
	}

	r, err := lim.Reserve(ctx, reserveReq)
	if err != nil {
		return false, err
	}
	return r.OK, nil
}

func (lim *RateLimiter) Reserve(ctx context.Context, req *ReserveRequest) (*Reservation, error) {
	return lim.driver.Reserve(ctx, req)
}
