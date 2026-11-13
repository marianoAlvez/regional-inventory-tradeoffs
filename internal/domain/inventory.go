// Package domain models inventory allocation and the survival of confirmation records.
package domain

import (
	"errors"
	"fmt"
)

// MaxUnits keeps the simulator's input range consistent across native Go and WebAssembly.
const MaxUnits = 1_000_000

type Region string

const (
	RegionA Region = "A"
	RegionB Region = "B"
	None    Region = "none"
)

type Strategy string

const (
	Quorum Strategy = "quorum"
	Quota  Strategy = "quota"
)

// Scenario describes the orders observed during a partition, followed by an
// optional permanent loss of one selling region.
type Scenario struct {
	Stock    int
	QuotaA   int
	DemandA  int
	DemandB  int
	Majority Region
	Loss     Region
}

func (s Scenario) Validate() error {
	if s.Stock < 0 || s.Stock > MaxUnits {
		return fmt.Errorf("el stock debe estar entre 0 y %d", MaxUnits)
	}
	if s.QuotaA < 0 || s.QuotaA > s.Stock {
		return errors.New("el cupo de la región A debe estar entre 0 y el stock total")
	}
	if s.DemandA < 0 || s.DemandB < 0 || s.DemandA > MaxUnits || s.DemandB > MaxUnits {
		return fmt.Errorf("la demanda de cada región debe estar entre 0 y %d", MaxUnits)
	}
	if !s.Majority.valid() {
		return errors.New("la región con mayoría debe ser A, B o none")
	}
	if !s.Loss.valid() {
		return errors.New("la región perdida debe ser A, B o none")
	}
	return nil
}

func (r Region) valid() bool {
	return r == RegionA || r == RegionB || r == None
}

type RegionalOutcome struct {
	Region    Region
	Confirmed int
	Pending   int
	Rejected  int
	Remaining int
}

// Outcome measures decisions at the end of the order window. Recoverable and
// Lost evaluate the later failure. Remaining is the pre-failure stock balance.
type Outcome struct {
	Strategy    Strategy
	Confirmed   int
	Pending     int
	Rejected    int
	Recoverable int
	Lost        int
	Remaining   int
	Stranded    int
	Regions     [2]RegionalOutcome
}

// regionalAllocation records a policy's decisions during the partition and
// the replicas that hold its confirmed orders before any region is lost.
type regionalAllocation struct {
	outcome RegionalOutcome
	copies  []Region
}

func (a regionalAllocation) surviving(loss Region) int {
	for _, site := range a.copies {
		if site != loss {
			return a.outcome.Confirmed
		}
	}
	return 0
}

// Compare evaluates both policies against the same validated scenario.
func Compare(s Scenario) ([2]Outcome, error) {
	if err := s.Validate(); err != nil {
		return [2]Outcome{}, err
	}
	return [2]Outcome{
		evaluate(s, Quorum, allocateByQuorum),
		evaluate(s, Quota, allocateByQuota),
	}, nil
}

func evaluate(s Scenario, strategy Strategy, policy allocationPolicy) Outcome {
	regions := policy(s)
	a, b := regions[0].outcome, regions[1].outcome
	confirmed := a.Confirmed + b.Confirmed
	recoverable := regions[0].surviving(s.Loss) + regions[1].surviving(s.Loss)
	return Outcome{
		Strategy:    strategy,
		Confirmed:   confirmed,
		Pending:     a.Pending + b.Pending,
		Rejected:    a.Rejected + b.Rejected,
		Recoverable: recoverable,
		Lost:        confirmed - recoverable,
		Remaining:   s.Stock - confirmed,
		Stranded:    min(a.Rejected, b.Remaining) + min(b.Rejected, a.Remaining),
		Regions:     [2]RegionalOutcome{a, b},
	}
}
