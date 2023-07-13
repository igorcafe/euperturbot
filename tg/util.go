package tg

import (
	"log"
	"math"
	"time"
)

type Retry struct {
	maxAttempts int
	delayFactor int
	delay       time.Duration
}

func (r *Retry) Do(fn func() error) error {
	var err error
	for i := 0; i < r.maxAttempts; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		multiplier := int(math.Pow(float64(r.delayFactor), float64(i)))
		t := r.delay * time.Duration(multiplier)
		time.Sleep(t)
	}
	if err != nil {
		log.Print(err)
	}
	return err
}

func ToPtr[T any](v T) *T {
	return &v
}
