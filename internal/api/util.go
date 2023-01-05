package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	ErrInvalidJSON = "Invalid JSON"
)

func (a *API) writeError(w http.ResponseWriter, code int, message string, data interface{}) {
	a.writeJSON(w, code, map[string]interface{}{
		"statusCode": code,
		"status":     http.StatusText(code),
		"message":    message,
		"data":       data,
	})
}

func (a *API) writeJSON(w http.ResponseWriter, code int, data interface{}) {
	bs, err := json.Marshal(data)
	if err != nil {
		a.log.WithError(err).WithField("data", data).Error("Failed to encode response as json")
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)

	w.Write(bs)
}

func readJSON[T interface{}](rd io.Reader) (*T, error) {
	bs, err := ioutil.ReadAll(rd)
	if err != nil {
		return nil, err
	}
	out := new(T)
	err = json.Unmarshal(bs, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func generateAPIKey() string {
	bs := make([]byte, 12)
	_, err := rand.Read(bs)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(bs)
}
