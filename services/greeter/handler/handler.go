// Package handler contains the greeter module's implementation.
// This is the only code the greeter team writes — everything else is generated.
package handler

import (
	"context"
	"fmt"
)

// Greeting matches the contract type. In a real project this would be
// imported from eca-gen, but the greeter module IS the producer so it
// defines the canonical type.
type Greeting struct {
	Message string `json:"message"`
	From    string `json:"from"`
}

// SayHello implements the greeter contract's SayHello operation.
func SayHello(_ context.Context, request any) (any, error) {
	// The adapter delivers the request as a map[string]any.
	req, ok := request.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid request type")
	}

	name, _ := req["name"].(string)
	if name == "" {
		name = "world"
	}

	return &Greeting{
		Message: fmt.Sprintf("Hello, %s!", name),
		From:    "greeter-module",
	}, nil
}
