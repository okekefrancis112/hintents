// Copyright (c) 2024 dotandev/hintents
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/dotandev/hintents/config"
	"github.com/dotandev/hintents/middleware"
)

// ─── Example RPC handler ──────────────────────────────────────────────────────

func rpcHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"jsonrpc": "2.0",
		"result":  "0xdeadbeef",
		"id":      1,
	})
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "ok")
}

func main() {
	// ── Option A: load from environment variables ─────────────────────────────
	//   LOG_VERBOSITY=full LOG_RPC_PAYLOAD=true go run main.go
	cfg := config.FromEnv()

	// ── Option B: build manually for specific environments ───────────────────
	// cfg := &config.LoggingConfig{
	//     Verbosity:     config.VerbosityFull,   // log bodies + headers
	//     MaxBodyBytes:  4096,
	//     LogRPCPayload: true,                   // extract method/params from JSON-RPC body
	//     PrettyPrint:   true,                   // indented JSON (local dev only)
	//     SkipPaths:     []string{"/healthz", "/readyz"},
	//     RedactHeaders: map[string]struct{}{
	//         "authorization": {},
	//         "x-api-key":     {},
	//     },
	// }

	// ── Structured logger setup (Go 1.21+ slog) ───────────────────────────────
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	// ── Wire middleware around your existing mux ──────────────────────────────
	mux := http.NewServeMux()
	mux.HandleFunc("/rpc", rpcHandler)
	mux.HandleFunc("/healthz", healthzHandler) // skipped — no log output

	loggedMux := middleware.Logger(cfg)(mux)

	slog.Info("server starting", "addr", ":8080", "verbosity", cfg.Verbosity)
	if err := http.ListenAndServe(":8080", loggedMux); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}