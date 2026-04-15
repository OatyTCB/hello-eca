// Package handler contains the gateway module's implementation.
// It calls the greeter module through the generated client in eca-gen/.
package handler

import (
	"context"
	"fmt"

	// This import is the ONLY way to call another module in ECA.
	// Never import greeter's source directly — use the generated client.
	"hello-eca/services/gateway/eca-gen/greeter"
)

// WelcomePage matches the contract type for this module's response.
type WelcomePage struct {
	Title    string `json:"title"`
	Greeting string `json:"greeting"`
}

// Handler holds the gateway's dependencies.
type Handler struct {
	Greeter *greeter.GreeterClient
}

// GetWelcome implements the gateway contract's GetWelcome operation.
// It calls the greeter module through the generated client.
func (h *Handler) GetWelcome(ctx context.Context, request any) (any, error) {
	req, ok := request.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid request type")
	}

	user, _ := req["user"].(string)
	if user == "" {
		user = "visitor"
	}

	// Call the greeter module — this goes through the adapter.
	// If both modules are in the same cell: in-process function call (no network).
	// If they're in different cells: gRPC/HTTP call (no code change needed).
	result, err := h.Greeter.SayHello(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("calling greeter: %w", err)
	}

	return &WelcomePage{
		Title:    fmt.Sprintf("Welcome, %s", user),
		Greeting: result.Message,
	}, nil
}
