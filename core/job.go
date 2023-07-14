package core

import (
	"errors"
	"sync"
)

var ErrJobLimitReached = errors.New("Job Limit Has Been Reached")

type JobLimiter struct {
	counter     uint
	counterLock sync.Mutex
	max         uint
}

func NewJobLimiter(max uint) *JobLimiter {
	return &JobLimiter{
		counter:     0,
		counterLock: sync.Mutex{},
		max:         max,
	}
}

func (jc *JobLimiter) Jobs() uint {
	jc.counterLock.Lock()
	defer jc.counterLock.Unlock()
	return jc.counter
}

func (jc *JobLimiter) AddJob() error {
	jc.counterLock.Lock()
	defer jc.counterLock.Unlock()
	if jc.max != 0 && jc.counter == jc.max {
		return ErrJobLimitReached
	}
	jc.counter++
	return nil
}

func (jc *JobLimiter) RemoveJob() {
	jc.counterLock.Lock()
	defer jc.counterLock.Unlock()
	jc.counter--
}
