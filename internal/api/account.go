package api

import (
	"net/http"
	"time"

	"github.com/nuuls/filehost/internal/database"
	"golang.org/x/crypto/bcrypt"
)

type signupRequest struct {
	Username string
	Password string
}

type Account struct {
	ID                 uint   `json:"id"`
	Username           string `json:"username"`
	APIKey             string `json:"apiKey"`
	DefaultDomainID    uint   `json:"defaultDomainId"`
	DefaultExpiryHours int    `json:"defaultExpiryHours"`
}

func (a *API) ToAccount(acc *database.Account) Account {
	return Account{
		ID:                 acc.ID,
		Username:           acc.Username,
		APIKey:             acc.APIKey,
		DefaultDomainID:    Or(acc.DefaultDomainID, a.cfg.Config.DefaultDomainID),
		DefaultExpiryHours: int(Or(acc.DefaultExpiry, time.Hour*24*365*100).Hours()),
	}
}

func (a *API) signup(w http.ResponseWriter, r *http.Request) {
	reqData, err := readJSON[signupRequest](r.Body)
	if err != nil {
		a.writeError(w, 400, ErrInvalidJSON, err.Error())
		return
	}
	username, err := sanitizeUsername(reqData.Username)
	if err != nil {
		a.writeError(w, 400, err.Error())
		return
	}
	password, err := bcrypt.GenerateFromPassword([]byte(reqData.Password), bcrypt.DefaultCost)
	acc, err := a.db.CreateAccount(database.Account{
		Username: username,
		Password: string(password),
		APIKey:   generateAPIKey(),
	})
	if err != nil {
		a.writeError(w, 500, "Failed to create account", err.Error())
		return
	}
	a.writeJSON(w, 201, a.ToAccount(acc))
}

func (a *API) login(w http.ResponseWriter, r *http.Request) {
	reqData, err := readJSON[signupRequest](r.Body)
	if err != nil {
		a.writeError(w, 400, ErrInvalidJSON, err.Error())
		return
	}
	username, err := sanitizeUsername(reqData.Username)
	if err != nil {
		a.writeError(w, 400, err.Error())
		return
	}
	acc, err := a.db.GetAccountByUsername(username)
	if err != nil {
		a.writeError(w, 404, "User not found")
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(acc.Password), []byte(reqData.Password))
	if err != nil {
		a.writeError(w, 400, "Invalid password")
		return
	}
	a.writeJSON(w, 201, a.ToAccount(acc))
}

func (a *API) getAccount(w http.ResponseWriter, r *http.Request) {
	acc := mustGetAccount(r)
	a.writeJSON(w, 200, a.ToAccount(acc))
}
