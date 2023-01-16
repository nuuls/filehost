package main

import (
	"errors"
	"mime"
	"os"
	"strings"
	"time"

	"github.com/nuuls/filehost/internal/config"
	"github.com/nuuls/filehost/internal/database"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func ImportFilesFromFS(cfg *config.Config, log logrus.FieldLogger, db *database.Database) error {
	entries, err := os.ReadDir(cfg.FallbackFilePath)
	if err != nil {
		return err
	}
	for i, entry := range entries {
		log := log.WithField("i", i).
			WithField("total", len(entries)).
			WithField("filename", entry.Name()).
			WithField("progress", i*100/len(entries))
		log.Info("Processing file")
		if entry.IsDir() {
			continue
		}
		_, err := db.GetUploadByFilename(entry.Name())
		if err == nil {
			// already present in db
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.WithError(err).Error("Unexpected error")
			continue
		}
		fileInfo, err := entry.Info()
		if err != nil {
			log.WithError(err).Error("Failed to read file info")
			continue
		}
		spl := strings.Split(entry.Name(), ".")
		ext := ".txt"
		if len(spl) > 1 {
			ext = "." + spl[1]
		}
		mimeType := mime.TypeByExtension(ext)
		mimeType = strings.Split(mimeType, ";")[0]
		if mimeType == "" {
			mimeType = "text/plain"
		}
		// import file
		upload := database.Upload{
			OwnerID:      nil,
			UploaderIP:   "127.0.0.1",
			TTLSeconds:   nil,
			SizeBytes:    uint(fileInfo.Size()),
			Filename:     entry.Name(),
			MimeType:     mimeType,
			DomainID:     cfg.DefaultDomainID,
			LastViewedAt: time.Now(),
		}

		_, err = db.CreateUpload(upload)
		if err != nil {
			log.WithError(err).Error("failed to import file")
			continue
		}
		log.Info("Imported file")
	}
	return nil
}
