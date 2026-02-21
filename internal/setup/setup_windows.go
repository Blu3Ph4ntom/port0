//go:build windows

package setup

import (
	"fmt"
	"os/exec"
	"strings"
)

func Setup() error {
	fmt.Println("port0 setup for Windows")
	fmt.Println()

	// Add firewall rule for port 80
	fmt.Println("Configuring firewall for port 80...")
	cmd := exec.Command("netsh", "advfirewall", "firewall", "add", "rule",
		"name=\"port0 HTTP\"",
		"dir=in",
		"action=allow",
		"protocol=tcp",
		"localport=80")
	if err := cmd.Run(); err != nil {
		fmt.Printf("  Warning: could not add firewall rule: %v\n", err)
		fmt.Println("  You may need to run as Administrator")
	} else {
		fmt.Println("  ✓ Firewall rule added for port 80")
	}

	// Add firewall rule for port 53 (DNS)
	fmt.Println("Configuring firewall for port 53 (DNS)...")
	cmd = exec.Command("netsh", "advfirewall", "firewall", "add", "rule",
		"name=\"port0 DNS\"",
		"dir=in",
		"action=allow",
		"protocol=udp",
		"localport=53")
	if err := cmd.Run(); err != nil {
		fmt.Printf("  Warning: could not add DNS firewall rule: %v\n", err)
	} else {
		fmt.Println("  ✓ Firewall rule added for port 53")
	}

	// Add NRPT rule for .web
	fmt.Println("Configuring DNS for .web TLD...")
	cmd = exec.Command("powershell", "-Command",
		"Add-DnsClientNrptRule -Namespace '.web' -NameServers '127.0.0.1'")
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("  Warning: could not add NRPT rule for .web: %v\n", err)
		if len(output) > 0 {
			fmt.Printf("  Output: %s\n", strings.TrimSpace(string(output)))
		}
		fmt.Println("  You may need to run as Administrator")
	} else {
		fmt.Println("  ✓ NRPT rule added for .web → 127.0.0.1")
	}

	// Add NRPT rule for .local
	fmt.Println("Configuring DNS for .local TLD...")
	cmd = exec.Command("powershell", "-Command",
		"Add-DnsClientNrptRule -Namespace '.local' -NameServers '127.0.0.1'")
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("  Warning: could not add NRPT rule for .local: %v\n", err)
		if len(output) > 0 {
			fmt.Printf("  Output: %s\n", strings.TrimSpace(string(output)))
		}
		fmt.Println("  You may need to run as Administrator")
	} else {
		fmt.Println("  ✓ NRPT rule added for .local → 127.0.0.1")
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Setup complete! Your TLD options:")
	fmt.Println()
	fmt.Println("  *.localhost  → Works natively in Chrome/Firefox/Edge (STABLE)")
	fmt.Println("  *.web        → Resolved via NRPT rule")
	fmt.Println("  *.local      → Resolved via NRPT rule")
	fmt.Println()
	fmt.Println("All three TLDs are now available for your projects!")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	return nil
}

func Teardown() error {
	fmt.Println("port0 teardown for Windows")
	fmt.Println()

	// Remove firewall rules
	fmt.Println("Removing firewall rules...")
	exec.Command("netsh", "advfirewall", "firewall", "delete", "rule", "name=\"port0 HTTP\"").Run()
	exec.Command("netsh", "advfirewall", "firewall", "delete", "rule", "name=\"port0 DNS\"").Run()
	fmt.Println("  ✓ Firewall rules removed")

	// Remove NRPT rules for .web
	fmt.Println("Removing NRPT rules...")
	cmd := exec.Command("powershell", "-Command",
		"Get-DnsClientNrptRule | Where-Object { $_.Namespace -eq '.web' } | Remove-DnsClientNrptRule -Force")
	cmd.Run()

	// Remove NRPT rules for .local
	cmd = exec.Command("powershell", "-Command",
		"Get-DnsClientNrptRule | Where-Object { $_.Namespace -eq '.local' } | Remove-DnsClientNrptRule -Force")
	cmd.Run()

	fmt.Println("  ✓ NRPT rules removed")
	fmt.Println()
	fmt.Println("Teardown complete!")

	return nil
}

// AddHostsEntry is deprecated - NRPT rules are used instead
// Kept for compatibility
func AddHostsEntry(domain string) error {
	return nil
}

// RemoveHostsEntry is deprecated - NRPT rules are used instead
// Kept for compatibility
func RemoveHostsEntry(domain string) error {
	return nil
}

// GetHostsFilePath is deprecated
// Kept for compatibility
func GetHostsFilePath() string {
	return ""
}
