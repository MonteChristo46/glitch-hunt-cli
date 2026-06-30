package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/MonteChristo46/glitch-hunt-cli/internal/client"
	"github.com/MonteChristo46/glitch-hunt-cli/internal/config"
	"github.com/MonteChristo46/glitch-hunt-cli/internal/device"

	"github.com/mdp/qrterminal/v3"
	"github.com/spf13/cobra"
)

func LoginCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate this CLI with your Glitch Hunt account",
		Long: `Authenticate this CLI by pairing it with your Glitch Hunt account.

This is required before running 'huntcli listen'.

HOW IT WORKS:
  1. The CLI generates a device ID from your machine's MAC address
  2. It requests a pairing code from the cloud and displays a QR code
  3. Open the URL or scan the QR code in your browser (must be logged in)
  4. Click "Claim Device" in the web UI
  5. The CLI detects the claim and saves the API key automatically

You only need to do this once. The token is stored in ~/.config/hunt/config.json.`,

		Example: `  huntcli login`,

		Run: func(cmd *cobra.Command, args []string) {
			if cfg.AuthToken != "" {
				force, _ := cmd.Flags().GetBool("force")
				if !force {
					fmt.Println("✓ Already authenticated.")
					return
				}
			}

			if cfg.DeviceID == "" {
				mac, err := device.GetMACAddress()
				if err == nil && mac != "" {
					cfg.DeviceID = "huntcli-" + mac
				} else {
					cfg.DeviceID = fmt.Sprintf("huntcli-%d", time.Now().Unix())
				}
				if err := config.Save(cfg); err != nil {
					fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
					os.Exit(1)
				}
				fmt.Printf("🔧 Generated device ID: %s\n", cfg.DeviceID)
			}

			apiClient := client.NewAPIClient(cfg.IngestEndpoint, 30*time.Second)

			fmt.Println("🔑 Requesting pairing code...")
			pairingResp, err := apiClient.RequestPairingCode(cfg.DeviceID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: pairing request failed: %v\n", err)
				os.Exit(1)
			}

			claimURL := fmt.Sprintf("%s/claim/%s", strings.TrimSuffix(cfg.WebClientURL, "/"), pairingResp.Code)

			fmt.Println("\n==========================================")
			fmt.Println(" 📱 SCAN TO CLAIM DEVICE")
			fmt.Printf(" Code: %s\n", pairingResp.Code)
			fmt.Printf(" URL:  %s\n", claimURL)
			fmt.Println("==========================================")

			qrterminal.GenerateHalfBlock(claimURL, qrterminal.L, os.Stdout)

			fmt.Println("\n⏳ Waiting for device to be claimed (Ctrl+C to skip)...")

			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()

			done := make(chan bool, 1)
			go func() {
				for range ticker.C {
					statusResp, err := apiClient.CheckPairingStatus(cfg.DeviceID, pairingResp.Code)
					if err != nil {
						fmt.Printf("  ⚠ Status check failed: %v\n", err)
						continue
					}

					switch statusResp.Status {
					case client.PairingStatusClaimed:
						fmt.Println("\n✅ Device successfully claimed!")
						if statusResp.APIKey != nil {
							cfg.AuthToken = *statusResp.APIKey
							if err := config.Save(cfg); err != nil {
								fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
								os.Exit(1)
							}
							fmt.Println("✓ Authentication token saved.")
						} else {
							fmt.Println("⚠ Claimed but no API key received.")
						}
						done <- true
						return
					case client.PairingStatusExpired:
						fmt.Println("\n❌ Pairing code expired. Run 'hunt login' again.")
						done <- true
						return
					}
				}
			}()

			<-done
		},
	}

	cmd.Flags().BoolP("force", "f", false, "Force re-authentication even if already logged in")
	return cmd
}
