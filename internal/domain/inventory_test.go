package domain_test

import (
	"testing"

	"github.com/marianoAlvez/regional-inventory-tradeoffs/internal/domain"
)

func TestAsymmetricDemandThenLoss(t *testing.T) {
	outcomes, err := domain.Compare(domain.Scenario{
		Stock: 120, QuotaA: 60, DemandA: 90, DemandB: 10,
		Majority: domain.RegionA, Loss: domain.RegionA,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := [2]domain.Outcome{
		{
			Strategy: domain.Quorum, Confirmed: 90, Pending: 10, Recoverable: 90, Remaining: 30,
			Regions: [2]domain.RegionalOutcome{
				{Region: domain.RegionA, Confirmed: 90, Remaining: 30},
				{Region: domain.RegionB, Pending: 10},
			},
		},
		{
			Strategy: domain.Quota, Confirmed: 70, Rejected: 30, Recoverable: 10,
			Lost: 60, Remaining: 50, Stranded: 30,
			Regions: [2]domain.RegionalOutcome{
				{Region: domain.RegionA, Confirmed: 60, Rejected: 30},
				{Region: domain.RegionB, Confirmed: 10, Remaining: 50},
			},
		},
	}
	if outcomes != want {
		t.Fatalf("outcomes = %+v\nwant = %+v", outcomes, want)
	}
}

func TestQuotaWinsOnImmediateConfirmationsWhenNoSideHasMajority(t *testing.T) {
	outcomes, err := domain.Compare(domain.Scenario{
		Stock: 120, QuotaA: 60, DemandA: 45, DemandB: 50,
		Majority: domain.None, Loss: domain.RegionB,
	})
	if err != nil {
		t.Fatal(err)
	}
	quorum, quota := outcomes[0], outcomes[1]
	if quorum.Confirmed != 0 || quorum.Pending != 95 || quorum.Remaining != 120 {
		t.Errorf("quorum = %+v", quorum)
	}
	if quota.Confirmed != 95 || quota.Recoverable != 45 || quota.Lost != 50 {
		t.Errorf("quota = %+v", quota)
	}
}

func TestConservationAcrossSmallScenarios(t *testing.T) {
	regions := []domain.Region{domain.RegionA, domain.RegionB, domain.None}
	for stock := 0; stock <= 4; stock++ {
		for quota := 0; quota <= stock; quota++ {
			for demandA := 0; demandA <= 6; demandA++ {
				for demandB := 0; demandB <= 6; demandB++ {
					for _, majority := range regions {
						for _, loss := range regions {
							s := domain.Scenario{
								Stock: stock, QuotaA: quota, DemandA: demandA, DemandB: demandB,
								Majority: majority, Loss: loss,
							}
							checkConservation(t, s)
						}
					}
				}
			}
		}
	}
}

func checkConservation(t *testing.T, s domain.Scenario) {
	t.Helper()
	outcomes, err := domain.Compare(s)
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range outcomes {
		if result.Confirmed+result.Pending+result.Rejected != s.DemandA+s.DemandB ||
			result.Confirmed+result.Remaining != s.Stock ||
			result.Recoverable+result.Lost != result.Confirmed ||
			result.Confirmed < 0 || result.Pending < 0 || result.Rejected < 0 ||
			result.Remaining < 0 || result.Recoverable < 0 || result.Lost < 0 ||
			result.Stranded < 0 || result.Stranded > min(result.Rejected, result.Remaining) {
			t.Fatalf("conservation failed: scenario %+v, result %+v", s, result)
		}
		if result.Strategy == domain.Quorum && result.Lost != 0 {
			t.Fatalf("quorum lost confirmed records: scenario %+v, result %+v", s, result)
		}
		if result.Strategy == domain.Quota && result.Pending != 0 {
			t.Fatalf("quota queued orders: scenario %+v, result %+v", s, result)
		}
	}
}

func TestRegionSwapPreservesTotals(t *testing.T) {
	scenario := domain.Scenario{
		Stock: 80, QuotaA: 55, DemandA: 47, DemandB: 92,
		Majority: domain.RegionB, Loss: domain.RegionB,
	}
	swapped := domain.Scenario{
		Stock: 80, QuotaA: 25, DemandA: 92, DemandB: 47,
		Majority: domain.RegionA, Loss: domain.RegionA,
	}
	original, err := domain.Compare(scenario)
	if err != nil {
		t.Fatal(err)
	}
	mirrored, err := domain.Compare(swapped)
	if err != nil {
		t.Fatal(err)
	}
	for i, result := range original {
		other := mirrored[i]
		result.Regions, other.Regions = [2]domain.RegionalOutcome{}, [2]domain.RegionalOutcome{}
		if result != other {
			t.Fatalf("region swap changed totals: %+v and %+v", result, other)
		}
	}
}
