package dto

type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type UpdateUserStatusRequest struct {
	Status string `json:"status"`
}
