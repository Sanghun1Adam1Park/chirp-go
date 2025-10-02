package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Sanghun1Adam1Park/chirp/internal/auth"
	"github.com/Sanghun1Adam1Park/chirp/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameter struct {
		Body string `json:"body"`
	}
	type response struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserId    uuid.UUID `json:"user_id"`
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

	if len(param.Body) > 140 {
		writeErrorResponse(w, errors.New("Chirp is too long"), http.StatusBadRequest)
		return
	}

	chirp, err := cfg.queries.CreateChirp(r.Context(),
		database.CreateChirpParams{
			UserID: userId,
			Body:   filterMessage(param.Body),
		},
	)
	if err != nil {
		writeErrorResponse(w, err, http.StatusConflict)
		return
	}

	res := response{
		Id:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	}
	writeSuccessResponse(w, res, http.StatusCreated)
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserId    uuid.UUID `json:"user_id"`
	}

	authorIDParam := r.URL.Query().Get("author_id")
	sortParam := r.URL.Query().Get("sort")
	if sortParam == "" {
		sortParam = "asc"
	}

	var (
		chirps      []database.Chirp
		err         error
		authorID    uuid.UUID
		hasAuthorID bool
	)

	if authorIDParam != "" {
		parsedAuthorID, parseErr := uuid.Parse(authorIDParam)
		if parseErr != nil {
			writeErrorResponse(w, fmt.Errorf("invalid author_id %q: %w", authorIDParam, parseErr), http.StatusBadRequest)
			return
		}
		authorID = parsedAuthorID
		hasAuthorID = true
	}

	switch sortParam {
	case "asc":
		if hasAuthorID {
			chirps, err = cfg.queries.GetChirpsByAuthorId(r.Context(), authorID)
		} else {
			chirps, err = cfg.queries.GetChirps(r.Context())
		}
	case "desc":
		if hasAuthorID {
			chirps, err = cfg.queries.GetChirpsByAuthorIdDesc(r.Context(), authorID)
		} else {
			chirps, err = cfg.queries.GetChirpsDesc(r.Context())
		}
	default:
		writeErrorResponse(w, fmt.Errorf("invalid sort value %q", sortParam), http.StatusBadRequest)
		return
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeErrorResponse(w, fmt.Errorf("chirp not found"), http.StatusNotFound)
			return
		}
		writeErrorResponse(w, err, http.StatusConflict)
		return
	}

	responses := make([]response, 0, len(chirps))
	for _, chirp := range chirps {
		responses = append(responses, response{
			Id:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserId:    chirp.UserID,
		})
	}

	writeSuccessResponse(w, responses, http.StatusOK)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	type response struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeErrorResponse(w, fmt.Errorf("invalid id %q: %w", idStr, err), http.StatusBadRequest)
		return
	}

	chirp, err := cfg.queries.GetChirpById(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeErrorResponse(w, fmt.Errorf("chirp not found"), http.StatusNotFound)
			return
		}
		writeErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	res := response{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}
	writeSuccessResponse(w, res, http.StatusOK)
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserId    uuid.UUID `json:"user_id"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		writeErrorResponse(w, err, http.StatusUnauthorized)
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		writeErrorResponse(w, err, http.StatusUnauthorized)
		return
	}

	chirpIDStr := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDStr)
	if err != nil {
		writeErrorResponse(w, err, http.StatusBadRequest)
		return
	}

	chirp, err := cfg.queries.GetChirpById(
		r.Context(),
		chirpID,
	)
	if err != nil {
		writeErrorResponse(w, err, http.StatusNotFound)
		return
	}

	if chirp.UserID != userId {
		writeErrorResponse(w, errors.New("chirp does not belong to user"), http.StatusForbidden)
		return
	}

	err = cfg.queries.DeleteChirp(
		r.Context(),
		chirp.ID,
	)
	if err != nil {
		writeErrorResponse(w, err, http.StatusNotFound)
		return
	}

	writeSuccessResponse(w, struct{}{}, http.StatusNoContent)
}
