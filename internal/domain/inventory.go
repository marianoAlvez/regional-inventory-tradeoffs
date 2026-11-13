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
	witness Region = "C"
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

// confirmationBatch represents one-unit orders with the same origin and
// durable placement. Its count is sufficient for this lab's aggregate report.
type confirmationBatch struct {
	quantity int
	copies   []Region
}

func (b confirmationBatch) surviving(loss Region) int {
	for _, site := range b.copies {
		if site != loss {
			return b.quantity
		}
	}
	return 0
}

// Compare evaluates both policies against the same validated scenario.
func Compare(s Scenario) ([2]Outcome, error) {
	if err := s.Validate(); err != nil {
		return [2]Outcome{}, err
	}
	return [2]Outcome{allocateByQuorum(s), allocateByQuota(s)}, nil
}

func allocateByQuorum(s Scenario) Outcome {
	regions := [2]RegionalOutcome{
		allocateQuorumRegion(RegionA, s.DemandA, s.Stock, s.Majority),
		allocateQuorumRegion(RegionB, s.DemandB, s.Stock, s.Majority),
	}
	// The fixture grants at most one selling region a majority. Its committed
	// entries have a copy in that region and one in the third failure domain.
	batch := confirmationBatch{
		quantity: regions[0].Confirmed + regions[1].Confirmed,
		copies:   []Region{s.Majority, witness},
	}
	return summarize(Quorum, s.Stock, regions, batch.surviving(s.Loss))
}

func allocateQuorumRegion(region Region, demand, stock int, majority Region) RegionalOutcome {
	if region != majority {
		return RegionalOutcome{Region: region, Pending: demand}
	}
	return allocateRegion(region, demand, stock)
}

func allocateByQuota(s Scenario) Outcome {
	a := allocateRegion(RegionA, s.DemandA, s.QuotaA)
	b := allocateRegion(RegionB, s.DemandB, s.Stock-s.QuotaA)
	batchA := confirmationBatch{quantity: a.Confirmed, copies: []Region{RegionA}}
	batchB := confirmationBatch{quantity: b.Confirmed, copies: []Region{RegionB}}
	recoverable := batchA.surviving(s.Loss) + batchB.surviving(s.Loss)
	outcome := summarize(Quota, s.Stock, [2]RegionalOutcome{a, b}, recoverable)
	// At most this many denied orders could use the other region's unused quota.
	outcome.Stranded = min(a.Rejected, b.Remaining) + min(b.Rejected, a.Remaining)
	return outcome
}

func allocateRegion(region Region, demand, stock int) RegionalOutcome {
	confirmed := min(demand, stock)
	return RegionalOutcome{
		Region:    region,
		Confirmed: confirmed,
		Rejected:  demand - confirmed,
		Remaining: stock - confirmed,
	}
}

func summarize(strategy Strategy, stock int, regions [2]RegionalOutcome, recoverable int) Outcome {
	a, b := regions[0], regions[1]
	confirmed := a.Confirmed + b.Confirmed
	return Outcome{
		Strategy:    strategy,
		Confirmed:   confirmed,
		Pending:     a.Pending + b.Pending,
		Rejected:    a.Rejected + b.Rejected,
		Recoverable: recoverable,
		Lost:        confirmed - recoverable,
		Remaining:   stock - confirmed,
		Regions:     regions,
	}
}
