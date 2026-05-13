package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var (
	updateVersion string
	updateDryRun  bool
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Self-update crackerbox-manager to the latest version",
	Long: `Downloads and replaces the current crackerbox-manager binary with a newer version
from the official GitHub releases.

Examples:
  crackerbox-manager update                    # Update to latest stable
  crackerbox-manager update --version 0.2.0    # Update to specific version
  crackerbox-manager update --dry-run          # Check for updates without installing`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repo := "devsprithvi/crakerbox"
		binaryName := "crackerbox-manager"
		arch := runtime.GOARCH // amd64 or arm64

		fmt.Printf("🔄 Checking for updates...\n")
		fmt.Printf("   Current version: %s\n\n", Version)

		// Resolve target version
		targetTag := updateVersion
		if targetTag == "" {
			tag, err := resolveLatestTag(repo)
			if err != nil {
				return fmt.Errorf("failed to resolve latest version: %w", err)
			}
			targetTag = tag
		}

		// Normalise tag
		if !strings.HasPrefix(targetTag, "v") && !strings.Contains(targetTag, "-") {
			targetTag = "v" + targetTag
		}

		fmt.Printf("   Target version:  %s\n\n", targetTag)

		if updateDryRun {
			fmt.Println("   (dry-run mode — no changes made)")
			return nil
		}

		// Download
		url := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s-linux-%s", repo, targetTag, binaryName, arch)
		fmt.Printf("   Downloading from: %s\n", url)

		execPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("could not determine executable path: %w", err)
		}

		resp, err := http.Get(url) //nolint:gosec
		if err != nil {
			return fmt.Errorf("download failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("download failed: HTTP %d — check that tag '%s' exists at https://github.com/%s/releases",
				resp.StatusCode, targetTag, repo)
		}

		// Write to temp file next to current binary
		tmpPath := execPath + ".update"
		out, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}

		if _, err := io.Copy(out, resp.Body); err != nil {
			out.Close()
			os.Remove(tmpPath)
			return fmt.Errorf("download interrupted: %w", err)
		}
		out.Close()

		// Atomic replace
		if err := os.Rename(tmpPath, execPath); err != nil {
			os.Remove(tmpPath)
			return fmt.Errorf("failed to replace binary: %w (try running with sudo)", err)
		}

		fmt.Printf("\n   ✅ Updated crackerbox-manager to %s\n", targetTag)
		fmt.Println("   Restart the manager to use the new version.")
		return nil
	},
}

// resolveLatestTag queries the GitHub API for the latest release tag.
// Falls back to the dev-latest pre-release if no stable release exists.
func resolveLatestTag(repo string) (string, error) {
	apiBase := fmt.Sprintf("https://api.github.com/repos/%s/releases", repo)

	// Try latest stable
	tag, err := fetchTagFromURL(apiBase + "/latest")
	if err == nil && tag != "" {
		return tag, nil
	}

	// Fallback: dev-latest pre-release
	tag, err = fetchTagFromURL(apiBase + "/tags/dev-latest")
	if err == nil && tag != "" {
		return tag, nil
	}

	return "", fmt.Errorf("no releases found at https://github.com/%s/releases", repo)
}

func fetchTagFromURL(url string) (string, error) {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Simple JSON parse for "tag_name": "..."
	bodyStr := string(body)
	idx := strings.Index(bodyStr, `"tag_name"`)
	if idx < 0 {
		return "", fmt.Errorf("tag_name not found")
	}
	rest := bodyStr[idx:]
	start := strings.Index(rest, `"`) + 1
	rest = rest[start:]
	start = strings.Index(rest, `"`) + 1
	rest = rest[start:]
	start = strings.Index(rest, `"`) + 1
	rest = rest[start:]
	end := strings.Index(rest, `"`)
	if end < 0 {
		return "", fmt.Errorf("malformed tag_name")
	}
	return rest[:end], nil
}

func init() {
	updateCmd.Flags().StringVar(&updateVersion, "version", "", "Target version to update to (default: latest)")
	updateCmd.Flags().BoolVar(&updateDryRun, "dry-run", false, "Check for updates without installing")
	rootCmd.AddCommand(updateCmd)
}
