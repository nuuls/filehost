package api

import (
	"net/http"

	"github.com/nuuls/filehost/internal/database"
)

type createDomainRequest struct {
	Domain           string
	AccessRequired   bool
	AllowedMimeTypes []string
}

type Domain struct {
	ID               int                   `json:"id"`
	Owner            *Account              `json:"owner"`
	Domain           string                `json:"domain"`
	AccessRequired   bool                  `json:"accessRequired"`
	AllowedMimeTypes []string              `json:"allowedMimeTypes"`
	Status           database.DomainStatus `json:"status"`
}

func (a *API) createDomain(w http.ResponseWriter, r *http.Request) {
	acc := mustGetAccount(r)
	data, err := readJSON[createDomainRequest](r.Body)
	if err != nil {
		a.writeError(w, 400, "Failed to decode json", err.Error())
		return
	}
	domain := database.Domain{
		OwnerID:          acc.ID,
		Domain:           data.Domain,
		AccessRequired:   data.AccessRequired,
		AllowedMimeTypes: data.AllowedMimeTypes,
		Status:           database.DomainStatusPending,
	}
	d, err := a.db.CreateDomain(domain)
	if err != nil {
		a.writeError(w, 500, "Failed to create domain", err.Error())
		return
	}
	a.writeJSON(w, 201, d) // TODO: map to proper type
}

func (a *API) getDomains(w http.ResponseWriter, r *http.Request) {
	domains, err := a.db.GetDomains(25, 0)
	if err != nil {
		a.writeError(w, 500, "Failed to get domains", err.Error())
		return
	}
	a.writeJSON(w, 200, PaginatedResponse{
		Total: -1,      // TODO: fix
		Data:  domains, // TODO: map to proper type
	})
}
