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

func successResponse(w http.ResponseWriter, body string, extra map[string]string) {
	fmt.Fprintf(w, "{\"data\": %s, \"ok\": true}", body)
}

func errorResponse(w http.ResponseWriter, body string, extra map[string]string) {
	fmt.Fprintf(w, "{\"error\": \"%s\", \"ok\": false}", body)
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
	successResponse(w, fmt.Sprintf("%s", string(b)), nil)

}

func (v *VMHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	job := v.q.GetJobByRequestID(vars["request_id"])
	if job == nil {
		errorResponse(w, "job not found", nil)
		return
	} else {
		b, err := json.Marshal(job)
		if err != nil {
			log.Errorf(`Failed to convert struct to json string.
				Programming err? (%s)`, err)
			errorResponse(w, fmt.Sprintf("%s", err), nil)
			return
		}
		successResponse(w, string(b), nil)

	}
}

func (v *VMHandler) UpdateProperty(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID := vars["request_id"]

	decoder, err := jason.NewObjectFromReader(r.Body)
	if err != nil {
		errorResponse(w, fmt.Sprintf("%s", err), nil)
	}

	priority, err := decoder.GetInt64("priority")
	if err == nil {
		j := v.q.GetJobByRequestID(requestID)
		if j == nil {
			errorResponse(w, fmt.Sprintf("job not found"), nil)
			return
		}
		err = v.q.SetJobPriority(j, priority)
		if err != nil {
			errorResponse(w, fmt.Sprintf("%s", err), nil)
			return
		}
		successResponse(w, "{}", nil)
		return
	}
	status, err := decoder.GetString("status")
	if err == nil && status != "" {
		j := v.q.GetJobByRequestID(requestID)
		if j == nil {
			errorResponse(w, fmt.Sprintf("job not found"), nil)
			return
		}
		err = v.q.SetJobStatus(j, status)
		if err != nil {
			errorResponse(w, fmt.Sprintf("{\"error\": \"%s\", \"ok\":false }", err), nil)
			return
		}
		successResponse(w, "{}", nil)
		return

	}
	errorResponse(w, fmt.Sprintf("not support operation"), nil)

}

func (v *VMHandler) AddJob(w http.ResponseWriter, r *http.Request) {
	decoder, err := jason.NewObjectFromReader(r.Body)
	if err != nil {
		errorResponse(w, fmt.Sprintf("input data is not json.(%s)", err), nil)
		return
	}

	requestID, err := decoder.GetString("request_id")
	if err != nil {
		errorResponse(w, fmt.Sprintf("request_id is not found(%s)", err), nil)
		return
	}
	ownerID, err := decoder.GetString("owner_id")
	if err != nil {
		errorResponse(w, fmt.Sprintf("owner_id is not found(%s)", err), nil)
		return
	}
	data, err := decoder.GetObject("data")
	if err != nil {
		errorResponse(w, fmt.Sprintf("data is not found(%s)", err), nil)
		return
	}
	callback, err := decoder.GetString("callback")
	if err != nil {
		errorResponse(w, fmt.Sprintf("callback url is not found(%s)", err), nil)
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
		errorResponse(w, fmt.Sprintf("%s", err), nil)
	}
	b, _ := json.Marshal(&j)
	successResponse(w, string(b), nil)
}

func (v *VMHandler) DeleteJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	job := v.q.GetJobByRequestID(vars["request_id"])
	if job == nil {
		errorResponse(w, fmt.Sprintf("job not found"), nil)
	} else {
		if job.RequestID == "" {
			errorResponse(w, fmt.Sprintf("job not found"), nil)
		} else {
			v.q.DeleteJob(job)
			successResponse(w, "{}", nil)
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
