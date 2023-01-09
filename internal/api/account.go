package api

import (
	"net/http"

	"github.com/nuuls/filehost/internal/database"
	"golang.org/x/crypto/bcrypt"
)

type signupRequest struct {
	Username string
	Password string
}

type Account struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	APIKey   string `json:"apiKey"`
}

func (a *API) signup(w http.ResponseWriter, r *http.Request) {
	reqData, err := readJSON[signupRequest](r.Body)
	if err != nil {
		a.writeError(w, 400, ErrInvalidJSON, err.Error())
		return
	}
	password, err := bcrypt.GenerateFromPassword([]byte(reqData.Password), bcrypt.DefaultCost)
	acc, err := a.db.CreateAccount(database.Account{
		Username: reqData.Username,
		Password: string(password),
		APIKey:   generateAPIKey(),
	})
	if err != nil {
		a.writeError(w, 500, "Failed to create account", err.Error())
		return
	}
	a.writeJSON(w, 201, Account{
		ID:       acc.ID,
		Username: acc.Username,
		APIKey:   acc.APIKey,
	})
}

func (a *API) login(w http.ResponseWriter, r *http.Request) {
	reqData, err := readJSON[signupRequest](r.Body)
	if err != nil {
		a.writeError(w, 400, ErrInvalidJSON, err.Error())
		return
	}
	acc, err := a.db.GetAccountByUsername(reqData.Username)
	if err != nil {
		a.writeError(w, 404, "User not found")
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(acc.Password), []byte(reqData.Password))
	if err != nil {
		a.writeError(w, 400, "Invalid password")
		return
	}
	a.writeJSON(w, 201, Account{
		ID:       acc.ID,
		Username: acc.Username,
		APIKey:   acc.APIKey,
	})
}

func (a *API) getAccount(w http.ResponseWriter, r *http.Request) {
	acc := mustGetAccount(r)
	a.writeJSON(w, 200, Account{
		ID:       acc.ID,
		Username: acc.Username,
		APIKey:   acc.APIKey,
	})
}
