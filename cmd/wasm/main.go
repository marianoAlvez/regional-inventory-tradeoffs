//go:build js && wasm

// Command wasm exposes the simulation use case to the browser.
package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/marianoAlvez/regional-inventory-tradeoffs/internal/adapter/jsoninput"
)

func main() {
	simulate := js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) != 1 || args[0].Type() != js.TypeString {
			return errorResponse("enviá un escenario como texto JSON")
		}
		report, err := jsoninput.Simulate(args[0].String())
		if err != nil {
			return errorResponse(err.Error())
		}
		payload, err := json.Marshal(report)
		if err != nil {
			return errorResponse("no se pudo serializar el resultado")
		}
		return string(payload)
	})
	defer simulate.Release()
	js.Global().Set("inventorySimulate", simulate)
	select {} // The page owns the lifetime of this WebAssembly instance.
}

func errorResponse(message string) string {
	payload, _ := json.Marshal(struct {
		Error string `json:"error"`
	}{Error: message})
	return string(payload)
}
