package users

import "github.com/rs/zerolog/log"

func CreateUser(userCreds UserCredentials) (int, error) {
	log.Info().Msg("CreateUser called")
	return 1, nil
}
