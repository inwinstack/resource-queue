package hypervisor

import (
	"sync"
)

type Hypervisor struct {
	Name  string `json:"Hypervisor Hostname"`
	IP    string `json:"Host IP"`
	State string `json:"State"`
	Type  string `json:"Hypervisor Type"`
}

type HypervisorDetail struct {
	Name       string `json:"hypervisor_hostname"`
	VCPU       int    `json:"vcpus"`
	VCPUUsed   int    `json:"vcpus_used"`
	FreeMemory int    `json:"free_ram_mb"`
	FreeDisk   int    `json:"free_disk_gb"`
}

type HypervisorResource struct {
	resource []*HypervisorDetail
	lock     sync.RWMutex
}

type Instance struct {
	VCPU   int
	Memory int
	Disk   int
}
