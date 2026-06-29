package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/MonteChristo46/glitch-hunt-cli/internal/client"

	"github.com/MonteChristo46/glitch-hunt-cli/internal/config"

	"github.com/spf13/cobra"
)

func unifiedBase(eventID, imageID string) map[string]interface{} {
	now := time.Now().UTC().Format(time.RFC3339)
	return map[string]interface{}{
		"version": "1.0",
		"event_metadata": map[string]interface{}{
			"event_id":    eventID,
			"timestamp_utc": now,
		},
		"account": map[string]interface{}{
			"id":   "acc_d1e1f8a2-3c4b-5d6e-7f8a-9b0c1d2e3f4a",
			"name": "CLI Simulation Factory",
		},
		"source": map[string]interface{}{
			"device_id":        "cli-simulator",
			"route_key":        "simulation",
			"device_metadata":  map[string]interface{}{},
		},
		"data": map[string]interface{}{
			"image_id":    imageID,
			"captured_at": now,
			"edge_context": map[string]interface{}{},
		},
		"links": map[string]interface{}{
			"view_in_dashboard": fmt.Sprintf("https://app.glitch-hunt.io/images/%s", imageID),
		},
	}
}

var mockPayloads = map[string]func() map[string]interface{}{
	"file.uploaded": func() map[string]interface{} {
		p := unifiedBase("evt_f1e2d3c4-b5a6-7f8e-9d0c-1b2a3f4e5d6c", "img_a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d")
		p["event_metadata"].(map[string]interface{})["event_type"] = "INGESTED"
		p["event_metadata"].(map[string]interface{})["severity"] = "INFO"
		p["analysis"] = map[string]interface{}{
			"is_anomalous":     false,
			"confidence_score": 0.0,
			"threshold_applied": 0.5,
			"model_info":       nil,
			"human_annotation": nil,
		}
		return p
	},
	"ai.completed": func() map[string]interface{} {
		p := unifiedBase("evt_e2f3e4d5-c6b7-8a9f-0e1d-2c3b4a5f6e7d", "img_b2c3d4e5-f6a7-8b9c-0d1e-2f3a4b5c6d7e")
		p["event_metadata"].(map[string]interface{})["event_type"] = "AI_COMPLETED"
		p["event_metadata"].(map[string]interface{})["severity"] = "INFO"
		p["analysis"] = map[string]interface{}{
			"is_anomalous":     false,
			"confidence_score": 0.02,
			"threshold_applied": 0.5,
			"model_info": map[string]interface{}{
				"id":      "mod_76543210-9876-5432-1098-76543210abcd",
				"version": "v2.1-stable",
			},
			"human_annotation": nil,
		}
		return p
	},
	"ai.anomaly": func() map[string]interface{} {
		p := unifiedBase("evt_f3a4b5c6-d7e8-9f0a-1b2c-3d4e5f6a7b8c", "img_c3d4e5f6-a7b8-9c0d-1e2f-3a4b5c6d7e8f")
		p["event_metadata"].(map[string]interface{})["event_type"] = "AI_COMPLETED"
		p["event_metadata"].(map[string]interface{})["severity"] = "WARNING"
		p["analysis"] = map[string]interface{}{
			"is_anomalous":     true,
			"confidence_score": 0.94,
			"threshold_applied": 0.5,
			"model_info": map[string]interface{}{
				"id":      "mod_76543210-9876-5432-1098-76543210abcd",
				"version": "v2.1-stable",
			},
			"human_annotation": nil,
		}
		return p
	},
	"ai.edge-case": func() map[string]interface{} {
		p := unifiedBase("evt_a4b5c6d7-e8f9-0a1b-2c3d-4e5f6a7b8c9d", "img_d4e5f6a7-b8c9-0d1e-2f3a-4b5c6d7e8f9a")
		p["event_metadata"].(map[string]interface{})["event_type"] = "AI_COMPLETED"
		p["event_metadata"].(map[string]interface{})["severity"] = "WARNING"
		p["analysis"] = map[string]interface{}{
			"is_anomalous":     false,
			"confidence_score": 0.48,
			"threshold_applied": 0.5,
			"model_info": map[string]interface{}{
				"id":      "mod_76543210-9876-5432-1098-76543210abcd",
				"version": "v2.1-stable",
			},
			"human_annotation": nil,
		}
		return p
	},
	"ai.anomaly-detected": func() map[string]interface{} {
		p := unifiedBase("evt_b5c6d7e8-f9a0-1b2c-3d4e-5f6a7b8c9d0e", "img_e5f6a7b8-c9d0-1e2f-3a4b-5c6d7e8f9a0b")
		p["event_metadata"].(map[string]interface{})["event_type"] = "AI_ANOMALY_DETECTED"
		p["event_metadata"].(map[string]interface{})["severity"] = "WARNING"
		p["analysis"] = map[string]interface{}{
			"is_anomalous":     true,
			"confidence_score": 0.97,
			"threshold_applied": 0.5,
			"model_info": map[string]interface{}{
				"id":      "mod_76543210-9876-5432-1098-76543210abcd",
				"version": "v2.1-stable",
			},
			"human_annotation": nil,
		}
		return p
	},
	"ai.edge-case-detected": func() map[string]interface{} {
		p := unifiedBase("evt_c6d7e8f9-a0b1-2c3d-4e5f-6a7b8c9d0e1f", "img_f6a7b8c9-d0e1-2f3a-4b5c-6d7e8f9a0b1c")
		p["event_metadata"].(map[string]interface{})["event_type"] = "AI_EDGE_CASE_DETECTED"
		p["event_metadata"].(map[string]interface{})["severity"] = "WARNING"
		p["analysis"] = map[string]interface{}{
			"is_anomalous":     false,
			"confidence_score": 0.49,
			"threshold_applied": 0.5,
			"model_info": map[string]interface{}{
				"id":      "mod_76543210-9876-5432-1098-76543210abcd",
				"version": "v2.1-stable",
			},
			"human_annotation": nil,
		}
		return p
	},
	"ai.normal": func() map[string]interface{} {
		p := unifiedBase("evt_d7e8f9a0-b1c2-3d4e-5f6a-7b8c9d0e1f2a", "img_a7b8c9d0-e1f2-3a4b-5c6d-7e8f9a0b1c2d")
		p["event_metadata"].(map[string]interface{})["event_type"] = "AI_NORMAL"
		p["event_metadata"].(map[string]interface{})["severity"] = "INFO"
		p["analysis"] = map[string]interface{}{
			"is_anomalous":     false,
			"confidence_score": 0.02,
			"threshold_applied": 0.5,
			"model_info": map[string]interface{}{
				"id":      "mod_76543210-9876-5432-1098-76543210abcd",
				"version": "v2.1-stable",
			},
			"human_annotation": nil,
		}
		return p
	},
	"review.anomaly-confirmed": func() map[string]interface{} {
		p := unifiedBase("evt_e8f9a0b1-c2d3-4e5f-6a7b-8c9d0e1f2a3b", "img_b8c9d0e1-f2a3-4b5c-6d7e-8f9a0b1c2d3e")
		p["event_metadata"].(map[string]interface{})["event_type"] = "ANOMALY_CONFIRMED_MANUAL"
		p["event_metadata"].(map[string]interface{})["severity"] = "WARNING"
		p["analysis"] = map[string]interface{}{
			"is_anomalous":     true,
			"confidence_score": 0.94,
			"threshold_applied": 0.5,
			"model_info": map[string]interface{}{
				"id":      "mod_76543210-9876-5432-1098-76543210abcd",
				"version": "v2.1-stable",
			},
			"human_annotation": map[string]interface{}{
				"label":        "TRUE_POSITIVE",
				"annotator_id": "user_441",
				"notes":        "Visual confirmation of stress fracture on housing.",
			},
		}
		return p
	},
	"review.anomaly-rejected": func() map[string]interface{} {
		p := unifiedBase("evt_f9a0b1c2-d3e4-5f6a-7b8c-9d0e1f2a3b4c", "img_c9d0e1f2-a3b4-5c6d-7e8f-9a0b1c2d3e4f")
		p["event_metadata"].(map[string]interface{})["event_type"] = "ANOMALY_REJECTED_MANUAL"
		p["event_metadata"].(map[string]interface{})["severity"] = "INFO"
		p["analysis"] = map[string]interface{}{
			"is_anomalous":     false,
			"confidence_score": 0.88,
			"threshold_applied": 0.5,
			"model_info": map[string]interface{}{
				"id":      "mod_76543210-9876-5432-1098-76543210abcd",
				"version": "v2.1-stable",
			},
			"human_annotation": map[string]interface{}{
				"label":        "FALSE_POSITIVE",
				"annotator_id": "user_441",
				"notes":        "False positive — dirt on lens, not a defect.",
			},
		}
		return p
	},
	"review.normal-confirmed": func() map[string]interface{} {
		p := unifiedBase("evt_a0b1c2d3-e4f5-6a7b-8c9d-0e1f2a3b4c5d", "img_d0e1f2a3-b4c5-6d7e-8f9a-0b1c2d3e4f5a")
		p["event_metadata"].(map[string]interface{})["event_type"] = "NORMAL_CONFIRMED_MANUAL"
		p["event_metadata"].(map[string]interface{})["severity"] = "INFO"
		p["analysis"] = map[string]interface{}{
			"is_anomalous":     false,
			"confidence_score": 0.03,
			"threshold_applied": 0.5,
			"model_info": map[string]interface{}{
				"id":      "mod_76543210-9876-5432-1098-76543210abcd",
				"version": "v2.1-stable",
			},
			"human_annotation": map[string]interface{}{
				"label":        "TRUE_NEGATIVE",
				"annotator_id": "user_441",
				"notes":        "Part is good — no defects found.",
			},
		}
		return p
	},
	"review.defect-missed": func() map[string]interface{} {
		p := unifiedBase("evt_b1c2d3e4-f5a6-7b8c-9d0e-1f2a3b4c5d6e", "img_e1f2a3b4-c5d6-7e8f-9a0b-1c2d3e4f5a6b")
		p["event_metadata"].(map[string]interface{})["event_type"] = "DEFECT_MISSED_MANUAL"
		p["event_metadata"].(map[string]interface{})["severity"] = "CRITICAL"
		p["analysis"] = map[string]interface{}{
			"is_anomalous":     true,
			"confidence_score": 0.08,
			"threshold_applied": 0.5,
			"model_info": map[string]interface{}{
				"id":      "mod_76543210-9876-5432-1098-76543210abcd",
				"version": "v2.1-stable",
			},
			"human_annotation": map[string]interface{}{
				"label":        "FALSE_NEGATIVE",
				"annotator_id": "user_441",
				"notes":        "AI missed a clear defect — model needs retraining.",
			},
		}
		return p
	},
	"review.unclear": func() map[string]interface{} {
		p := unifiedBase("evt_c2d3e4f5-a6b7-8c9d-0e1f-2a3b4c5d6e7f", "img_f2a3b4c5-d6e7-8f9a-0b1c-2d3e4f5a6b7c")
		p["event_metadata"].(map[string]interface{})["event_type"] = "REVIEW_UNCLEAR_MANUAL"
		p["event_metadata"].(map[string]interface{})["severity"] = "WARNING"
		p["analysis"] = map[string]interface{}{
			"is_anomalous":     false,
			"confidence_score": 0.45,
			"threshold_applied": 0.5,
			"model_info": map[string]interface{}{
				"id":      "mod_76543210-9876-5432-1098-76543210abcd",
				"version": "v2.1-stable",
			},
			"human_annotation": map[string]interface{}{
				"label":        "UNCLEAR",
				"annotator_id": "user_441",
				"notes":        "Cannot determine from image — request re-inspection.",
			},
		}
		return p
	},
}

func TriggerCmd(cfg *config.Config) *cobra.Command {
	var forwardURL string

	cmd := &cobra.Command{
		Use:   "trigger <event_type>",
		Short: "Inject a simulated event into your local dev server",
		Long: `Send a mock event directly to your local development server without
going through the cloud. Useful for rapid testing without needing
live data or internet connectivity.

All payloads use the exact UnifiedEvent JSON schema that production
webhooks deliver, so your handler sees identical data.

USAGE:
  huntcli trigger <event_type> [--forward-to <url>]

AVAILABLE EVENT TYPES:

  File Ingest:
    file.uploaded              INGESTED — File uploaded by edge device

  AI Results (legacy format, routed as AI_COMPLETED):
    ai.completed               Normal AI inference (no defect, score 0.02)
    ai.anomaly                 AI detected anomaly (score 0.94)
    ai.edge-case               AI borderline result (score 0.48)

  AI Results (specific event types):
    ai.anomaly-detected        AI_ANOMALY_DETECTED — Defect found (score 0.97)
    ai.edge-case-detected      AI_EDGE_CASE_DETECTED — Borderline (score 0.49)
    ai.normal                  AI_NORMAL — No defect (score 0.02)

  Manual Review Results:
    review.anomaly-confirmed   ANOMALY_CONFIRMED_MANUAL — TRUE_POSITIVE
    review.anomaly-rejected    ANOMALY_REJECTED_MANUAL — FALSE_POSITIVE
    review.normal-confirmed    NORMAL_CONFIRMED_MANUAL — TRUE_NEGATIVE
    review.defect-missed       DEFECT_MISSED_MANUAL — FALSE_NEGATIVE (CRITICAL)
    review.unclear             REVIEW_UNCLEAR_MANUAL — Operator unsure

EXAMPLES:
  huntcli trigger ai.anomaly-detected
  huntcli trigger review.defect-missed --forward-to http://localhost:9000/hooks
  huntcli trigger file.uploaded -f http://localhost:8080/webhooks`,
		Example: `  huntcli trigger ai.anomaly-detected
  huntcli trigger review.defect-missed --forward-to http://localhost:9000/hooks`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			eventType := strings.ToLower(args[0])

			builder, ok := mockPayloads[eventType]
			if !ok {
				fmt.Printf("Error: unknown event type '%s'.\n", eventType)
				fmt.Println("Run 'huntcli trigger --help' for available event types.")
				os.Exit(1)
			}

			if forwardURL == "" {
				forwardURL = cfg.DefaultForward
			}

			payload := builder()
			payloadBytes, _ := json.MarshalIndent(payload, "", "  ")

			fmt.Printf("⚡ Mocking event payload for: %s\n", eventType)
			fmt.Printf("➡ Sending HTTP POST to %s\n", forwardURL)

			apiClient := client.NewAPIClient(cfg.APIEndpoint, 10*time.Second)
			var wg sync.WaitGroup
			wg.Add(1)

			go func() {
				defer wg.Done()
				statusCode, err := apiClient.ForwardEvent(forwardURL, payloadBytes, eventType)
				if err != nil {
					fmt.Printf("❌ [ERROR] Failed to send: %v\n", err)
				} else {
					fmt.Printf("✓ [%d %s] Response received from local server.\n", statusCode, httpStatusText(statusCode))
				}
			}()

			wg.Wait()
		},
	}

	cmd.Flags().StringVarP(&forwardURL, "forward-to", "f", "", "Local URL to forward the mock event to (default: from config)")

	return cmd
}

func httpStatusText(code int) string {
	switch code {
	case 200:
		return "OK"
	case 201:
		return "Created"
	case 202:
		return "Accepted"
	case 204:
		return "No Content"
	case 400:
		return "Bad Request"
	case 401:
		return "Unauthorized"
	case 403:
		return "Forbidden"
	case 404:
		return "Not Found"
	case 500:
		return "Internal Server Error"
	case 502:
		return "Bad Gateway"
	case 503:
		return "Service Unavailable"
	default:
		return ""
	}
}
