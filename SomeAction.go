package main

import "encoding/json"

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

func (s SomeAction) execute(p interface{}) (interface{}, error) {
	pa := p.([]byte)
	json.Unmarshal(pa, &s)
	s.Result = someActionResult(s.Payload)
	return s.Result, nil
}
