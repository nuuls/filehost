package api

import (
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nuuls/filehost/filehost"
	"github.com/nuuls/filehost/internal/database"
)

func (a *API) getUploads(w http.ResponseWriter, r *http.Request) {
	acc := mustGetFromContext[*database.Account](r, ContextKeyAccount)
	uploads, err := a.db.GetUploadsByAccount(acc.ID, 25, 0)
	if err != nil {
		a.writeError(w, 500, "Failed to get uploads", err.Error())
		return
	}
	a.writeJSON(w, 200, map[string]interface{}{
		"data": uploads,
	})
}

func (a *API) upload(w http.ResponseWriter, r *http.Request) {
	acc := getFromContext[*database.Account](r, ContextKeyAccount)
	l := a.log
	defer r.Body.Close()

	mpHeader, err := getFirstFile(r)
	if err != nil {
		a.writeError(w, 400, "Failed to read file", err.Error())
		return
	}

	file, err := mpHeader.Open()
	if err != nil {
		l.Error(err)
	}

	name := filehost.RandString(5)

	l = l.WithField("file", name)
	l.Info("uploading...")
	mimeType := mpHeader.Header.Get("Content-Type")

	allowedMimeTypes := []string{"*/*"}

	if !whiteListed(allowedMimeTypes, mimeType) {
		spl := strings.Split(mpHeader.Filename, ".")
		if len(spl) > 1 {
			ext := spl[len(spl)-1]
			mimeType = mime.TypeByExtension("." + ext)
			l.WithField("mime-type", mimeType).Debug("type from ext")
		}
	}

	if mimeType == MimeTypeOctetStream || mimeType == "" {
		mimeType = "text/plain"
	}

	l = l.WithField("mime-type", mimeType)

	if !whiteListed(allowedMimeTypes, mimeType) {
		l.Warning("mime type not allowed")
		http.Error(w, "Unsupported Media Type", 415)
		return
	}

	extension := ExtensionFromMime(mimeType)
	if extension != "" {
		extension = "." + extension
	}

	fullName := name + extension
	dstPath := filepath.Join("./files", fullName)
	// TODO: check if file exists
	dst, err := os.Create(dstPath)
	if err != nil {
		l.WithError(err).Error("cannot create file")
		http.Error(w, "Internal Server Error", 500)
		return
	}
	_, err = io.Copy(dst, file)
	if err != nil {
		l.WithError(err).Error("failed to save file")
		http.Error(w, "Internal Server Error", 500)
		return
	}

	fileURL := "http://localhost:7417/" + fullName

	w.Write([]byte(fileURL))
	l.Info("uploaded to ", fileURL)

	if acc != nil {
		_, err := a.db.CreateUpload(database.Upload{
			OwnerID:   &(**acc).ID,
			ExpiresAt: time.Now().Add(time.Hour * 24 * 365 * 100), // TODO: get from settings or default
			Filename:  fullName,
			MimeType:  mimeType,
			DomainID:  0, // TODO: get from settings or default
		})
		if err != nil {
			l.WithError(err).Error("Failed to insert upload into DB")
		}
	} else {
		_, err := a.db.CreateUpload(database.Upload{
			OwnerID:   nil,
			ExpiresAt: time.Now().Add(time.Hour * 24 * 365 * 100), // TODO: get from default
			Filename:  fullName,
			MimeType:  mimeType,
			DomainID:  0, // TODO: get from default
		})
		if err != nil {
			l.WithError(err).Error("Failed to insert upload into DB")
		}
	}
}

func getFirstFile(r *http.Request) (*multipart.FileHeader, error) {
	err := r.ParseMultipartForm(1024 * 1024 * 64)
	if err != nil {
		return nil, err
	}
	if r.MultipartForm == nil {
		return nil, errors.New("No file attached")
	}
	files := r.MultipartForm.File
	for _, headers := range files {
		for _, h := range headers {
			return h, nil
		}
	}
	return nil, errors.New("No file attached")
}
