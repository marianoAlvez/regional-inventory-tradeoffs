// Command inventory runs the simulation with a JSON scenario from stdin.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/marianoAlvez/regional-inventory-tradeoffs/internal/adapter/jsoninput"
)

func main() {
	if err := run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(input io.Reader, output io.Writer) error {
	payload, err := io.ReadAll(io.LimitReader(input, jsoninput.MaxInputBytes+1))
	if err != nil {
		return fmt.Errorf("leer escenario: %w", err)
	}
	report, err := jsoninput.Simulate(string(payload))
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}
