// hello-eca demonstrates the full ECA loop in ~60 lines.
//
// Both modules (greeter + gateway) run in the same cell, so calls between
// them are in-process — no network, no serialization overhead.
//
// If you split them into separate cells, the only thing that changes is
// the adapter — module code stays the same.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"hello-eca/services/gateway/eca-gen/greeter"
	gatewayImpl "hello-eca/services/gateway/handler"
	greeterImpl "hello-eca/services/greeter/handler"

	"github.com/eca-dev/eca/pkg/adapters"
)

func main() {
	// 1. Create an in-process adapter — both modules share this cell.
	adapter := adapters.NewInProcessAdapter()

	// 2. Register the greeter module's handler.
	//    In a real project, each module does this in its init/startup code.
	adapter.Register("greeter", "SayHello", greeterImpl.SayHello)

	// 3. Create the gateway's greeter client backed by the in-process adapter.
	//    The gateway never imports greeter's source — only the generated client.
	greeterClient := greeter.NewGreeterClient(adapter)
	gateway := &gatewayImpl.Handler{Greeter: greeterClient}

	// 4. Wire up an HTTP endpoint so you can test it.
	//    ?locale=es etc. to see greeter v2's localisation.
	http.HandleFunc("/welcome/{user}", func(w http.ResponseWriter, r *http.Request) {
		user := r.PathValue("user")
		locale := r.URL.Query().Get("locale")
		if locale == "" {
			locale = "en-US"
		}
		result, err := gateway.GetWelcome(context.Background(), map[string]any{
			"user":   user,
			"locale": locale,
		})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	// 8081 so we don't clash with eca-web's default 8080.
	fmt.Println("hello-eca running on http://localhost:8081")
	fmt.Println("  Try: curl http://localhost:8081/welcome/alice")
	fmt.Println("       curl 'http://localhost:8081/welcome/alice?locale=fr'")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
