package model

import "time"

type Operation struct {
	name       string
	state      string
	startTime  time.Time
	duration   float64
	shouldFail bool
}

const (
	PROGRESSING = "in_progress"
	SUCCEEDED   = "succeeded"
	FAILED      = "failed"
)

func (operation *Operation) Name() *string {
	return &operation.name
}

func (operation *Operation) State() *string {
	if operation.state == "in_progress" {
		if time.Now().Sub(operation.startTime).Seconds() >= operation.duration {
			if operation.shouldFail {
				operation.state = FAILED
			} else {
				operation.state = SUCCEEDED
			}
		}
	}
	return &operation.state
}

func NewOperation(name string, duration float64) *Operation {
	return &Operation{
		name:      name,
		state:     PROGRESSING,
		startTime: time.Now(),
		duration:  duration,
	}
}
