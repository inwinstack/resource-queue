// Package go provides ...
package worker

import (
	"fmt"
	"time"
)

type IWorker interface {
	IsStop() bool
	Stop() error
	Run()
}

type Worker struct {
	stopChan chan int
	stopped  bool
	execute  func()
	interval time.Duration
}

func (v *Worker) SetIntervalt(t time.Duration) {
	v.interval = t
}

func (v *Worker) IsStop() bool {
	return v.stopped
}

func (v *Worker) Stop() error {
	v.stopped = true
	<-v.stopChan
	return nil
}

func (v *Worker) Run() {
	fmt.Printf("Worker is running\n")
	for !v.IsStop() {
		v.execute()
		time.Sleep(time.Second)
	}
	fmt.Printf("Worker is stopped\n")
	v.stopChan <- 1
}

func (v *Worker) SetExecute(f func()) error {
	v.execute = f
	return nil
}
