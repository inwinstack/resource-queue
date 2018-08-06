package httpHandler

import (
	"fmt"
	"net/http"
	_ "strconv"

	"github.com/antonholmquist/jason"
	"github.com/gorilla/mux"
	"github.com/kjelly/resource-queue/hypervisor"
	"github.com/kjelly/resource-queue/queue"
)

type Handler interface {
	Check(w http.ResponseWriter, r *http.Request)
	GetJobs(w http.ResponseWriter, r *http.Request)
	SetPriority(w http.ResponseWriter, r *http.Request)
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
	fmt.Printf("Test\n")
	fmt.Fprintf(w, "%v\n", jobs)
}

func (v *VMHandler) GetJobs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ownerID := r.FormValue("owner_id")
	status := r.FormValue("status")
	jobs := v.q.GetJobs(v.Kind(), status, ownerID)
	fmt.Fprintf(w, "%s %s %s\n", vars["job_id"], vars["priority"], jobs)
}

func (v *VMHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	owenerID := r.FormValue("owner_id")
	status := r.FormValue("status")
	jobs := v.q.GetJobs(v.Kind(), status, owenerID)
	fmt.Fprintf(w, "%s %s %s\n", vars["job_id"], vars["priority"], jobs)
}

func (v *VMHandler) SetPriority(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID := vars["request_id"]

	decoder, err := jason.NewObjectFromReader(r.Body)
	priority, err := decoder.GetInt64("priority")
	j := v.q.GetJobByRequestID(requestID)
	v.q.SetJobpriority(j, priority)

	allJobs := v.q.GetJobs(v.Kind(), "queued", "")
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(w, "%v\n", allJobs)
}

func (v *VMHandler) AddJob(w http.ResponseWriter, r *http.Request) {
	typeID := v.Kind()

	decoder, err := jason.NewObjectFromReader(r.Body)
	if err != nil {
		fmt.Fprintf(w, "input is not json")
		return
	}

	requestID, err := decoder.GetString("request_id")
	if err != nil {
		fmt.Fprintf(w, "request_id is not found")
		return
	}
	ownerID, err := decoder.GetString("owner_id")
	if err != nil {
		fmt.Fprintf(w, "owner_id is not found")
		return
	}
	data, err := decoder.GetObject("data")
	if err != nil {
		fmt.Fprintf(w, "data is not found")
		return
	}

	var j queue.Job

	j.RequestID = requestID
	j.Type = v.Kind()
	j.OwnerID = ownerID
	j.Data = data.String()

	v.q.NewJob(&j)
	fmt.Fprintf(w, "%s %s\n", j, typeID)
}

func (v *VMHandler) DeleteJob(w http.ResponseWriter, r *http.Request) {

}
func InitVMHandler() *VMHandler {
	r := VMHandler{}
	//r.resource.Update()
	r.kind = "vm"
	r.q = queue.Init("sqlite3", "test.db")
	return &r
}
