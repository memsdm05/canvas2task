package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

const TaskBaseApi = "https://www.meistertask.com/api"

type Atom struct {
	Name string
	Id   int
}

type Atoms []Atom

func (a Atoms) Find(name string) *Atom {
	for _, atom := range a {
		if atom.Name == name {
			return &atom
		}
	}
	return nil
}

func mtReq(method string, path string, body io.Reader) *http.Request {
	if req, err := http.NewRequest(method, TaskBaseApi+path, body); err != nil {
		panic(err)
	} else {
		return req
	}
}

func retrieveAndFind(req *http.Request, context, target string) Atom {
	var atoms Atoms
	Client.Task.DoJson(req, &atoms)
	maybe := atoms.Find(target)
	if maybe == nil {
		PrintfExit(`%s "%s" could not be found`, context, target)
	}
	return *maybe
}

func RetrieveInfo() (section Atom, labels Atoms) {
	// find project
	project := retrieveAndFind(
		mtReq("GET", "/projects", nil),
		"Project",
		Config.ProjectName,
	)
	projPathvar := PathVar("projects", project.Id)

	// find section
	section = retrieveAndFind(
		mtReq("GET", projPathvar+"/sections", nil),
		"Section",
		Config.SectionName,
	)

	// get labels
	Client.Task.DoJson(
		mtReq("GET", projPathvar+"/labels", nil),
		&labels,
	)

	return
}

type Task struct {
	Name   string     `json:"name"`
	Notes  string     `json:"notes,omitempty"`
	Labels []int      `json:"label_ids,omitempty"`
	Status int        `json:"status,omitempty"`
	Due    *time.Time `json:"due,omitempty"`
}

func (t Task) Create() {
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(t)
	resp, _ := Client.Task.Post(
		TaskBaseApi+PathVar("sections", Cache.SectionId)+"/tasks",
		"application/json",
		&buf,
	)
	resp.Body.Close()
	// b, _ := io.ReadAll(resp.Body)
	// fmt.Println(string(b))
}
