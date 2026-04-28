// Package handler contains the greeter module's implementation.
// This is the only code the greeter team writes — everything else is generated.
package handler

import (
	"context"
	"fmt"
)

// Greeting matches the v2 contract type. The `message` field from v1 was
// renamed to `text`, and `locale` was added.
type Greeting struct {
	Text   string `json:"text"`
	From   string `json:"from"`
	Locale string `json:"locale"`
}

// SayHello implements the greeter contract's SayHello operation.
// v2 takes `locale` as a required input.
func SayHello(_ context.Context, request any) (any, error) {
	req, ok := request.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid request type")
	}

	name, _ := req["name"].(string)
	if name == "" {
		name = "world"
	}
	locale, _ := req["locale"].(string)
	if locale == "" {
		return nil, fmt.Errorf("locale is required (v2 breaking change)")
	}

	return &Greeting{
		Text:   greetingFor(name, locale),
		From:   "greeter-module",
		Locale: locale,
	}, nil
}

func greetingFor(name, locale string) string {
	switch locale {
	case "en-US", "en-GB", "en":
		return fmt.Sprintf("Hello, %s!", name)
	case "es":
		return fmt.Sprintf("¡Hola, %s!", name)
	case "fr":
		return fmt.Sprintf("Bonjour, %s !", name)
	case "de":
		return fmt.Sprintf("Hallo, %s!", name)
	default:
		return fmt.Sprintf("Hello, %s!", name)
	}
}
