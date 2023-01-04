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

func (a *API) signup(w http.ResponseWriter, r *http.Request) {
	acc, err := a.db.CreateAccount(database.Account{
		Username: "nuuls",
		Password: "password",
	})
	if err != nil {
		a.writeError(w, 500, "Failed to create account", err.Error())
		return
	}
	a.writeJSON(w, 201, acc)
}
