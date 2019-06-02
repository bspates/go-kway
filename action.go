package main

import "encoding/json"

type Action interface {
	execute() error
	payload([]byte) error
	result() interface{}
}

type WorkerAction struct {
	action Action
	task   *Task
}

type someActionPayload struct {
	One   int
	Two   bool
	Three string
}

type someActionResult struct {
	One   int
	Two   bool
	Three string
}

type SomeAction struct {
	Payload someActionPayload
	Result  someActionResult
}

func (s SomeAction) execute() error {
	s.Result = someActionResult(s.Payload)
	return nil
}

func (s SomeAction) payload(p []byte) error {
	return json.Unmarshal(p, &s.Payload)
}

func (s SomeAction) result() interface{} {
	return s.Result
}
