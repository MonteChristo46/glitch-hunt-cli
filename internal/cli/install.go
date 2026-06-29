package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/MonteChristo46/glitch-hunt-cli/internal/client"
	"github.com/MonteChristo46/glitch-hunt-cli/internal/config"
	"github.com/MonteChristo46/glitch-hunt-cli/internal/device"
	"github.com/MonteChristo46/glitch-hunt-cli/internal/util"

	"github.com/mdp/qrterminal/v3"
	"github.com/spf13/cobra"
)

func getDefaultInstallDir() string {
	if runtime.GOOS == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData != "" {
			return filepath.Join(localAppData, "huntcli")
		}
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "huntcli")
	}
	return filepath.Join(util.GetRealUserHome(), ".local", "bin")
}

func promptInstall(label, defaultValue string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [%s]: ", label, defaultValue)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}
	return input
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	info, err := os.Stat(src)
	if err == nil {
		err = os.Chmod(dst, info.Mode())
	}
	return err
}

func isDirInPath(dir string) bool {
	pathEnv := os.Getenv("PATH")
	for _, p := range filepath.SplitList(pathEnv) {
		abs, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		target, err := filepath.Abs(dir)
		if err != nil {
			continue
		}
		if abs == target {
			return true
		}
	}
	return false
}

func InstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install huntcli to your system and authenticate",
		Long: `Install huntcli to a permanent location on your machine and pair it
with your Glitch Hunt account.

The installer copies the binary to a directory in your PATH (default:
	~/.local/bin on Linux/macOS, %LOCALAPPDATA%\huntcli on Windows),
generates a device ID, and walks you through account pairing.

No administrator privileges are required. Everything stays in your
user directory.

After installation you can run 'huntcli' from anywhere.`,
		Example: `  huntcli install
  huntcli install --dir /custom/path`,
		Run: func(cmd *cobra.Command, args []string) {
			skipLogin, _ := cmd.Flags().GetBool("skip-login")
			dirFlag, _ := cmd.Flags().GetString("dir")

			defaultDir := getDefaultInstallDir()
			if dirFlag != "" {
				defaultDir = dirFlag
			}
			targetDir := promptInstall("Install Directory", defaultDir)

			absTarget, err := filepath.Abs(targetDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] Invalid path: %v\n", err)
				os.Exit(1)
			}

			if err := os.MkdirAll(absTarget, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] Cannot create directory: %v\n", err)
				os.Exit(1)
			}

			currentExe, err := os.Executable()
			if err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] Cannot find current executable: %v\n", err)
				os.Exit(1)
			}

			exeName := filepath.Base(currentExe)
			targetExe := filepath.Join(absTarget, exeName)

			realCurrent, _ := filepath.EvalSymlinks(currentExe)
			realTarget, _ := filepath.EvalSymlinks(targetExe)

			if realCurrent != realTarget {
				fmt.Printf("[STATUS] Copying binary to %s...\n", targetExe)
				os.Remove(targetExe)
				if err := copyFile(currentExe, targetExe); err != nil {
					fmt.Fprintf(os.Stderr, "[ERROR] Copy failed: %v\n", err)
					os.Exit(1)
				}
				if runtime.GOOS == "darwin" {
					fmt.Println("[STATUS] Applying ad-hoc code signature...")
					exec.Command("codesign", "--force", "--deep", "-s", "-", targetExe).Run()
				}
				fmt.Println("[SUCCESS] Binary installed.")
			} else {
				fmt.Println("[STATUS] Already running from target location. Skipping copy.")
			}

			pathInDir := isDirInPath(absTarget)
			if !pathInDir {
				fmt.Println("\n[INFO] Target directory is not in your PATH.")
				fmt.Println("To add it, run one of the following commands:")
				switch runtime.GOOS {
				case "windows":
					fmt.Printf("  setx PATH \"%%PATH%%;%s\"\n", absTarget)
				default:
					shell := filepath.Base(os.Getenv("SHELL"))
					switch shell {
					case "zsh":
						fmt.Printf("  echo 'export PATH=\"$PATH:%s\"' >> ~/.zshrc && source ~/.zshrc\n", absTarget)
					case "bash":
						fmt.Printf("  echo 'export PATH=\"$PATH:%s\"' >> ~/.bashrc && source ~/.bashrc\n", absTarget)
					default:
						fmt.Printf("  export PATH=\"$PATH:%s\"\n", absTarget)
					}
				}
			}

			cfgPath, err := config.ConfigPath()
			if err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] Cannot determine config path: %v\n", err)
				os.Exit(1)
			}

			cfg, err := config.Load()
			if err != nil {
				fmt.Fprintf(os.Stderr, "[WARN] Could not load config: %v\n", err)
				cfg = config.Defaults()
			}

			if cfg.DeviceID == "" {
				mac, err := device.GetMACAddress()
				if err == nil && mac != "" {
					cfg.DeviceID = "huntcli-" + mac
				} else {
					cfg.DeviceID = fmt.Sprintf("huntcli-%d", time.Now().Unix())
				}
				fmt.Printf("[CONFIG] Generated device ID: %s\n", cfg.DeviceID)
			}

			apiEndpoint := promptInstall("API Endpoint", cfg.APIEndpoint)
			cfg.APIEndpoint = apiEndpoint

			ingestEndpoint := promptInstall("Ingestion Endpoint", cfg.IngestEndpoint)
			cfg.IngestEndpoint = ingestEndpoint

			webClientURL := promptInstall("Web Client URL", cfg.WebClientURL)
			cfg.WebClientURL = webClientURL

			forwardURL := promptInstall("Default Forward URL", cfg.DefaultForward)
			cfg.DefaultForward = forwardURL

			if err := config.Save(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] Failed to save config: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("[CONFIG] Configuration saved to %s.\n", cfgPath)

			if cfg.AuthToken == "" && !skipLogin {
				fmt.Println("\n[STATUS] Device not paired. Initiating pairing sequence...")

				apiClient := client.NewAPIClient(cfg.IngestEndpoint, 30*time.Second)
				pairingResp, err := apiClient.RequestPairingCode(cfg.DeviceID)
				if err != nil {
					fmt.Printf("[WARN] Pairing request failed: %v\n", err)
					fmt.Println("   You can pair later by running 'huntcli login'.")
				} else {
					claimURL := fmt.Sprintf("%s/claim/%s", strings.TrimSuffix(cfg.WebClientURL, "/"), pairingResp.Code)

					fmt.Println("\n==========================================")
					fmt.Printf(" CODE: %s\n", pairingResp.Code)
					fmt.Printf(" URL:  %s\n", claimURL)
					fmt.Println("==========================================")

					qrterminal.GenerateHalfBlock(claimURL, qrterminal.L, os.Stdout)

					fmt.Println("\nWaiting for device to be claimed (Ctrl+C to skip)...")

					ticker := time.NewTicker(5 * time.Second)
					defer ticker.Stop()

					paired := false
				pollLoop:
					for {
						select {
						case <-ticker.C:
							statusResp, err := apiClient.CheckPairingStatus(cfg.DeviceID, pairingResp.Code)
							if err != nil {
								continue
							}
							switch statusResp.Status {
							case client.PairingStatusClaimed:
								fmt.Println("\n[SUCCESS] Device successfully claimed!")
								if statusResp.APIKey != nil {
									cfg.AuthToken = *statusResp.APIKey
									if err := config.Save(cfg); err != nil {
										fmt.Fprintf(os.Stderr, "[ERROR] Failed to save auth token: %v\n", err)
									}
								} else {
									fmt.Println("[WARN] Claimed but no API key received.")
								}
								paired = true
								break pollLoop
							case client.PairingStatusExpired:
								fmt.Println("\n[ERROR] Pairing code expired. Run 'huntcli login' later.")
								break pollLoop
							}
						}
					}

					if !paired {
						fmt.Println("   Proceeding without pairing. Run 'huntcli login' when ready.")
					}
				}
			} else if cfg.AuthToken != "" {
				fmt.Println("[STATUS] Already authenticated. Skipping pairing.")
			} else {
				fmt.Println("[STATUS] Skipping pairing. Run 'huntcli login' when ready.")
			}

			fmt.Println("\nInstallation complete.")
			fmt.Printf("  Binary: %s\n", targetExe)
			fmt.Printf("  Config: %s\n", cfgPath)
			if pathInDir {
				fmt.Println("  Tip:   Run 'huntcli listen' to start receiving events.")
			} else {
				fmt.Println("  Tip:   Add the install directory to your PATH, then run 'huntcli listen'.")
			}
		},
	}

	cmd.Flags().Bool("skip-login", false, "Skip the login prompt after installation")
	cmd.Flags().String("dir", "", "Custom install directory")
	return cmd
}
