// Filename: internal/data/healthcheck.go
package data

type HealthcheckData struct {
	Status string `json:"status,omitempty"`
	Enviornment string `json:"enviornment,omitempty"`
	Version string `json:"version,omitempty"`
}

