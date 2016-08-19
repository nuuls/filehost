package filehost

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func upload(w http.ResponseWriter, r *http.Request) {
	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		w.Write([]byte("error uploading image"))
		return
	}
	// check if body contains any data
	if len(bs) < 1 {
		w.Write([]byte("https://i.imgur.com/r7FGMh8.png"))
		return
	}
	r.ParseForm()
	if key := r.Form.Get("key"); key != cfg.Key {
		w.Write([]byte("https://i.imgur.com/r7FGMh8.png"))
		return
	}
	var typ string
	if r.Header.Get("file-type") != "" {
		typ, bs = validateFileType(strings.ToLower(r.Header.Get("file-type")), bs)
	} else if ct := r.Header.Get("Content-Type"); ct != "" {
		spl := strings.Split(ct, "/")
		if len(spl) > 1 {
			if strings.Contains(ct, "form-data") { // used by sharex
				typ, bs = getFormat(bs)
			} else {
				typ, bs = validateFileType(strings.ToLower(spl[1]), bs)
			}
		}
	} else {
		typ, bs = getFormat(bs)
	}
	fileName := randString(cfg.UrlLength) + "." + typ
	log.Info(fileName)
	log.Debug(r.Header)
	file, err := os.Create("./files/" + fileName)
	if err != nil {
		log.Error(err)
		w.Write([]byte("error uploading image"))
		return
	}
	defer file.Close()
	file.Write(bs)
	w.Write([]byte(cfg.BaseURL + fileName))
	log.Info("uploaded", fileName, "to", cfg.BaseURL+fileName)
}

var mimeRegex = regexp.MustCompile(fmt.Sprintf(`(\-+\w+)%s.+%sContent-Type: \w+\/(\w+)%s`, "\r\n", "\r\n", "\r\n\r\n"))

func getFormat(file []byte) (string, []byte) {
	log.Debug(len(file))
	matches := mimeRegex.FindSubmatch(file)
	if len(matches) < 3 {
		log.Error("no mime type found")
		return "png", file
	}
	fileType := matches[2]
	log.Debug(string(fileType))
	file = bytes.Replace(file, matches[0], []byte(""), 1)
	file = bytes.Replace(file, matches[1], []byte(""), 1)
	return validateFileType(string(fileType), file)
}

func validateFileType(fileType string, file []byte) (string, []byte) {
	if cfg.isBlocked(fileType) {
		return "fuckyou", []byte("LUL")
	}
	return fileType, file
}
