package util

import (
	"fmt"
	"strings"
	"time"
)

func RelativeDuration(d time.Duration) string {
	times := []string{}

	durationFormats := []struct {
		nameSingular string
		namePlural   string
		duration     time.Duration
	}{
		{"dia", "dias", 24 * time.Hour},
		{"hora", "horas", time.Hour},
		{"minuto", "minutos", time.Minute},
		{"segundo", "segundos", time.Second},
	}

	for _, format := range durationFormats {
		if len(times) == 2 {
			break
		}
		div := d / format.duration
		if div == 0 {
			continue
		}
		d -= div * format.duration

		s := fmt.Sprint(int(div)) + " "
		if div == 1 {
			s += format.nameSingular
		} else {
			s += format.namePlural
		}
		times = append(times, s)
	}

	return strings.Join(times, " e ")
}

func Debounce(d time.Duration, fn func()) func() {
	end := time.Now().Add(d)
	return func() {
		for time.Now().Before(end) {
			end = time.Now().Add(d)
			time.Sleep(d)
		}
		fn()
	}
}
