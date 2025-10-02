package main

import (
	"net/http"
	"time"

	"github.com/Sanghun1Adam1Park/chirp/internal/auth"
)

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		writeErrorResponse(w, err, http.StatusUnauthorized)
		return
	}

	refreshToken, err := cfg.queries.GetRefreshTokenByToken(r.Context(), token)
	if err != nil {
		writeErrorResponse(w, err, http.StatusUnauthorized)
		return
	}

	newToken, err := auth.MakeJWT(refreshToken.UserID, cfg.secret, time.Hour)
	if err != nil {
		writeErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	res := response{
		Token: newToken,
	}
	writeSuccessResponse(w, res, http.StatusOK)
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		writeErrorResponse(w, err, http.StatusUnauthorized)
		return
	}

	err = cfg.queries.RevokeUserToken(r.Context(), token)
	if err != nil {
		writeErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	writeSuccessResponse(w, nil, http.StatusNoContent)
}
