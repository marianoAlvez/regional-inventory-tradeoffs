package domain

const replicaC Region = "C"

// allocationPolicy decides what can be sold during the partition and where
// confirmations are stored. The evaluator handles the later loss.
type allocationPolicy func(Scenario) [2]regionalAllocation

func allocateByQuorum(s Scenario) [2]regionalAllocation {
	return [2]regionalAllocation{
		allocateQuorumRegion(RegionA, s.DemandA, s.Stock, s.Majority),
		allocateQuorumRegion(RegionB, s.DemandB, s.Stock, s.Majority),
	}
}

func allocateQuorumRegion(region Region, demand, stock int, majority Region) regionalAllocation {
	if region != majority {
		return regionalAllocation{outcome: RegionalOutcome{Region: region, Pending: demand}}
	}
	// The majority stores the committed record in the selling region and C.
	return regionalAllocation{
		outcome: allocateRegion(region, demand, stock),
		copies:  []Region{region, replicaC},
	}
}

func allocateByQuota(s Scenario) [2]regionalAllocation {
	return [2]regionalAllocation{
		{outcome: allocateRegion(RegionA, s.DemandA, s.QuotaA), copies: []Region{RegionA}},
		{outcome: allocateRegion(RegionB, s.DemandB, s.Stock-s.QuotaA), copies: []Region{RegionB}},
	}
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
