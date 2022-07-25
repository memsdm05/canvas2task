package main

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Assignment struct {
	Name string
	// Description    string
	Id             int
	PointsPossible int        `json:"points_possible"`
	CourseId       int        `json:"course_id"`
	DueAt          *time.Time `json:"due_at"`
	LockAt         *time.Time `json:"lock_at"`
	Page           string     `json:"html_url"`
}

type AssingmentInfo struct {
	Domain       string
	CourseId     int
	AssignmentId int
}

func (a AssingmentInfo) apiUrl() string {
	return fmt.Sprintf("https://%s/api/v1/courses/%d/assignments/%d", a.Domain, a.CourseId, a.AssignmentId)
}

func ParseAssginmentLink(link string) (AssingmentInfo, error) {
	var ret AssingmentInfo
	parsed, err := url.Parse(link)
	if err != nil {
		return ret, err
	}

	ret.Domain = parsed.Host
	fmt.Sscanf(parsed.Path, "/courses/%d/assignments/%d", &ret.CourseId, &ret.AssignmentId)

	return ret, nil
}

func GetAssignment(info AssingmentInfo) Assignment {
	url := info.apiUrl()
	var ret Assignment
	req, _ := http.NewRequest("GET", url, nil)
	Client.Canvas.DoJson(req, &ret)
	return ret
}
