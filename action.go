package main

type Action interface {
	execute(interface{}) (interface{}, error)
}

type WorkerAction struct {
	action Action
	task   *Task
}
