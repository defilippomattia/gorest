package healthz

import (
	"context"
)

type HealthCheckOutput struct {
	Body struct {
		Message string `json:"message"`
	} `json:"body"`
}

func GetHealth(ctx context.Context, input *struct{}) (*HealthCheckOutput, error) {
	resp := &HealthCheckOutput{}
	resp.Body.Message = "OK"
	return resp, nil
}
