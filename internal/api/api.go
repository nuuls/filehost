package api

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/nuuls/filehost/internal/database"
	"github.com/sirupsen/logrus"
)

func New(cfg Config) *API {
	return &API{
		cfg: cfg,
		db:  cfg.DB,
		log: cfg.Log,
	}
}

type API struct {
	cfg Config
	db  *database.Database
	log logrus.FieldLogger
}

type Config struct {
	DB   *database.Database
	Log  logrus.FieldLogger
	Addr string
}

func (a *API) Run() error {
	return http.ListenAndServe(a.cfg.Addr, a.newRouter())
}

func (a *API) newRouter() chi.Router {
	r := chi.NewRouter()

	r.Route("/v1", func(r chi.Router) {
		r.Post("/signup", a.signup)
	})

	return r
}
