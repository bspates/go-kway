package main

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Task struct {
	TaskId      int
	Created     time.Time
	Modified    pq.NullTime
	Name        string
	Pipe        []string
	Status      string
	Attempts    int
	MaxAttempts int
	Backoff     time.Time
	Payload     []byte
	Result      interface{}
}

type Registry map[string]Action

type Worker struct {
	workerId int
	in       chan *WorkerAction
	out      chan *WorkerAction
}

type Kway struct {
	pollInterval time.Duration
	inFlight     int
	paused       bool
	workers      []*Worker
	registry     Registry
}

func (k *Kway) pickWorker() *Worker {
	return k.workers[0]
}

func (w *Worker) start() {
	for wa := range w.in {
		res, err := wa.action.execute(wa.task.Payload)
		if err != nil {
			wa.task.Status = STATUS_ERROR
			wa.task.Result = err
			fmt.Println("error: ", err)
		} else {
			wa.task.Result = res
			wa.task.Status = STATUS_COMPLETED
		}

		w.out <- wa
	}
}

func (k *Kway) poll(db *sqlx.DB, concurrency int) <-chan bool {
	k.workers = []*Worker{}
	in := make(chan *WorkerAction, concurrency)
	out := make(chan *WorkerAction, concurrency)
	for w := 0; w < concurrency; w++ {
		worker := Worker{w, in, out}
		go worker.start()
		k.workers = append(k.workers, &worker)
	}

	ticker := time.NewTicker(k.pollInterval * time.Millisecond)
	go func() {
		for t := range ticker.C {
			fmt.Println("tick ", t)

			var toDo []*Task
			var done []*Task

			// Update completed/errored tasks in db
		Loop:
			for {
				select {
				case wa := <-out:
					done = append(done, wa.task)
					k.inFlight--
				default:
					break Loop
				}
			}
			if len(done) > 0 {
				err := result(db, done)
				if err != nil {
					panic(err)
				}
			}

			// Pull new tasks
			pullNum := concurrency - k.inFlight
			if pullNum <= 0 {
				continue
			}
			toDo, err := dequeue(db, pullNum)
			if err != nil {
				panic(err)
			}
			for _, task := range toDo {
				worker := k.pickWorker()
				k.inFlight++
				a, ok := k.registry[task.Name]
				if !ok {
					continue
				}
				worker.in <- &WorkerAction{action: a, task: task}
			}
		}
	}()
	fin := make(chan bool)
	return fin
}
