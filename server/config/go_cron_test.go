package config

import "testing"

func TestCronOptIn(t *testing.T) {
	for _, tc := range []struct {
		value        string
		enabled, bad bool
	}{{"", false, false}, {"false", false, false}, {"true", true, false}, {"oops", false, true}} {
		t.Run(tc.value, func(t *testing.T) {
			t.Setenv("CRON_ENABLED", tc.value)
			got, err := CronEnabled()
			if got != tc.enabled || (err != nil) != tc.bad {
				t.Fatalf("got=%v err=%v", got, err)
			}
		})
	}
}
