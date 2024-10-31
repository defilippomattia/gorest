package users

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

type UserHandler struct {
	repo UserRepository
}

func NewUserHandler(repo UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var usRegReq UserRegistrationRequest

	//todo: try to simplify (duplicate code for err handling)
	err := json.NewDecoder(r.Body).Decode(&usRegReq)
	if err != nil {
		log.Error().Err(err).Msg("could not decode usRegReq")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(UserRegistrationErrorResponse{
			ResponseType: "error",
			Message:      "invalid request - username and password must be provided",
		})
		return
	}

	validate := validator.New()
	err = validate.Struct(usRegReq)
	if err != nil {
		log.Error().Err(err).Msg("json body in request not valid")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(UserRegistrationErrorResponse{
			ResponseType: "error",
			Message:      "invalid request - username and password must be provided",
		})
		return
	}

	userId, err := h.repo.Register(context.Background(), &usRegReq)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(UserRegistrationErrorResponse{
			ResponseType: "error",
			Message:      err.Error(),
		})
		return
	}

	log.Info().
		Int("user_id", userId).
		Msg("user registered successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(UserRegistrationSuccessResponse{
		ResponseType: "success",
		Message:      "user registered successfully",
		UserID:       userId,
	})

}
