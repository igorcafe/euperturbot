package util

import (
	"testing"
	"time"
)

func Test_RelativeDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{time.Second, "1 segundo"},
		{59 * time.Second, "59 segundos"},
		{65 * time.Second, "1 minuto e 5 segundos"},
		{62 * time.Minute, "1 hora e 2 minutos"},
		{24 * time.Hour, "1 dia"},
	}

	for _, tt := range tests {
		got := RelativeDuration(tt.d)
		if tt.want != got {
			t.Errorf("want: '%s', got: '%s", tt.want, got)
		}
	}
}
