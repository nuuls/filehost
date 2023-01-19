package api

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/nuuls/filehost/internal/config"
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
	DB     *database.Database
	Log    logrus.FieldLogger
	Config *config.Config
}

func (a *API) Run() error {
	a.log.WithField("addr", a.cfg.Config.Addr).Info("Starting api")
	return http.ListenAndServe(a.cfg.Config.Addr, a.newRouter())
}

func (a *API) newRouter() chi.Router {
	r := chi.NewRouter()

	r.Use(realIPMiddleware)
	r.Use(corsMiddleware)
	r.Use(middleware.Logger)

	r.Route("/v1", func(r chi.Router) {
		r.Post("/signup", a.signup)
		r.Post("/login", a.login)
		r.With(a.authMiddleware).Get("/account", a.getAccount)

		r.With(a.authMiddleware).Get("/uploads", a.getUploads)
		r.With(a.optionalAuthMiddleware).Post("/uploads", a.upload)
		r.With(a.optionalAuthMiddleware).Delete("/uploads/{filename}", a.deleteUpload)

		r.Route("/domains", func(r chi.Router) {
			r.Use(a.authMiddleware)

			r.Post("/", a.createDomain)
			r.Get("/", a.getDomains)
			r.Get("/{id}", a.getDomain)
		})
	})

	r.With(a.optionalAuthMiddleware).Post("/upload", a.upload)
	r.Get("/{filename}", a.serveFile)

	return r
}
