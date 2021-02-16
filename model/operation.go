package model

import "time"

type Operation struct {
	name      string
	State     string
	StartTime time.Time
	Duration  int
}

func (o *Operation) Name() *string {
	return &o.name
}

func NewOperation(name string, duration int) *Operation {
	return &Operation{
		name:      name,
		State:     "in_progress",
		StartTime: time.Now(),
		Duration:  duration,
	}
}
