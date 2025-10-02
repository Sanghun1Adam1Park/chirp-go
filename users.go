package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Sanghun1Adam1Park/chirp/internal/auth"
	"github.com/Sanghun1Adam1Park/chirp/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameter struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	type response struct {
		Id          uuid.UUID `json:"id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Email       string    `json:"email"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		writeErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	var param parameter
	if err := json.Unmarshal(data, &param); err != nil {
		writeErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	hashedPassword, err := auth.HashPassword(param.Password)
	if err != nil {
		writeErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	user, err := cfg.queries.CreateUser(r.Context(),
		database.CreateUserParams{
			Email:          param.Email,
			HashedPassword: hashedPassword,
		},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeErrorResponse(w, fmt.Errorf("chirp not found"), http.StatusNotFound)
			return
		}
		writeErrorResponse(w, err, http.StatusConflict)
		return
	}

	res := response{
		Id:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}
	writeSuccessResponse(w, res, http.StatusCreated)
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameter struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	type response struct {
		Id           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		IsChirpyRed  bool      `json:"is_chirpy_red"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		writeErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	var param parameter
	if err := json.Unmarshal(data, &param); err != nil {
		writeErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	user, err := cfg.queries.GetUserByEmail(r.Context(), param.Email)
	if err != nil {
		writeErrorResponse(w, err, http.StatusNotFound)
		return
	}

	if err = auth.CheckPasswordHash(param.Password, user.HashedPassword); err != nil {
		writeErrorResponse(w, err, http.StatusUnauthorized)
		return
	}

	jwt, err := auth.MakeJWT(user.ID, cfg.secret, time.Second*60*60)
	if err != nil {
		writeErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	refreshTokenString, err := auth.MakeRefreshToken()
	if err != nil {
		writeErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	refreshToken, err := cfg.queries.CreateRefreshToken(r.Context(),
		database.CreateRefreshTokenParams{
			Token:     refreshTokenString,
			UserID:    user.ID,
			ExpiresAt: time.Now().AddDate(0, 0, 60),
		},
	)

	res := response{
		Id:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		IsChirpyRed:  user.IsChirpyRed,
		Token:        jwt,
		RefreshToken: refreshToken.Token,
	}
	writeSuccessResponse(w, res, http.StatusOK)
}

func (cfg *apiConfig) handlerUpdateCrendentials(w http.ResponseWriter, r *http.Request) {
	type parameter struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	type response struct {
		Id          uuid.UUID `json:"id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Email       string    `json:"email"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		writeErrorResponse(w, err, http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	param := parameter{}
	if err := decoder.Decode(&param); err != nil {
		writeErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		writeErrorResponse(w, err, http.StatusUnauthorized)
		return
	}

	hashedPassword, err := auth.HashPassword(param.Password)
	if err != nil {
		writeErrorResponse(w, err, http.StatusUnauthorized)
		return
	}

	newUserCred, err := cfg.queries.UpdateUserCredential(
		r.Context(),
		database.UpdateUserCredentialParams{
			ID:             userId,
			Email:          param.Email,
			HashedPassword: hashedPassword,
		},
	)
	if err != nil {
		writeErrorResponse(w, err, http.StatusNotFound)
		return
	}

	res := response{
		Id:          newUserCred.ID,
		CreatedAt:   newUserCred.CreatedAt,
		UpdatedAt:   newUserCred.UpdatedAt,
		Email:       newUserCred.Email,
		IsChirpyRed: newUserCred.IsChirpyRed,
	}
	writeSuccessResponse(w, res, http.StatusOK)
}

func (cfg *apiConfig) handlerUpdateChirpyRed(w http.ResponseWriter, r *http.Request) {
	type parameter struct {
		Event string `json:"event"`
		Data  struct {
			UserId uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		writeErrorResponse(w, err, http.StatusUnauthorized)
		return
	}

	if token != cfg.polka_key {
		writeErrorResponse(w, err, http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	param := parameter{}
	if err := decoder.Decode(&param); err != nil {
		writeErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	if param.Event != "user.upgraded" {
		writeSuccessResponse(w, struct{}{}, http.StatusNoContent)
		return
	}

	userId, err := uuid.Parse(param.Data.UserId.String())
	if err != nil {
		writeErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	_, err = cfg.queries.UpgradeToChirpyRed(
		r.Context(),
		userId,
	)
	if err != nil {
		writeErrorResponse(w, err, http.StatusNotFound)
		return
	}

	writeSuccessResponse(w, struct{}{}, http.StatusNoContent)
}
