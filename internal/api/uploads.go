package api

import (
	"errors"
	"fmt"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/nuuls/filehost/internal/database"
)

type Upload struct {
	ID           uint      `json:"id"`
	Owner        *Account  `json:"owner"`
	Filename     string    `json:"filename"`
	MimeType     string    `json:"mimeType"`
	SizeBytes    uint      `json:"sizeBytes"`
	Domain       Domain    `json:"domain"`
	TTLSeconds   *uint     `json:"ttlSeconds"`
	LastViewedAt time.Time `json:"lastViewedAt"`
	Views        uint      `json:"views"`
}

func ToUpload(d *database.Upload) *Upload {
	u := &Upload{}
	u.ID = d.ID
	// u.Owner.From(d.Owner)
	u.Filename = d.Filename
	u.MimeType = d.MimeType
	u.SizeBytes = d.SizeBytes
	u.TTLSeconds = d.TTLSeconds
	u.LastViewedAt = d.LastViewedAt
	u.Views = d.Views
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
	var domain *database.Domain
	var err error
	if acc != nil && acc.DefaultDomainID != nil {
		domain, err = a.db.GetDomainByID(*acc.DefaultDomainID)
	} else {
		domain, err = a.db.GetDomainByID(a.cfg.Config.DefaultDomainID)
	}
	if err != nil {
		a.writeError(w, 500, "Failed to load Domain config")
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

	name := RandomString(5)

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

	err = a.files.Create(fullName, file)
	if err != nil {
		l.WithError(err).Error("Failed to create file")
		a.writeError(w, 500, "Failed to upload file")
		return
	}

	// TODO: fix localhost
	fileURL := fmt.Sprintf("https://%s/%s", domain.Domain, fullName)

	w.Write([]byte(fileURL))
	l.Info("uploaded to ", fileURL)

	upload := database.Upload{
		OwnerID:      nil,
		UploaderIP:   r.RemoteAddr,
		TTLSeconds:   nil,
		SizeBytes:    uint(mpHeader.Size),
		Filename:     fullName,
		MimeType:     mimeType,
		DomainID:     a.cfg.Config.DefaultDomainID,
		LastViewedAt: time.Now(),
	}

	if acc != nil {
		upload.OwnerID = &acc.ID
		if acc.DefaultDomainID != nil {
			upload.DomainID = *acc.DefaultDomainID
		}
	}

	_, err = a.db.CreateUpload(upload)
	if err != nil {
		l.WithError(err).Error("Failed to create upload entry")
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

func (a *API) deleteUpload(w http.ResponseWriter, r *http.Request) {
	acc := getAccount(r)

	filename := chi.URLParam(r, "filename")

	upload, err := a.db.GetUploadByFilename(filename)
	if err != nil {
		a.writeError(w, 404, "Upload not found", err.Error())
		return
	}

	// User is logged in and uploaded the file
	hasAccess := acc != nil &&
		upload.OwnerID != nil &&
		*upload.OwnerID == acc.ID

	// User has the same IP as uploader and file is newer than 24 hours
	if upload.UploaderIP == r.RemoteAddr && time.Since(upload.CreatedAt) < time.Hour*24 {
		hasAccess = true
	}

	if !hasAccess {
		a.writeError(w, 403, "You do not have access to delete this file")
		return
	}

	err = a.files.Delete(filename)
	if err != nil {
		a.writeError(w, 500, "Failed to remove file", err.Error())
		return
	}
	err = a.db.DeleteUpload(upload.ID)
	if err != nil {
		a.writeError(w, 500, "Failed to remove file database entry", err.Error())
		return
	}
	a.writeJSON(w, 200, ToUpload(upload))
}
