package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"
"github.com/MonteChristo46/glitch-hunt-cli/internal/client"

	"github.com/MonteChristo46/glitch-hunt-cli/internal/config"

	"github.com/spf13/cobra"
)

func ListenCmd(cfg *config.Config) *cobra.Command {
	var (
		forwardURL string
		eventTypes []string
	)

	cmd := &cobra.Command{
		Use:   "listen",
		Short: "Forward live cloud events to your local dev server",
		Long: `Connect to the Glitch Hunt cloud and forward real events as HTTP POST
requests to your local development server.

REQUIREMENTS:
  - Run 'huntcli login' first to authenticate this machine
  - Have your local dev server running on the target URL

EVENTS FORWARDED (when no --event-type filter is set):
  INGESTED                File uploaded by an edge device
  AI_COMPLETED            AI finished processing an image
  AI_ANOMALY_DETECTED     AI found a probable defect
  AI_EDGE_CASE_DETECTED   AI result is near the decision boundary
  AI_NORMAL               AI confirmed the part is good
  ANOMALY_CONFIRMED_MANUAL  Operator agreed with AI anomaly
  ANOMALY_REJECTED_MANUAL   Operator disagreed (false positive)
  NORMAL_CONFIRMED_MANUAL   Operator confirmed normal
  DEFECT_MISSED_MANUAL      Operator found a defect AI missed
  REVIEW_UNCLEAR_MANUAL     Operator could not determine

Each event is sent as an HTTP POST with:
  - Content-Type: application/json
  - X-Event-Type: <event_type>
  - Body: UnifiedEvent JSON payload

Use --event-type to subscribe to specific event types only.

EXAMPLES:
  # Forward all events to your local webhook endpoint
  huntcli listen --forward-to http://localhost:8080/webhooks

  # Forward only anomaly and edge-case events
  huntcli listen --forward-to http://localhost:8080/webhooks \
    --event-type AI_ANOMALY_DETECTED \
    --event-type AI_EDGE_CASE_DETECTED

  # Forward to a default URL (configured in ~/.config/hunt/config.json)
  huntcli listen`,
		Example: `  huntcli listen
  huntcli listen --forward-to http://localhost:8080/webhooks
  huntcli listen -f http://localhost:9000/hooks --event-type AI_ANOMALY_DETECTED --event-type AI_EDGE_CASE_DETECTED`,
		Run: func(cmd *cobra.Command, args []string) {
			if cfg.AuthToken == "" {
				fmt.Println("Error: not authenticated. Run 'hunt login' first.")
				os.Exit(1)
			}

			if forwardURL == "" {
				forwardURL = cfg.DefaultForward
			}

			if !strings.HasPrefix(forwardURL, "http://") && !strings.HasPrefix(forwardURL, "https://") {
				forwardURL = "http://" + forwardURL
			}

			apiClient := client.NewAPIClient(cfg.APIEndpoint, 10*time.Second)

			fmt.Printf("🎯 Connecting to cloud infrastructure...\n")
			fmt.Printf("🚀 Forwarding cloud events → %s\n", forwardURL)
			if len(eventTypes) > 0 {
				fmt.Printf("📋 Filtering event types: %s\n", strings.Join(eventTypes, ", "))
			}
			fmt.Println("Ready for events (Press Ctrl+C to quit)...")
			fmt.Println()

			wsClient := client.NewWSClient(cfg.APIEndpoint, cfg.AuthToken, eventTypes)

			wsClient.OnEvent(func(eventType string, payload json.RawMessage) {
				ts := time.Now().Format("15:04:05")
				statusCode, err := apiClient.ForwardEvent(forwardURL, []byte(payload), eventType)

				var label string
				if eventType == "" {
					label = "EVENT"
				} else {
					label = eventType
				}

				if err != nil {
					fmt.Printf("[%s] ❌ %s → forward failed: %v\n", ts, label, err)
				} else {
					fmt.Printf("[%s] ✅ %s → %s (%d)\n", ts, label, forwardURL, statusCode)
				}
			})

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, os.Interrupt)

			go func() {
				<-sigCh
				fmt.Println("\n👋 Shutting down...")
				cancel()
			}()

			for {
				if err := wsClient.Connect(ctx); err != nil {
					if ctx.Err() != nil {
						return
					}
					fmt.Printf("❌ Connection error: %v\n", err)
					return
				}

				if err := wsClient.Listen(ctx); err != nil {
					if ctx.Err() != nil {
						return
					}
					fmt.Printf("⚠ Connection lost: %v (reconnecting...)\n", err)
					wsClient.Close()
				}
			}
		},
	}

	cmd.Flags().StringVarP(&forwardURL, "forward-to", "f", "", "Local URL to forward events to (default: from config)")
	cmd.Flags().StringArrayVar(&eventTypes, "event-type", nil, "Filter specific event types (can be specified multiple times)")

	return cmd
}
