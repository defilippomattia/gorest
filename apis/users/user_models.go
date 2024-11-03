package users

type User struct {
	ID       int
	Username string
	Password string
}

type UserRegistrationRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UserRegistrationSuccessResponse struct {
	ResponseType string `json:"response_type" validate:"required"`
	Message      string `json:"message" validate:"required"`
	UserID       int    `json:"user_id" validate:"required"`
}

type UserRegistrationErrorResponse struct {
	ResponseType string `json:"response_type" validate:"required"`
	Message      string `json:"message" validate:"required"`
}

type UserLoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UserLoginErrorResponse struct {
	ResponseType string `json:"response_type" validate:"required"`
	Message      string `json:"message" validate:"required"`
}
