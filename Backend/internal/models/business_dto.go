package models

type BusinessDTO struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type UpdateBusinessRequest struct {
	Name string  `json:"name"`
	Meta JSONMap `json:"meta"`
}
