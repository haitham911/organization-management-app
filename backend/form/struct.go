package form

type EmailRequest struct {
	Email string `json:"email" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
type InviteUserRequest struct {
	Email          string `json:"email" binding:"required"`
	OrganizationID uint   `json:"organization_id" binding:"required"`
	Role           string `json:"role" binding:"required"`
}
type TokenResponse struct {
	Token string `json:"token"`
}
type JsonResponse struct {
	Data interface{} `json:"data"`
}
