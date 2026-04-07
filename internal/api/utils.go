package api

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/nats-io/nats.go/jetstream"
)

// natsErrMsg extracts a clean error message from NATS JetStream errors.
// "nats: API error: code=400 err_code=10065 description=subjects overlap..."
// becomes "subjects overlap with an existing stream (10065)".
func natsErrMsg(err error) string {
	var jsErr jetstream.JetStreamError
	if ok := errors.As(err, &jsErr); ok {
		if apiErr := jsErr.APIError(); apiErr != nil && apiErr.Description != "" {
			return fmt.Sprintf("%s (%d)", apiErr.Description, apiErr.ErrorCode)
		}
	}
	return err.Error()
}

// ParseByteSize parses human-readable byte sizes like "100MB", "1GB", "256KB".
// Suffixes must be checked longest-first to avoid "B" matching before "MB".
func ParseByteSize(s string) (int64, error) {
	if s == "" {
		return 0, nil
	}
	s = strings.TrimSpace(s)
	s = strings.ToUpper(s)

	// Ordered longest-suffix-first to prevent "B" matching "MB"
	suffixes := []struct {
		suffix string
		mult   int64
	}{
		{"TB", 1024 * 1024 * 1024 * 1024},
		{"GB", 1024 * 1024 * 1024},
		{"MB", 1024 * 1024},
		{"KB", 1024},
		{"B", 1},
	}

	for _, entry := range suffixes {
		if strings.HasSuffix(s, entry.suffix) {
			numStr := strings.TrimSpace(strings.TrimSuffix(s, entry.suffix))
			if numStr == "" {
				return 0, fmt.Errorf("invalid size: %s", s)
			}
			val, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid size: %s", s)
			}
			return int64(val * float64(entry.mult)), nil
		}
	}

	// Try plain number (bytes)
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size: %s", s)
	}
	return val, nil
}
