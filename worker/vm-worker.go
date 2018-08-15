// Package go provides ...
package worker

import (
	"github.com/antonholmquist/jason"
	"github.com/kjelly/resource-queue/hypervisor"
	"github.com/kjelly/resource-queue/queue"
	log "github.com/sirupsen/logrus"
	"time"
)

type VMWorker struct {
	Worker
	q       *queue.Queue
	r       hypervisor.HypervisorResource
	running bool
}

func (v *VMWorker) CheckJob() error {
	job := v.q.GetOneJob("vm")
	if job == nil {
		log.Debug("No job found. Skip.")
		return nil
	}
	log.Info("Check resource for job, (%v)", job)
	decoder, err := jason.NewObjectFromBytes([]byte(job.Data))
	if err != nil {
		v.q.SetJobStatus(job, "error")
	}
	var instance hypervisor.Instance

	value, err := decoder.GetInt64("VCPU")
	if err != nil {
		log.Error("Error to get int from VCPU. (%v)", job)
	}
	instance.VCPU = int(value)

	value, err = decoder.GetInt64("Memory")
	if err != nil {
		log.Error("Error to get int from Memory. (%v)", job)
	}
	instance.Memory = int(value)

	value, err = decoder.GetInt64("Disk")
	if err != nil {
		log.Error("Error to get int from Disk. (%v)", job)
	}
	instance.Disk = int(value)

	if v.r.Check(instance) {
		v.q.SetJobStatus(job, "done")
		log.Info("Finish job\n")

	} else {
		log.Info("Resource not enough")
	}

	time.Sleep(1 * time.Second)
	return nil
}

func InitVMWorker(q *queue.Queue) *VMWorker {
	ret := new(VMWorker)
	ret.q = q
	ret.name = "VM worker"
	ret.stopped = false
	ret.stopChan = make(chan int)
	ret.SetIntervalt(time.Second)
	ret.r.Update()
	ret.SetExecute(func() {
		ret.CheckJob()
	})
	return ret
}
