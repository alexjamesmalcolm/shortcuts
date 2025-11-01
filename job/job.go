package job

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"shortcuts/counter"
	"shortcuts/safemap"
	"strings"
	"time"
)

type Executer[O any] interface {
	Execute() (O, error)
}

type Status uint8

const (
	StatusDone       Status = 0
	StatusError      Status = 1
	StatusInProgress Status = 2
)

func (s Status) String() string {
	switch s {
	case StatusDone:
		return "done"
	case StatusError:
		return "error"
	case StatusInProgress:
		return "in_progress"
	default:
		return "unknown"
	}
}

func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

type Task struct {
	ID     string          `json:"task_id"`
	Result json.RawMessage `json:"result,omitempty"`
	Status Status          `json:"status"`
	Time   time.Time       `json:"time"`
}

type TaskMaster struct {
	safemap.SafeMap[string, Task]
}

var tasks = safemap.New[string, Task]()
var taskIDCounter counter.Counter

func StartTaskMaster() {
	http.HandleFunc("/tasks/", func(w http.ResponseWriter, r *http.Request) {
		ID := strings.TrimPrefix(r.URL.Path, "/tasks/")
		task, ok := tasks.Get(ID)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "404 Could not find Task with ID %v", ID)
			return
		}
		json, err := json.Marshal(task)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		_, err = w.Write(json)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	})
}

func executeJobAndReportToTask[Input Executer[Output], Output any](task Task, i Input) {
	result, err := i.Execute()
	if err != nil {
		task.Status = StatusError
		tasks.Set(task.ID, task)
		return
	}
	jsonStringResult, err := json.Marshal(result)
	if err != nil {
		task.Status = StatusError
		tasks.Set(task.ID, task)
		return
	}
	task.Result = jsonStringResult
	task.Status = StatusDone
	tasks.Set(task.ID, task)
}

func DefineJob[Input Executer[Output], Output any](path string) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotImplemented)
			fmt.Fprintf(w, "Not Implemented: %v", r.Method)
			return
		}
		var input Input
		err := json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			log.Printf("unexpected error while parsing JSON: %v\n", err)
			_, err = w.Write([]byte(err.Error()))
			if err != nil {
				log.Println(err)
			}
			return
		}
		var task = Task{
			ID:     fmt.Sprint(taskIDCounter.Increment()),
			Result: []byte{},
			Status: StatusInProgress,
			Time:   time.Now(),
		}
		tasks.Set(task.ID, task)
		go executeJobAndReportToTask(task, input)

		taskJSON, err := json.Marshal(task)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, err = w.Write([]byte(err.Error()))
			if err != nil {
				log.Println(err)
			}
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write(taskJSON)
	})
}
