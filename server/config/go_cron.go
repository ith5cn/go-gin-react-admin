package config

import (
	"fmt"
	"os"
	"strconv"
)

// CronEnabled is opt-in so additional HTTP replicas never start a scheduler by default.
func CronEnabled() (bool, error) {
	value := os.Getenv("CRON_ENABLED")
	if value == "" {
		return false, nil
	}
	enabled, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("CRON_ENABLED must be a boolean: %w", err)
	}
	return enabled, nil
}
