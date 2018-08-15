// Package go provides ...
package worker

import (
	log "github.com/sirupsen/logrus"
	"time"
)

type IWorker interface {
	IsStop() bool
	Stop() error
	Run()
}

type Worker struct {
	name     string
	stopChan chan int
	stopped  bool
	running  bool
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
	if v.running {
		<-v.stopChan
		v.running = false
	}
	return nil
}

func (v *Worker) Run() {
	log.Debugf("Worker, %s, is running\n", v.name)
	v.running = true
	for !v.IsStop() {
		v.execute()
		time.Sleep(time.Second)
	}
	log.Debugf("Worker, %s, is stopped\n", v.name)
	v.stopChan <- 1
}

func (v *Worker) SetExecute(f func()) error {
	v.execute = f
	return nil
}
