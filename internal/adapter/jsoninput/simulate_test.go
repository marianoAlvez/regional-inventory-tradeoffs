package jsoninput_test

import (
	"strings"
	"testing"

	"github.com/marianoAlvez/regional-inventory-tradeoffs/internal/adapter/jsoninput"
)

func TestSimulateRejectsInvalidScenarios(t *testing.T) {
	for name, input := range map[string]string{
		"missing":          `{"stock":5,"quotaA":3,"demandA":1,"majority":"A","loss":"none"}`,
		"null":             `{"stock":null,"quotaA":0,"demandA":0,"demandB":0,"majority":"A","loss":"none"}`,
		"fraction":         `{"stock":5.5,"quotaA":3,"demandA":1,"demandB":1,"majority":"A","loss":"none"}`,
		"outside range":    `{"stock":1000001,"quotaA":3,"demandA":1,"demandB":1,"majority":"A","loss":"none"}`,
		"negative demand":  `{"stock":5,"quotaA":3,"demandA":-1,"demandB":1,"majority":"A","loss":"none"}`,
		"unfunded quota":   `{"stock":5,"quotaA":6,"demandA":1,"demandB":1,"majority":"A","loss":"none"}`,
		"unknown region":   `{"stock":5,"quotaA":3,"demandA":1,"demandB":1,"majority":"C","loss":"none"}`,
		"unknown property": `{"stock":5,"quotaA":3,"demandA":1,"demandB":1,"majority":"A","loss":"none","region":"A"}`,
		"trailing JSON":    `{"stock":5,"quotaA":3,"demandA":1,"demandB":1,"majority":"A","loss":"none"} {}`,
		"oversize":         strings.Repeat(" ", jsoninput.MaxInputBytes+1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := jsoninput.Simulate(input); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
