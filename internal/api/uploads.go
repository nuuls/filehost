package api

import (
	"errors"
	"fmt"
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

type Upload struct {
	ID        uint      `json:"id"`
	Owner     *Account  `json:"owner"`
	ExpiresAt time.Time `json:"expiresAt"`
	Filename  string    `json:"filename"`
	MimeType  string    `json:"mimeType"`
	Domain    Domain    `json:"domain"`
}

func ToUpload(d *database.Upload) *Upload {
	u := &Upload{}
	u.ID = d.ID
	// u.Owner.From(d.Owner)
	u.ExpiresAt = d.ExpiresAt
	u.Filename = d.Filename
	u.MimeType = d.MimeType
	u.Domain = *ToDomain(&d.Domain)
	return u
}

func (a *API) getUploads(w http.ResponseWriter, r *http.Request) {
	acc := mustGetFromContext[*database.Account](r, ContextKeyAccount)
	uploads, err := a.db.GetUploadsByAccount(acc.ID, 25, 0)
	if err != nil {
		a.writeError(w, 500, "Failed to get uploads", err.Error())
		return
	}
	a.writeJSON(w, 200, PaginatedResponse{
		Total: -1, // TODO: fix
		Data:  Map(uploads, ToUpload),
	})
}

func (a *API) upload(w http.ResponseWriter, r *http.Request) {
	acc := getAccount(r)
	domain, err := a.db.GetDomainByID(a.cfg.Config.DefaultDomainID)
	if err != nil {
		a.writeError(w, 500, "Failed to load default config")
		return
	}
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

	if !whiteListed(domain.AllowedMimeTypes, mimeType) {
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

	if !whiteListed(domain.AllowedMimeTypes, mimeType) {
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

	// TODO: fix localhost
	fileURL := fmt.Sprintf("https://%s/%s", domain.Domain, fullName)

	w.Write([]byte(fileURL))
	l.Info("uploaded to ", fileURL)

	if acc != nil {
		domainID := a.cfg.Config.DefaultDomainID
		if acc.DefaultDomainID != nil {
			domainID = *acc.DefaultDomainID
		}
		_, err := a.db.CreateUpload(database.Upload{
			OwnerID:   &acc.ID,
			ExpiresAt: time.Now().Add(time.Hour * 24 * 365 * 100), // TODO: get from settings or default
			Filename:  fullName,
			MimeType:  mimeType,
			DomainID:  domainID,
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
			DomainID:  a.cfg.Config.DefaultDomainID,
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
