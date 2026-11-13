// Package jsoninput is the shared JSON boundary for the CLI and browser.
package jsoninput

import (
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/marianoAlvez/regional-inventory-tradeoffs/internal/application"
	"github.com/marianoAlvez/regional-inventory-tradeoffs/internal/domain"
)

const MaxInputBytes = 4096

type scenarioRequest struct {
	Stock    *int           `json:"stock"`
	QuotaA   *int           `json:"quotaA"`
	DemandA  *int           `json:"demandA"`
	DemandB  *int           `json:"demandB"`
	Majority *domain.Region `json:"majority"`
	Loss     *domain.Region `json:"loss"`
}

func Simulate(payload string) (application.Report, error) {
	if len(payload) > MaxInputBytes {
		return application.Report{}, errors.New("la entrada excede el tamaño admitido")
	}
	var request scenarioRequest
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return application.Report{}, errors.New("el escenario debe ser un objeto JSON con los campos esperados y valores del tipo indicado")
	}
	if err := decoder.Decode(new(json.RawMessage)); err != io.EOF {
		return application.Report{}, errors.New("enviá un único escenario JSON")
	}
	if request.Stock == nil || request.QuotaA == nil || request.DemandA == nil ||
		request.DemandB == nil || request.Majority == nil || request.Loss == nil {
		return application.Report{}, errors.New("completá stock, quotaA, demandA, demandB, majority y loss")
	}
	return application.Simulate(domain.Scenario{
		Stock:    *request.Stock,
		QuotaA:   *request.QuotaA,
		DemandA:  *request.DemandA,
		DemandB:  *request.DemandB,
		Majority: *request.Majority,
		Loss:     *request.Loss,
	})
}
