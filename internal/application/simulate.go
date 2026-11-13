// Package application defines the inbound simulation use case used by the adapters.
package application

import (
	"fmt"

	"github.com/marianoAlvez/regional-inventory-tradeoffs/internal/domain"
)

// Report is the read model returned to either delivery adapter.
type Report struct {
	Results [2]Result `json:"results"`
}

type Result struct {
	Strategy    domain.Strategy   `json:"strategy"`
	Confirmed   int               `json:"confirmed"`
	Pending     int               `json:"pending"`
	Rejected    int               `json:"rejected"`
	Recoverable int               `json:"recoverable"`
	Lost        int               `json:"lost"`
	Remaining   int               `json:"remaining"`
	Stranded    int               `json:"stranded"`
	Regions     [2]RegionalResult `json:"regions"`
	Timeline    []Event           `json:"timeline"`
}

type RegionalResult struct {
	Region    domain.Region `json:"region"`
	Confirmed int           `json:"confirmed"`
	Pending   int           `json:"pending"`
	Rejected  int           `json:"rejected"`
	Remaining int           `json:"remaining"`
}

type Event struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

func Simulate(s domain.Scenario) (Report, error) {
	outcomes, err := domain.Compare(s)
	if err != nil {
		return Report{}, err
	}
	return Report{
		Results: [2]Result{
			resultFor(outcomes[0], quorumTimeline(s, outcomes[0])),
			resultFor(outcomes[1], quotaTimeline(s, outcomes[1])),
		},
	}, nil
}

func resultFor(outcome domain.Outcome, timeline []Event) Result {
	regions := [2]RegionalResult{}
	for i, region := range outcome.Regions {
		regions[i] = RegionalResult{
			Region:    region.Region,
			Confirmed: region.Confirmed,
			Pending:   region.Pending,
			Rejected:  region.Rejected,
			Remaining: region.Remaining,
		}
	}
	return Result{
		Strategy:    outcome.Strategy,
		Confirmed:   outcome.Confirmed,
		Pending:     outcome.Pending,
		Rejected:    outcome.Rejected,
		Recoverable: outcome.Recoverable,
		Lost:        outcome.Lost,
		Remaining:   outcome.Remaining,
		Stranded:    outcome.Stranded,
		Regions:     regions,
		Timeline:    timeline,
	}
}

func quorumTimeline(s domain.Scenario, outcome domain.Outcome) []Event {
	topology := "Ninguna región de venta alcanza dos de las tres réplicas."
	if s.Majority != domain.None {
		topology = fmt.Sprintf("La región %s conserva mayoría junto al tercer dominio.", s.Majority)
	}
	return []Event{
		{Title: "Partición de red", Detail: topology},
		{
			Title: "Pedidos durante la partición",
			Detail: fmt.Sprintf(
				"Se confirman %d, quedan %d pendientes sin reservar stock y se rechazan %d por falta de unidades.",
				outcome.Confirmed, outcome.Pending, outcome.Rejected,
			),
		},
		recoveryEvent(s.Loss, outcome),
	}
}

func quotaTimeline(s domain.Scenario, outcome domain.Outcome) []Event {
	return []Event{
		{
			Title: "Partición de red",
			Detail: fmt.Sprintf(
				"La región A tiene %d unidades asignadas y la B tiene %d. Los cupos quedan fijos.",
				s.QuotaA, s.Stock-s.QuotaA,
			),
		},
		{
			Title: "Pedidos durante la partición",
			Detail: fmt.Sprintf(
				"Se confirman %d y se rechazan %d. El stock libre en otra región habría cubierto %d pedidos rechazados.",
				outcome.Confirmed, outcome.Rejected, outcome.Stranded,
			),
		},
		recoveryEvent(s.Loss, outcome),
	}
}

func recoveryEvent(loss domain.Region, outcome domain.Outcome) Event {
	if loss == domain.None {
		return Event{
			Title:  "Sin pérdida regional",
			Detail: fmt.Sprintf("Se conservan los registros de las %d confirmaciones.", outcome.Recoverable),
		}
	}
	detail := fmt.Sprintf(
		"Al restablecer la conectividad entre sobrevivientes, se recuperan %d confirmaciones y se pierden %d registros.",
		outcome.Recoverable, outcome.Lost,
	)
	return Event{Title: fmt.Sprintf("Pérdida permanente de la región %s", loss), Detail: detail}
}
