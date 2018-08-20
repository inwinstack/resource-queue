package httpHandler

import (
	"fmt"
	"net/http"
	_ "strconv"

	"encoding/json"
	"github.com/antonholmquist/jason"
	"github.com/gorilla/mux"
	"github.com/kjelly/resource-queue/hypervisor"
	"github.com/kjelly/resource-queue/queue"
	log "github.com/sirupsen/logrus"
)

type Handler interface {
	Check(w http.ResponseWriter, r *http.Request)
	GetJobs(w http.ResponseWriter, r *http.Request)
	GetJob(w http.ResponseWriter, r *http.Request)
	UpdateProperty(w http.ResponseWriter, r *http.Request)
	AddJob(w http.ResponseWriter, r *http.Request)
	DeleteJob(w http.ResponseWriter, r *http.Request)
	Test(w http.ResponseWriter, r *http.Request)
	GetQueue() *queue.Queue
	Kind() string
}

type requestBody struct {
	RequestID string `json:"request_id"`
	OwnerID   string `json:"owner_id"`
}

type VMHandler struct {
	resource hypervisor.HypervisorResource
	q        *queue.Queue
	kind     string
}

func (v *VMHandler) GetQueue() *queue.Queue {
	return v.q
}

func (v *VMHandler) Kind() string {
	return v.kind
}

func (v *VMHandler) Check(w http.ResponseWriter, r *http.Request) {
	result := v.resource.Check(hypervisor.Instance{
		VCPU:   24,
		Disk:   4,
		Memory: 4,
	})
	fmt.Fprintf(w, "<h1>%v</h1>", result)
}

func (v *VMHandler) Test(w http.ResponseWriter, r *http.Request) {
	jobs := v.q.GetOneJob(v.Kind())
	fmt.Fprintf(w, "%v\n", jobs)
}

func (v *VMHandler) GetJobs(w http.ResponseWriter, r *http.Request) {
	log.Debug("Get jobs")
	ownerID := r.FormValue("owner_id")
	status := r.FormValue("status")
	jobs := v.q.GetJobs(v.Kind(), status, ownerID)
	b, _ := json.Marshal(&jobs)
	fmt.Fprintf(w, "{\"jobs\": %s, \"ok\": true}", string(b))
}

func (v *VMHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	job := v.q.GetJobByRequestID(vars["request_id"])
	if job == nil {
		fmt.Fprintf(w, "{\"ok\": false}")
	} else {
		b, err := json.Marshal(job)
		if err != nil {
			log.Errorf(`Failed to convert struct to json string.
				Programming err? (%s)`, err)
			fmt.Fprintf(w, "{\"ok\": false, \"error\": \"%s\"}", err)
			return
		}
		fmt.Fprintf(w, "%s", string(b))
	}
}

func (v *VMHandler) UpdateProperty(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID := vars["request_id"]

	decoder, err := jason.NewObjectFromReader(r.Body)
	if err != nil {
		fmt.Fprintf(w, "{\"error\": \"%s\", \"ok\":false }", err)
	}

	priority, err := decoder.GetInt64("priority")
	if err == nil {
		j := v.q.GetJobByRequestID(requestID)
		if j == nil {
			fmt.Fprintf(w, "{\"ok\": false, \"error\": \"job not found\"}")
			return
		}
		err = v.q.SetJobPriority(j, priority)
		if err != nil {
			fmt.Fprintf(w, "{\"error\": \"%s\", \"ok\":false }", err)
			return
		}
		fmt.Fprintf(w, "{\"ok\":true}")
		return
	}
	status, err := decoder.GetString("status")
	if err == nil && status != "" {
		j := v.q.GetJobByRequestID(requestID)
		if j == nil {
			fmt.Fprintf(w, "{\"ok\": false, \"error\": \"job not found\"}")
			return
		}
		err = v.q.SetJobStatus(j, status)
		if err != nil {
			fmt.Fprintf(w, "{\"error\": \"%s\", \"ok\":false }", err)
			return
		}
		fmt.Fprintf(w, "{\"ok\":true}")
		return

	}
	fmt.Fprintf(w, "{\"ok\":false, \"error\": \"not support operation\"}")

}

func (v *VMHandler) AddJob(w http.ResponseWriter, r *http.Request) {
	decoder, err := jason.NewObjectFromReader(r.Body)
	if err != nil {
		fmt.Fprintf(w, "{\"error\": \"input data is not json.(%s)\", \"ok\":false }", err)
		return
	}

	requestID, err := decoder.GetString("request_id")
	if err != nil {
		fmt.Fprintf(w, "{\"error\": \"request_id is not found(%s)\", \"ok\":false }", err)
		return
	}
	ownerID, err := decoder.GetString("owner_id")
	if err != nil {
		fmt.Fprintf(w, "{\"error\": \"owner_id is not found(%s)\", \"ok\":false }", err)
		return
	}
	data, err := decoder.GetObject("data")
	if err != nil {
		fmt.Fprintf(w, "{\"error\": \"data is not found(%s)\", \"ok\":false }", err)
		return
	}
	callback, err := decoder.GetString("callback")
	if err != nil {
		fmt.Fprintf(w, "{\"error\": \"callback url is not found(%s)\", \"ok\":false }", err)
		return
	}

	var j queue.Job

	j.RequestID = requestID
	j.Type = v.Kind()
	j.OwnerID = ownerID
	j.Data = data.String()
	j.Callback = callback

	err = v.q.NewJob(&j)
	if err != nil {
		fmt.Fprintf(w, "{\"ok\": false, \"error\": \"%s\"}", err)
	}
	b, _ := json.Marshal(&j)
	fmt.Fprintf(w, "%s", string(b))
}

func (v *VMHandler) DeleteJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	job := v.q.GetJobByRequestID(vars["request_id"])
	if job == nil {
		fmt.Fprintf(w, "{\"ok\": false, \"error\": \"job not found\"}")
	} else {
		if job.RequestID == "" {
			fmt.Fprintf(w, "{\"ok\": false, \"error\": \"job not found\"}")
		} else {
			v.q.DeleteJob(job)
			fmt.Fprintf(w, "{\"ok\": true}")
		}
	}
}

func InitVMHandler(dbType string, uri string) *VMHandler {
	r := VMHandler{}
	//r.resource.Update()
	r.kind = "vm"
	r.q = queue.Init(dbType, uri)
	return &r
}
