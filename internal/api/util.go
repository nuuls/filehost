package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	rand2 "math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/nuuls/filehost/internal/database"
	"github.com/sirupsen/logrus"
)

func init() {
	rand2.Seed(time.Now().UnixNano())
}

const (
	ErrInvalidJSON = "Invalid JSON"
)

const MimeTypeOctetStream = "application/octet-stream"

func (a *API) writeError(w http.ResponseWriter, code int, message string, data ...interface{}) {
	out := map[string]interface{}{
		"statusCode": code,
		"status":     http.StatusText(code),
		"message":    message,
	}
	if len(data) == 1 {
		out["data"] = data[0]
	} else if len(data) > 1 {
		out["data"] = data
	}
	a.log.WithFields(logrus.Fields{
		"statusCode": code,
		"message":    message,
		"data":       data,
	}).Warning("Responding with error")
	a.writeJSON(w, code, out)
}

func (a *API) writeJSON(w http.ResponseWriter, code int, data interface{}) {
	// TODO: check if data is coming from database package and refuse to
	// send it to the client
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
	bs := make([]byte, 16)
	_, err := rand.Read(bs)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(bs)
}

func getFromContext[T interface{}](r *http.Request, key interface{}) *T {
	val := r.Context().Value(key)
	if val == nil {
		return nil
	}
	out := val.(T)
	return &out
}

func mustGetFromContext[T interface{}](r *http.Request, key interface{}) T {
	val := getFromContext[T](r, key)
	if val == nil {
		panic("Failed to get context value")
	}
	return *val
}

func mustGetAccount(r *http.Request) *database.Account {
	return mustGetFromContext[*database.Account](r, ContextKeyAccount)
}

func getAccount(r *http.Request) *database.Account {
	acc := getFromContext[*database.Account](r, ContextKeyAccount)
	if acc == nil {
		return nil
	}
	return *acc
}

func whiteListed(allowed []string, input string) bool {
	spl := strings.Split(input, "/")
	if len(spl) < 2 {
		return false
	}
	s1, s2 := spl[0], spl[1]
	for _, a := range allowed {
		if input == a {
			return true
		}
		spl := strings.Split(a, "/")
		if len(spl) < 2 {
			panic("Invalid mime type in allow list")
		}
		passed := 0
		if spl[0] == "*" || spl[0] == s1 {
			passed++
		}
		if spl[1] == "*" || spl[1] == s2 {
			passed++
		}
		if passed > 1 {
			return true
		}
	}
	return false
}

func ExtensionFromMime(mimeType string) string {
	spl := strings.Split(mimeType, "/")
	if len(spl) < 2 {
		return ""
	}
	s1, s2 := spl[0], spl[1]
	switch s1 {
	case "audio":
		switch s2 {
		case "wav", "x-wav":
			return "wav"
		default:
			return "mp3"
		}
	case "image":
		switch s2 {
		case "bmp", "x-windows-bmp":
			return "bmp"
		case "gif":
			return "gif"
		case "x-icon":
			return "ico"
		case "jpeg", "pjpeg":
			return "jpeg"
		case "tiff", "x-tiff":
			return "tif"
		default:
			return "png"
		}
	case "text":
		switch s2 {
		case "html":
			return "html"
		case "css":
			return "css"
		case "javascript":
			return "js"
		case "richtext":
			return "rtf"
		default:
			return "txt"
		}
	case "application":
		switch s2 {
		case "json":
			return "json"
		case "x-gzip":
			return "gz"
		case "javascript", "x-javascript", "ecmascript":
			return "js"
		case "pdf":
			return "pdf"
		case "xml":
			return "xml"
		case "x-compressed", "x-zip-compressed", "zip":
			return "zip"
		}
	case "video":
		switch s2 {
		case "avi":
			return "avi"
		case "quicktime":
			return "mov"
		default:
			return "mp4"
		}
	default:
		return "txt"
	}
	return "txt"
}

type PaginatedResponse struct {
	Total int         `json:"total"`
	Data  interface{} `json:"data"`
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Headers", "*")
		w.Header().Add("Access-Control-Allow-Methods", "*")
		if r.Method == http.MethodOptions {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func Map[T interface{}, O interface{}](arr []T, fn func(T) O) []O {
	out := make([]O, len(arr))
	for i, item := range arr {
		out[i] = fn(item)
	}
	return out
}

func Or[T interface{}](val *T, fallback T) T {
	if val != nil {
		return *val
	}
	return fallback
}

var usernameRegex = regexp.MustCompile(`^\w{3,20}$`)

func sanitizeUsername(username string) (string, error) {
	username = strings.TrimSpace(strings.ToLower(username))
	if !usernameRegex.MatchString(username) {
		return "", errors.New("Username is not allowed")
	}
	return username, nil
}

func RandomString(length int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"
	out := make([]byte, length)
	for i := 0; i < length; i++ {
		out[i] = letters[rand2.Intn(len(letters))]
	}
	return string(out)
}

func realIPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.Header.Get("CF-Connecting-IP")
		if ip != "" {
			r.RemoteAddr = ip
		}
		next.ServeHTTP(w, r)
	})
}
