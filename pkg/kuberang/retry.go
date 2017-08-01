package kuberang

import "time"

func retry(times int, f func() bool) bool {
	attempt := 0
	for attempt < times {
		if ok := f(); ok {
			return true
		}
		time.Sleep(1 * time.Second)
		attempt++
	}
	return false
}

func retryWithBackoff(times uint, f func() bool) bool {
	var attempt uint
	for attempt < times {
		if ok := f(); ok {
			return true
		}
		time.Sleep((1 << attempt) * time.Second)
		attempt++
	}
	return false
}
