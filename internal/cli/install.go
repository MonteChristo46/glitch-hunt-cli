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
				fmt.Printf("\nNote: %s is not in your PATH.\n", absTarget)
				addPath := promptInstall("Add it to your PATH now", "Y")
				if addPath != "n" && addPath != "N" && addPath != "no" && addPath != "NO" {
					shell := os.Getenv("SHELL")
					shell = strings.TrimSpace(shell)
					shellBase := filepath.Base(shell)
					var rcFile string
					home, _ := os.UserHomeDir()
					switch shellBase {
					case "zsh":
						rcFile = filepath.Join(home, ".zshrc")
					case "bash":
						if f, _ := os.Stat(filepath.Join(home, ".bash_profile")); f != nil {
							rcFile = filepath.Join(home, ".bash_profile")
						} else {
							rcFile = filepath.Join(home, ".bashrc")
						}
					default:
						rcFile = filepath.Join(home, ".profile")
					}

					line := fmt.Sprintf("export PATH=\"$PATH:%s\"", absTarget)
					data, _ := os.ReadFile(rcFile)
					if strings.Contains(string(data), line) {
						fmt.Printf("[OK] %s already configured in %s.\n", absTarget, rcFile)
					} else {
						f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_WRONLY, 0644)
						if err == nil {
							fmt.Fprintln(f, "")
							fmt.Fprintln(f, "# Added by huntcli installer")
							fmt.Fprintln(f, line)
							f.Close()
							fmt.Printf("[OK] Added to PATH in %s.\n", rcFile)
							fmt.Printf("     Restart your terminal or run: source %s\n", rcFile)
						}
					}
				} else {
					fmt.Println("")
					fmt.Println("To add it manually, run:")
					fmt.Printf("  export PATH=\"$PATH:%s\"\n", absTarget)
					shell := filepath.Base(os.Getenv("SHELL"))
					switch shell {
					case "zsh":
						fmt.Printf("  echo 'export PATH=\"$PATH:%s\"' >> ~/.zshrc\n", absTarget)
					case "bash":
						fmt.Printf("  echo 'export PATH=\"$PATH:%s\"' >> ~/.bashrc\n", absTarget)
					default:
						fmt.Printf("  echo 'export PATH=\"$PATH:%s\"' >> ~/.profile\n", absTarget)
					}
				}
				fmt.Println("")
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

			_, exeOnly := filepath.Split(targetExe)
			fmt.Printf("\n[OK] Installed to: %s\n\n", targetExe)
			if !pathInDir {
				fmt.Printf("Note: %s is not in your PATH.\n", absTarget)
				fmt.Println("Add it by running:")
				fmt.Printf("  export PATH=\"$PATH:%s\"\n", absTarget)
				shell := filepath.Base(os.Getenv("SHELL"))
				var rcName string
				switch shell {
				case "zsh":
					rcName = "~/.zshrc"
				case "bash":
					if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), ".bash_profile")); err == nil {
						rcName = "~/.bash_profile"
					} else {
						rcName = "~/.bashrc"
					}
				default:
					rcName = "~/.profile"
				}
				fmt.Printf("  echo 'export PATH=\"$PATH:%s\"' >> %s\n", absTarget, rcName)
				fmt.Println("")
			}
			fmt.Printf("Now run '%s install' to complete setup:\n", exeOnly)
			fmt.Printf("  %s install\n", targetExe)
			fmt.Println("")
			fmt.Println("Or authenticate directly:")
			fmt.Printf("  %s login\n", targetExe)
			fmt.Printf("  %s listen --forward-to http://localhost:8080/webhooks\n", targetExe)
		},
	}

	cmd.Flags().Bool("skip-login", false, "Skip the login prompt after installation")
	cmd.Flags().String("dir", "", "Custom install directory")
	return cmd
}
