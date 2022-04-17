package utils

import (
	"time"
)

func RetryLoop(attempts int, sleep time.Duration, f func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		if i > 0 {
			time.Sleep(sleep)
		}
		err = f()
		if err == nil {
			return nil
		}
	}
	return err
}
