package api

import (
	"encoding/json"
	"net/http"
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
