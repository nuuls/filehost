package api

import (
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
)

func (a *API) serveFile(w http.ResponseWriter, r *http.Request) {
	l := a.log
	name := chi.URLParam(r, "filename")
	name = filepath.Base(name)
	l = l.WithField("file", name)
	// if ratelimited(r.RemoteAddr) {
	// 	l.Warning("ratelimited")
	// 	http.Error(w, "Rate Limit Exceeded", 429)
	// 	return
	// }
	file, err := a.files.Get(name)
	if err != nil {
		http.Error(w, "File not found", 404)
		return
	}
	defer func() {
		// TODO: fix
		if closer, ok := file.(io.Closer); ok {
			closer.Close()
		}
	}()
	spl := strings.Split(name, ".")
	extension := ""
	if len(spl) > 1 {
		extension = spl[len(spl)-1]
	}
	mimeType := mime.TypeByExtension("." + extension)
	// TODO: Get mime type from DB
	if mimeType == MimeTypeOctetStream {
		switch extension {
		case "mp3":
			mimeType = "audio/mpeg"
		case "wav":
			mimeType = "audio/wav"
		default:
			mimeType = "text/plain"
		}
	}
	if strings.HasPrefix(mimeType, "text") {
		mimeType += "; charset=utf-8"
	}
	if dl, _ := strconv.ParseBool(r.URL.Query().Get("download")); dl {
		mimeType = MimeTypeOctetStream
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", mimeType)

	http.ServeContent(w, r, "", time.Time{}, file)

	err = a.db.IncUploadViews(name)
	if err != nil {
		l.WithError(err).Error("Failed to increment file views")
		return
	}
}
