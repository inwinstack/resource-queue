// Package go provides ...
package worker

import (
	"fmt"
	"github.com/antonholmquist/jason"
	"github.com/kjelly/resource-queue/hypervisor"
	"github.com/kjelly/resource-queue/queue"
	"time"
)

type VMWorker struct {
	Worker
	q *queue.Queue
	r hypervisor.HypervisorResource
}

func (v *VMWorker) CheckJob() error {
	job := v.q.GetOneJob("vm")
	if job == nil {
		return nil
	}
	decoder, err := jason.NewObjectFromBytes([]byte(job.Data))
	if err != nil {
		v.q.SetJobStatus(job, "error")
	}
	var instance hypervisor.Instance

	value, err := decoder.GetInt64("VCPU")
	if err != nil {
		fmt.Printf("Error to get int")
	}
	instance.VCPU = int(value)

	value, err = decoder.GetInt64("Memory")
	if err != nil {
		fmt.Printf("Error to get int\n")
	}
	instance.Memory = int(value)

	value, err = decoder.GetInt64("Disk")
	if err != nil {
		fmt.Printf("Error to get int\n")
	}
	instance.Disk = int(value)

	if v.r.Check(instance) {
		v.q.SetJobStatus(job, "done")
		fmt.Printf("Finish job\n")

	} else {
		fmt.Printf("Resource not enough\n")
	}

	fmt.Printf("%v\n", job)
	time.Sleep(1 * time.Second)
	return nil
}

func InitVMWorker(q *queue.Queue) *VMWorker {
	ret := new(VMWorker)
	ret.q = q
	ret.stopped = false
	ret.stopChan = make(chan int)
	ret.SetIntervalt(time.Second)
	ret.r.Update()
	ret.SetExecute(func() {
		ret.CheckJob()
	})
	return ret
}
