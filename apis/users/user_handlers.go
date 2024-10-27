package users

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/defilippomattia/gorest/apis"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

func Register(sd *apis.ServerDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("Register called")

		var userCreds UserCredentials
		if err := json.NewDecoder(r.Body).Decode(&userCreds); err != nil {
			http.Error(w, "invalid JSON format", http.StatusBadRequest)
			log.Error().Err(err).Msg("failed to parse JSON body")
			return
		}

		validate := validator.New()
		if err := validate.Struct(userCreds); err != nil {
			http.Error(w, "request not in valid format", http.StatusBadRequest)
			log.Error().Err(err).Msg("request not in valid format")
			return
		}

		uId, err := CreateUser(userCreds)
		if err != nil {
			http.Error(w, "error saving user", http.StatusInternalServerError)
			log.Error().Err(err).Msg("failed to register user")
		}
		fmt.Println(uId)

		// if err := CreateUser(userCreds); err != nil {
		// 	http.Error(w, "error saving user", http.StatusInternalServerError)
		// 	log.Error().Err(err).Msg("failed to save user to database")
		// 	return
		// }

		log.Info().Msg("Register completed")

	}
}
