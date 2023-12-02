package util

import (
	"math"
	"time"
)

type Retry struct {
	MaxAttempts int
	DelayFactor int
	Delay       time.Duration
}

func (r *Retry) Do(fn func() error) error {
	var err error

	for i := 0; i < r.MaxAttempts; i++ {
		err = fn()
		if err == nil {
			return nil
		}

		multiplier := int(math.Pow(float64(r.DelayFactor), float64(i)))
		t := r.Delay * time.Duration(multiplier)
		time.Sleep(t)
	}

	return err
}
