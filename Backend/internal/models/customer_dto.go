package models

type CreateCustomerRequest struct {
	Name    string  `json:"name"`
	Email   *string `json:"email"`
	Contact *string `json:"contact"`
	Meta    JSONMap `json:"meta"`
}

type UpdateCustomerRequest struct {
	Name    string  `json:"name"`
	Email   *string `json:"email"`
	Contact *string `json:"contact"`
	Meta    JSONMap `json:"meta"`
}
