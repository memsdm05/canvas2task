package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

const CacheExpiration = 7 * 24 * time.Hour

var NewEntryCreated = errors.New("new entry created")

var (
	Config = config{
		CanvasToken:      "replace with access token (https://CANVAS_DOMAIN/profile/settings)",
		MeistertaskToken: "replace with personal access token (https://www.mindmeister.com/api)",
		ProjectName:      "the project in question",
		SectionName:      "the section that the task will be created in",
		TagToCourse: map[string]int{
			"meistertask tag name": 1234567,
		},
	}
	Cache = cache{
		LastUpdated:     utcNow().Add(-CacheExpiration).Add(-time.Minute),
		CourseIdToTagId: map[int]int{},
	}
	Client client
)

func init() {
	dataDir := initDataDir()

	switch err := load(dataDir+"/config.json", &Config); err {
	case NewEntryCreated:
		PrintfExit("initalized config.json, please go edit it (%s)\n", dataDir)
	case nil:
	default:
		panic(err)
	}

	if err := load(dataDir+"/cache.json", &Cache); err != nil && err != NewEntryCreated {
		panic(err)
	}

	Client.Canvas = NewAuthedClient(Config.CanvasToken)
	Client.Task = NewAuthedClient(Config.MeistertaskToken)

	verifyCache(dataDir + "/cache.json")

}

type config struct {
	CanvasToken      string         `json:"canvas_token"`
	MeistertaskToken string         `json:"meistertask_token"`
	ProjectName      string         `json:"project_name"`
	SectionName      string         `json:"section_name"`
	TagToCourse      map[string]int `json:"tag_to_course"`
}

type cache struct {
	LastUpdated     time.Time   `json:"last_updated"`
	Digest          string      `json:"digest"`
	SectionId       int         `json:"section_id"`
	CourseIdToTagId map[int]int `json:"course_id_to_tag_id"`
}

type client struct {
	Canvas *AuthedClient
	Task   *AuthedClient
}

func utcNow() time.Time {
	return time.Now().UTC()
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func initDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	home += "/.canvas2task"
	if !exists(home) {
		os.Mkdir(home, 0o666)
		fmt.Printf("created new data directory @ %s\n", home)
	}

	return home
}

func save(path string, val any) error {
	f, err := os.Create(path)
	defer f.Close()
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err = enc.Encode(val); err != nil {
		return err
	}
	return NewEntryCreated
}

func load(path string, val any) error {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return save(path, val)
	}
	if err = json.NewDecoder(f).Decode(val); err != nil {
		return err
	}
	return nil
}

func digestConfig() string {
	digest := md5.New()
	json.NewEncoder(digest).Encode(Config)
	return hex.EncodeToString(digest.Sum(nil))
}

func verifyCache(path string) {
	if digestConfig() != Cache.Digest {
		refreshCache(path)
	}

	if time.Since(Cache.LastUpdated) > CacheExpiration {
		refreshCache(path)
	}
}

func refreshCache(path string) {
	section, labels := RetrieveInfo()
	Cache.LastUpdated = utcNow()
	Cache.Digest = digestConfig()
	Cache.SectionId = section.Id
	for _, label := range labels {
		if courseId, ok := Config.TagToCourse[label.Name]; ok {
			Cache.CourseIdToTagId[courseId] = label.Id
		}
	}
	save(path, &Cache)
}
