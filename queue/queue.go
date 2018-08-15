package queue

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Job struct {
	gorm.Model
	RequestID string
	Type      string
	OwnerID   string
	Data      string
	Callback  string
	Priority  int
	Status    string
}

type Queue struct {
	database *gorm.DB
}

func (q *Queue) GetJobByRequestID(ID string) *Job {
	job := new(Job)
	q.database.First(&job, "request_id = ?", ID)
	return job
}

func (q *Queue) SetJobpriority(job *Job, priority int64) error {
	return q.database.Model(&job).Update("priority", priority).Error
}

func (q *Queue) SetJobStatus(job *Job, status string) error {
	return q.database.Model(&job).Update("status", status).Error
}

func (q *Queue) NewJob(job *Job) error {
	job.Status = "queued"
	job.Priority = int(time.Now().UnixNano() - 1532584260621743520)
	return q.database.Create(job).Error
}

func (q *Queue) Close() error {
	return q.database.Close()
}

func (q *Queue) Migration() error {
	return q.database.AutoMigrate(&Job{}).Error
}

func (q *Queue) GetOneJob(kind string) *Job {
	var jobs []Job
	db := q.database.Order("priority")
	db.Take(&jobs, "type = ? and status = ?", kind, "queued")
	if len(jobs) == 1 {
		return &jobs[0]
	}
	return nil
}

func (q *Queue) GetJobs(kind string, status string, ownerID string) []Job {
	var jobs []Job
	db := q.database.Order("priority")
	if ownerID != "" {
		db = db.Where("owner_id = ?", ownerID)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	db.Find(&jobs, "type = ?", kind)
	return jobs
}

func (q *Queue) GetJobsByOwnerID(ownerID string) []Job {
	var jobs []Job
	q.database.Order("priority").Find(&jobs, "owner_id = ?", ownerID)
	return jobs
}

func (j Job) String() string {
	return fmt.Sprintf("%s -> %s,%d,%s", j.RequestID, j.Status, j.Priority, j.Data)
}

func Init(engine string, connectionString string) *Queue {
	queue := new(Queue)
	db, err := gorm.Open(engine, connectionString)
	db.LogMode(false)
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&Job{})
	queue.database = db
	return queue
}
