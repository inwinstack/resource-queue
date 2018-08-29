package hypervisor

import (
	"encoding/json"
	"fmt"
	"github.com/antonholmquist/jason"
	"os/exec"
)

func GetHypervisorList() ([]Hypervisor, error) {

	var h []Hypervisor
	cmd := exec.Command("openstack", "hypervisor", "list", "-f", "json")
	o, err := cmd.Output()
	if err != nil {
		err = fmt.Errorf("Run openstack failed. "+
			"Make sure you can run openstack command without error. (%s)", err.Error())
		return nil, err
	}
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

func GetAllHypervisorDetail(h []Hypervisor) (map[string]*HypervisorDetail, error) {
	ch := make(chan *HypervisorDetail)
	var d *HypervisorDetail
	threadCount := 0
	for _, v := range h {
		a := v
		if a.State == "up" {
			go func() {
				d, _ = GetHypervisorDetail(a.Name)
				ch <- d
			}()
			threadCount++
		}
	}
	ret := make(map[string]*HypervisorDetail)
	for i := 0; i < threadCount; i++ {
		d = <-ch
		if d != nil {
			ret[d.Name] = d
		}
	}
	return ret, nil
}

func GetAggregateList() ([]string, error) {
	cmd := exec.Command("openstack", "aggregate", "list", "-f", "json")
	o, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var ret []string

	decoder, err := jason.NewObjectFromBytes([]byte("{\"data\":" + string(o) + "}"))
	if err != nil {
		return nil, err
	}
	arr, err := decoder.GetObjectArray("data")
	if err != nil {
		return nil, err
	}
	for _, v := range arr {
		name, _ := v.GetString("Name")
		ret = append(ret, name)
	}
	return ret, nil
}

func GetAggregateHost(name string) ([]string, error) {
	cmd := exec.Command("openstack", "aggregate", "show", name, "-f", "json")
	o, err := cmd.Output()
	if err != nil {
		panic(err)
	}
	var ret []string
	decoder, err := jason.NewObjectFromBytes(o)
	if err != nil {
		panic(err)
	}
	ret, err = decoder.GetStringArray("hosts")
	if err != nil {
		panic(err)
	}

	return ret, nil
}

func GetAggregateHostMap() (map[string][]*HypervisorDetail, error) {
	aggregateList, err := GetAggregateList()
	hypervisorList, err := GetHypervisorList()
	if err != nil {
		panic(err)
	}
	hypervisorMap, err := GetAllHypervisorDetail(hypervisorList)
	ret := make(map[string][]*HypervisorDetail)
	for _, v := range aggregateList {
		hostNames, _ := GetAggregateHost(v)
		var hosts []*HypervisorDetail
		for _, h := range hostNames {
			hosts = append(hosts, hypervisorMap[h])
		}
		ret[v] = hosts

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
