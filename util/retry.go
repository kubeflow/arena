package util

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"time"
)

func Retry(attempts int, sleep time.Duration, callback func() error) (err error) {
	for i := 0; ; i++ {
		err = callback()
		if err == nil {
			log.Infof("Exit the func %v successfully.", callback)
			return nil
		} else if !(IsNeedWaitError(err) ||
			IsConnectionRefusedError(err) ||
			IsUnexpectedEOFError(err)) {
			log.Infof("Still need to wait for func, err:%s\n", err.Error())
			return err
		}

		if i >= (attempts - 1) {
			break
		}

		time.Sleep(sleep)

		log.Infoln("Retrying after error:", err)
	}
	return fmt.Errorf("After %d attempts, last error: %s", attempts, err)
}

func RetryDuring(duration time.Duration, sleep time.Duration, callback func() error) (err error) {
	start := time.Now()
	i := 0
	for {
		i++
		err = callback()
		if err == nil {
			log.Infof("Exit the func successfully.")
			return nil
		} else if !(IsNeedWaitError(err) ||
			IsConnectionRefusedError(err) ||
			IsUnexpectedEOFError(err)) {
			log.Warnf("Unexpected err %v", err)
			return err
		} else {
			log.Infof("Still need to wait for func, err:%s\n", err.Error())
		}

		delta := time.Now().Sub(start)
		if delta > duration {
			return fmt.Errorf("After %d attempts (during %s), last error: %s", i, delta, err)
		}

		time.Sleep(sleep)

		log.Infoln("Retrying after error:", err)
	}
}
