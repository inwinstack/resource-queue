package hypervisor

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

func GetHypervisorList() ([]Hypervisor, error) {

	var h []Hypervisor
	cmd := exec.Command("openstack", "hypervisor", "list", "-f", "json")
	o, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	fmt.Printf("%s\n", o)
	err = json.Unmarshal(o, &h)
	if err != nil {
		return nil, err
	}
	return h, nil
}

func GetHypervisorDetail(name string) (*HypervisorDetail, error) {
	var h *HypervisorDetail
	cmd := exec.Command("openstack", "hypervisor", "show", name, "-f", "json")
	o, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(o, &h)
	if err != nil {
		return nil, err
	}
	return h, nil
}

func GetAllHypervisorDetail(h []Hypervisor) ([]*HypervisorDetail, error) {
	ch := make(chan *HypervisorDetail)
	var d *HypervisorDetail
	var err error
	threadCount := 0
	for _, v := range h {
		a := v
		if a.State == "up" {
			go func() {
				d, err = GetHypervisorDetail(a.Name)
				ch <- d
			}()
			threadCount++
		}
	}
	var ret []*HypervisorDetail
	for i := 0; i < threadCount; i++ {
		d = <-ch
		if d != nil {
			ret = append(ret, d)
		}
	}
	return ret, nil
}

func (r *HypervisorResource) Update() {
	r.lock.Lock()
	hypervisorList, err := GetHypervisorList()
	if err != nil {
		panic(err)
	}
	ds, _ := GetAllHypervisorDetail(hypervisorList)
	r.resource = ds
	r.lock.Unlock()
}

func (r *HypervisorResource) Show() {
	for _, v := range r.resource {
		fmt.Printf("%v\n", v)
	}
}
func (r *HypervisorResource) Check(ins Instance) bool {
	for _, h := range r.resource {
		freeVCPU := h.VCPU - h.VCPUUsed
		if ins.VCPU < freeVCPU && ins.Memory < h.FreeMemory && ins.Disk < h.FreeDisk {
			return true
		}
	}
	return false
}
