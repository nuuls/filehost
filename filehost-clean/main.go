package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func main() {
	log.SetOutput(os.Stdout)
	var (
		dbPath = flag.String("db", "database.json", "path to database.json file")
		// filePath   = flag.String("files", ".", "path to files")
		maxClicks  = flag.Int("max-clicks", 10, "dont delete files with more than max-clicks clicks")
		maxAgeDays = flag.Int("max-age-days", 365, "dont delete files younger than max-age-days")
		dryRun     = flag.Bool("dry-run", false, "dry run")
	)

	flag.Parse()

	db := loadDatabase(*dbPath)

	filesChecked := 0
	filesDeleted := 0

	for _, item := range db {
		log.Println("Checking: ", item.Name)
		filesChecked++
		if item.Clicks > *maxClicks ||
			time.Since(item.Time) < time.Hour*24*time.Duration(*maxAgeDays) {
			log.Printf("Keeping: %#v", item)
			continue
		}
		log.Printf("Deleting: %#v", item)
		if *dryRun {
			filesDeleted++
			continue
		}
		err := os.Remove(item.Path)
		if err != nil {
			log.Println(err)
		} else {
			filesDeleted++
		}
	}
	log.Printf("Checked files: %d, deleted files: %d", filesChecked, filesDeleted)
}

type File struct {
	Name     string
	Path     string
	MimeType string
	Uploader Uploader
	Time     time.Time
	Expire   int
	Clicks   int
}

type Uploader struct {
	IP        string `json:"ip"`
	UserAgent string `json:"user-agent"`
}

func loadDatabase(path string) map[string]File {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("Failed to read DB file:", err)
	}
	data := map[string]File{}
	err = json.Unmarshal(f, &data)
	if err != nil {
		log.Fatal("failed to decode db:", err)
	}
	return data
}
