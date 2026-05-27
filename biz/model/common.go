package model

type HealthResponse struct {
	Service string `json:"service"`
	Version string `json:"version"`
	Env     string `json:"env"`
	Status  string `json:"status"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
