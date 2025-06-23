// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
			log.Info("Exit the func successfully.")
			return nil
		} else if !IsNeedWaitError(err) && !IsConnectionRefusedError(err) && !IsUnexpectedEOFError(err) {
			log.Infof("Still need to wait for func, err:%s\n", err.Error())
			return err
		}

		if i >= (attempts - 1) {
			break
		}

		time.Sleep(sleep)

		log.Infoln("Retrying after error:", err)
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
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
		} else if !IsNeedWaitError(err) && !IsConnectionRefusedError(err) && !IsUnexpectedEOFError(err) {
			log.Warnf("Unexpected err %v", err)
			return err
		} else {
			log.Infof("Still need to wait for func, err:%s\n", err.Error())
		}

		delta := time.Since(start)
		if delta > duration {
			return fmt.Errorf("after %d attempts (during %s), last error: %s", i, delta, err)
		}

		time.Sleep(sleep)

		log.Infoln("Retrying after error:", err)
	}
}
