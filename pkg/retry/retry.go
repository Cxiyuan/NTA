package retry

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
)

type Config struct {
	MaxAttempts int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

func DefaultConfig() Config {
	return Config{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
	}
}

func Do(ctx context.Context, config Config, logger *logrus.Logger, fn func() error) error {
	var lastErr error
	delay := config.InitialDelay

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		if attempt == config.MaxAttempts {
			break
		}

		if logger != nil {
			logger.Warnf("Attempt %d/%d failed: %v. Retrying in %v...", 
				attempt, config.MaxAttempts, err, delay)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}

		delay = time.Duration(float64(delay) * config.Multiplier)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}

	return errors.New("max retry attempts exceeded: " + lastErr.Error())
}

func DoWithResult[T any](ctx context.Context, config Config, logger *logrus.Logger, fn func() (T, error)) (T, error) {
	var result T
	var lastErr error
	delay := config.InitialDelay

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		res, err := fn()
		if err == nil {
			return res, nil
		}

		lastErr = err

		if attempt == config.MaxAttempts {
			break
		}

		if logger != nil {
			logger.Warnf("Attempt %d/%d failed: %v. Retrying in %v...", 
				attempt, config.MaxAttempts, err, delay)
		}

		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-time.After(delay):
		}

		delay = time.Duration(float64(delay) * config.Multiplier)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}

	return result, errors.New("max retry attempts exceeded: " + lastErr.Error())
}
