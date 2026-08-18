package dto

import (
	"testing"
	"time"
)

func TestParseTimeout(t *testing.T) {
	for _, test := range []struct {
		name    string
		value   string
		want    time.Duration
		wantErr bool
	}{
		{name: "default", want: 20 * time.Second},
		{name: "milliseconds", value: "1500ms", want: 1500 * time.Millisecond},
		{name: "seconds", value: "60s", want: 60 * time.Second},
		{name: "below minimum", value: "99ms", wantErr: true},
		{name: "decimal", value: "1.5s", wantErr: true},
		{name: "over maximum", value: "61s", wantErr: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := ParseTimeout(test.value, 20*time.Second, 60*time.Second)
			if (err != nil) != test.wantErr || got != test.want {
				t.Fatalf("ParseTimeout() = %s, %v; want %s, error=%v", got, err, test.want, test.wantErr)
			}
		})
	}
}
