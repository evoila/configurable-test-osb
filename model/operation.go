package model

import "time"

type Operation struct {
	name                      string
	state                     string
	startTime                 time.Time
	duration                  float64
	shouldFail                bool
	updateRepeatableAfterFail *bool
	updateRepeatable          *bool
	instanceUsableAfterFail   *bool
	instanceUsable            *bool
	async                     *bool
}

func (operation *Operation) SupposedToFail() bool {
	return operation.shouldFail
}

func (operation *Operation) Async() *bool {
	return operation.async
}

const (
	PROGRESSING = "in_progress"
	SUCCEEDED   = "succeeded"
	FAILED      = "failed"
)

func (operation *Operation) UpdateRepeatable() *bool {
	return operation.updateRepeatable
}

func (operation *Operation) InstanceUsable() *bool {
	return operation.instanceUsable
}

func (operation *Operation) Name() *string {
	return &operation.name
}

func (operation *Operation) State() *string {
	if operation.state == "in_progress" {
		if time.Now().Sub(operation.startTime).Seconds() >= operation.duration {
			if operation.shouldFail {
				operation.state = FAILED
				operation.updateRepeatable = operation.updateRepeatableAfterFail
				operation.instanceUsable = operation.instanceUsableAfterFail
			} else {
				operation.state = SUCCEEDED
			}
		}
	}
	return &operation.state
}

func NewOperation(name string, duration float64, shouldFail bool, updateRepeatableAfterFail *bool, instanceUsableAfterFail *bool, async bool) *Operation {
	return &Operation{
		name:                      name,
		state:                     PROGRESSING,
		startTime:                 time.Now(),
		duration:                  duration,
		shouldFail:                shouldFail,
		updateRepeatableAfterFail: updateRepeatableAfterFail,
		instanceUsableAfterFail:   instanceUsableAfterFail,
		async:                     &async,
	}
}
