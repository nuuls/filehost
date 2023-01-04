package api

import (
	"net/http"

	"github.com/nuuls/filehost/internal/database"
	"github.com/sirupsen/logrus"
)

func New(cfg Config) *API {
	return &API{
		cfg: cfg,
	}
}

type API struct {
	cfg Config
}

type Config struct {
	DB   *database.Database
	Log  logrus.FieldLogger
	Addr string
}

func (a *API) Run() error {
	return http.ListenAndServe(a.cfg.Addr, nil)
}
