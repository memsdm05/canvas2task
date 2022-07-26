package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type AuthedClient struct {
	http.Client
	Token string
}

func (a *AuthedClient) RoundTrip(request *http.Request) (*http.Response, error) {
	request.Header.Set("Authorization", "Bearer "+a.Token)
	request.Header.Set("User-Agent", "canvas2task/0.0.0")
	return http.DefaultTransport.RoundTrip(request)
}

func NewAuthedClient(token string) *AuthedClient {
	ac := &AuthedClient{
		Client: http.Client{
			Timeout: 5 * time.Second,
		},
		Token: token,
	}
	ac.Client.Transport = ac
	return ac
}

func (a *AuthedClient) DoJson(req *http.Request, v any) error {
	resp, err := a.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

// ==== utils but im too lazy to make it another file ====

func PrintfExit(format string, a ...any) {
	fmt.Printf(format, a...)
	os.Exit(0)
}

func PathVar(name string, id int) string {
	return fmt.Sprintf("/%s/%d", name, id)
}

func mdLink(text, url string) string {
	return fmt.Sprintf("[%s](%s)", text, url)
}

// ========================================================

func Assignment2Task(a Assignment) Task {
	labels := []int{}
	if id, ok := Cache.CourseIdToTagId[a.CourseId]; ok {
		labels = []int{id}
	}

	return Task{
		Name:   a.Name,
		Due:    a.DueAt,
		Labels: labels,
		Notes:  mdLink("link", a.Page),
		Status: 1,
	}
}

func main() {
	if len(os.Args) < 2 {
		PrintfExit("please provide an assignment link\n")
	}

	link := os.Args[1]
	info, err := ParseAssginmentLink(link)
	if err != nil {
		PrintfExit("%s is not a valid assignment link\n", link)
	}

	assignment := GetAssignment(info)
	Assignment2Task(assignment).Create()
}
