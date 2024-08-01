package ratelimiter

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type KV struct {
	Key   string `json:"key" gorm:"primaryKey;not null;"`
	Value string `json:"value" gorm:"not null;"`
}

func DriverGORM(db *gorm.DB) Driver {
	return DriverFunc(func(ctx context.Context, req *ReserveRequest) (*Reservation, error) {
		if req.Key == "" || req.Now.IsZero() || req.DurationPerToken <= 0 || req.Burst <= 0 || req.Tokens <= 0 || req.Tokens > req.Burst {
			return nil, errors.Errorf("ratelimiter: invalid parameters: %v", req)
		}

		select {
		case <-ctx.Done():
			return nil, errors.Wrap(ctx.Err(), "ratelimiter: context done")
		default:
		}

		resetValue := req.Now.Add(-time.Duration(req.Burst) * req.DurationPerToken)

		var baseTime time.Time
		var timeToAct time.Time
		var ok bool

		db := db.WithContext(ctx)
		err := db.Transaction(func(tx *gorm.DB) error {
			var kv KV

			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&kv, "key = ?", req.Key).Error; err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}

				baseTime = resetValue
				kv = KV{Key: req.Key, Value: fmt.Sprintf("%d", baseTime.UnixNano())}
				if err := tx.Create(&kv).Error; err != nil {
					return err
				}
			} else {
				baseTimeUnix, err := strconv.ParseInt(kv.Value, 10, 64)
				if err != nil {
					return errors.Wrap(err, "ratelimiter: failed to parse base time")
				}
				baseTime = time.Unix(0, baseTimeUnix)

				if baseTime.Before(resetValue) {
					baseTime = resetValue
				}
			}

			tokensDuration := req.DurationPerToken * time.Duration(req.Tokens)
			timeToAct = baseTime.Add(tokensDuration)

			if timeToAct.After(req.Now.Add(req.MaxFutureReserve)) {
				ok = false
				return nil
			}

			kv.Value = fmt.Sprintf("%d", timeToAct.UnixNano())
			if err := tx.Save(&kv).Error; err != nil {
				return errors.Wrap(err, "ratelimiter: failed to save time to act")
			}
			ok = true
			return nil
		})
		if err != nil {
			return nil, err
		}

		return &Reservation{
			ReserveRequest: req,
			OK:             ok,
			TimeToAct:      timeToAct,
		}, nil
	})
}
