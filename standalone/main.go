package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/nuuls/filehost"

	"github.com/gorilla/mux"
)

func main() {
	cfg := loadConfig()
	m := mux.NewRouter()
	filehost.Init(m)
	log.Fatal(http.ListenAndServe(cfg.Host, m))
}

type config struct {
	Host string `json:"host"`
}

func loadConfig() *config {
	file, err := ioutil.ReadFile("filehost.json")
	if err != nil {
		log.Fatal(err)
	}
	var c config
	err = json.Unmarshal(file, &c)
	if err != nil {
		log.Fatal(err)
	}
	return &c
}
