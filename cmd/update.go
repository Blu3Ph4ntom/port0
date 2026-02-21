package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update port0 to the latest version",
	Long:  "Download and install the latest version of port0 from GitHub releases.",
	RunE:  runUpdate,
}

func runUpdate(cmd *cobra.Command, args []string) error {
	// Detect OS and architecture
	osName := runtime.GOOS
	arch := runtime.GOARCH

	// Map OS names to release naming
	if osName == "darwin" {
		osName = "darwin"
	} else if osName == "linux" {
		osName = "linux"
	} else if osName == "windows" {
		osName = "windows"
	} else {
		return fmt.Errorf("unsupported OS: %s", osName)
	}

	// Map architectures
	if arch == "amd64" || arch == "x86_64" {
		arch = "amd64"
	} else if arch == "arm64" || arch == "aarch64" {
		arch = "arm64"
	} else {
		return fmt.Errorf("unsupported architecture: %s", arch)
	}

	// Build download URL
	ext := ""
	if osName == "windows" {
		ext = ".exe"
	}
	binaryName := fmt.Sprintf("port0-%s-%s%s", osName, arch, ext)
	downloadURL := fmt.Sprintf("https://github.com/blu3ph4ntom/port0/releases/latest/download/%s", binaryName)

	fmt.Printf("Checking for updates...\n")
	fmt.Printf("  OS: %s\n", osName)
	fmt.Printf("  Arch: %s\n", arch)

	// Get current executable path
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine current executable path: %w", err)
	}

	// Download new version
	fmt.Printf("Downloading from %s\n", downloadURL)

	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	// Create temp file
	tempFile := currentExe + ".new"
	out, err := os.Create(tempFile)
	if err != nil {
		// Try with temp dir
		tempFile = fmt.Sprintf("%s%cport0%s.new", os.TempDir(), os.PathSeparator, ext)
		out, err = os.Create(tempFile)
		if err != nil {
			return fmt.Errorf("cannot create temp file: %w", err)
		}
	}

	_, err = io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("download incomplete: %w", err)
	}

	// Make executable
	if err := os.Chmod(tempFile, 0755); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("cannot set permissions: %w", err)
	}

	// Backup old version
	backup := currentExe + ".old"
	os.Remove(backup) // Remove old backup if exists

	// On Windows, we can't replace a running executable directly
	// We need to rename the old one and place the new one
	if runtime.GOOS == "windows" {
		// Rename current to .old
		if err := os.Rename(currentExe, backup); err != nil {
			// If we can't rename, try to write to a different location
			installDir := os.Getenv("USERPROFILE") + "\\bin"
			newPath := installDir + "\\port0.exe"
			if err := os.MkdirAll(installDir, 0755); err != nil {
				os.Remove(tempFile)
				return fmt.Errorf("cannot create install directory: %w", err)
			}
			if err := os.Rename(tempFile, newPath); err != nil {
				os.Remove(tempFile)
				return fmt.Errorf("cannot install update: %w", err)
			}
			os.Remove(backup)
			fmt.Printf("\n✓ Updated successfully to %s\n", newPath)
			fmt.Println("  Restart your terminal to use the new version.")
			return nil
		}
	} else {
		// On Unix, rename current to .old
		if err := os.Rename(currentExe, backup); err != nil {
			// If we can't rename (might be in /usr/local/bin), try with sudo hint
			os.Remove(tempFile)
			return fmt.Errorf("cannot replace binary (try running with sudo): %w", err)
		}
	}

	// Move new version to current location
	if err := os.Rename(tempFile, currentExe); err != nil {
		// Try to restore backup
		os.Rename(backup, currentExe)
		os.Remove(tempFile)
		return fmt.Errorf("cannot install update: %w", err)
	}

	// Clean up backup
	os.Remove(backup)

	fmt.Printf("\n✓ Updated successfully!\n")
	fmt.Println("  Restart your terminal to use the new version.")

	// Show version info
	if version != "" {
		fmt.Printf("  Version: %s\n", version)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
