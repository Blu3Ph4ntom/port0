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

	// Firewall rules (netsh)
	// NOTE: Use rule names without spaces to avoid quoting/escaping issues.
	fmt.Println("Configuring firewall...")

	type rule struct {
		display string
		args    []string
	}

	rules := []rule{
		{
			display: "HTTP (TCP 80)",
			args: []string{
				"advfirewall", "firewall", "add", "rule",
				"name=port0-http",
				"dir=in",
				"action=allow",
				"protocol=tcp",
				"localport=80",
			},
		},
		{
			display: "DNS (UDP 53)",
			args: []string{
				"advfirewall", "firewall", "add", "rule",
				"name=port0-dns-udp",
				"dir=in",
				"action=allow",
				"protocol=udp",
				"localport=53",
			},
		},
		{
			display: "DNS (TCP 53)",
			args: []string{
				"advfirewall", "firewall", "add", "rule",
				"name=port0-dns-tcp",
				"dir=in",
				"action=allow",
				"protocol=tcp",
				"localport=53",
			},
		},
	}

	for _, r := range rules {
		fmt.Printf("  - %s...\n", r.display)
		out, err := runNetsh(r.args...)
		if err != nil {
			// netsh returns non-zero for a bunch of cases; treat "already exists" as OK.
			if isNetshAlreadyExists(out) {
				fmt.Printf("    ✓ already configured\n")
				continue
			}
			fmt.Printf("    Warning: failed (%v)\n", err)
			if strings.TrimSpace(out) != "" {
				fmt.Printf("    Output: %s\n", strings.TrimSpace(out))
			}
			fmt.Println("    Tip: ensure you are running an elevated (Administrator) terminal.")
			continue
		}

		// netsh success usually prints "Ok."
		fmt.Printf("    ✓ configured\n")
		if strings.TrimSpace(out) != "" && strings.TrimSpace(out) != "Ok." {
			fmt.Printf("    Output: %s\n", strings.TrimSpace(out))
		}
	}

	// NRPT (Name Resolution Policy Table) rules for .web and .local
	// These require admin and may not be available on all Windows editions.
	fmt.Println()
	fmt.Println("Configuring DNS NRPT rules (.web, .local)...")

	// Use a non-interactive PowerShell invocation and make it idempotent.
	psScript := `
$ErrorActionPreference = "Stop"

function Has-Cmdlet($name) {
  return [bool](Get-Command $name -ErrorAction SilentlyContinue)
}

if (-not (Has-Cmdlet "Add-DnsClientNrptRule")) {
  Write-Output "NRPT cmdlets not available on this system. Skipping .web/.local configuration."
  exit 0
}

function Ensure-Nrpt($ns) {
  if (-not (Has-Cmdlet "Get-DnsClientNrptRule")) {
    Write-Output ("NRPT query cmdlet not available. Skipping " + $ns)
    return
  }

  $existing = Get-DnsClientNrptRule -ErrorAction SilentlyContinue | Where-Object { $_.Namespace -eq $ns }
  if ($null -ne $existing) {
    Write-Output ("OK: NRPT already exists for " + $ns)
    return
  }

  Add-DnsClientNrptRule -Namespace $ns -NameServers "127.0.0.1" | Out-Null
  Write-Output ("OK: NRPT added for " + $ns + " -> 127.0.0.1")
}

Ensure-Nrpt ".web"
Ensure-Nrpt ".local"
`
	out, err := runPowerShell(psScript)
	if err != nil {
		fmt.Printf("  Warning: NRPT configuration failed (%v)\n", err)
		if strings.TrimSpace(out) != "" {
			fmt.Printf("  Output: %s\n", strings.TrimSpace(out))
		}
		fmt.Println("  Tip: NRPT cmdlets vary by Windows version/edition; see output above.")
	} else {
		if strings.TrimSpace(out) != "" {
			for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
				fmt.Printf("  %s\n", strings.TrimSpace(line))
			}
		} else {
			fmt.Println("  ✓ NRPT configured")
		}
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Setup complete! Your TLD options:")
	fmt.Println()
	fmt.Println("  *.localhost  → Works natively in Chrome/Firefox/Edge (STABLE)")
	fmt.Println("  *.web        → Resolved via NRPT (if configured)")
	fmt.Println("  *.local      → Resolved via NRPT (if configured)")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	return nil
}

func Teardown() error {
	fmt.Println("port0 teardown for Windows")
	fmt.Println()

	fmt.Println("Removing firewall rules...")
	firewallRules := []string{"port0-http", "port0-dns-udp", "port0-dns-tcp"}
	for _, name := range firewallRules {
		_, _ = runNetsh("advfirewall", "firewall", "delete", "rule", "name="+name)
	}
	fmt.Println("  ✓ firewall rules removed (best-effort)")

	fmt.Println("Removing NRPT rules...")
	psScript := `
$ErrorActionPreference = "SilentlyContinue"

function Has-Cmdlet($name) {
  return [bool](Get-Command $name -ErrorAction SilentlyContinue)
}

if (-not (Has-Cmdlet "Get-DnsClientNrptRule")) {
  exit 0
}

Get-DnsClientNrptRule | Where-Object { $_.Namespace -eq ".web" -or $_.Namespace -eq ".local" } | Remove-DnsClientNrptRule -Force
Write-Output "OK: NRPT rules removed for .web/.local (if present)"
`
	out, err := runPowerShell(psScript)
	if err != nil {
		fmt.Printf("  Warning: NRPT removal failed (%v)\n", err)
		if strings.TrimSpace(out) != "" {
			fmt.Printf("  Output: %s\n", strings.TrimSpace(out))
		}
	} else if strings.TrimSpace(out) != "" {
		fmt.Printf("  %s\n", strings.TrimSpace(out))
	} else {
		fmt.Println("  ✓ NRPT rules removed (best-effort)")
	}

	fmt.Println()
	fmt.Println("Teardown complete!")
	return nil
}

func runNetsh(args ...string) (string, error) {
	cmd := exec.Command("netsh", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func runPowerShell(script string) (string, error) {
	// Use powershell.exe explicitly, and run non-interactive to avoid hangs/prompts.
	// Keep it as a single -Command string.
	cmd := exec.Command(
		"powershell.exe",
		"-NoProfile",
		"-NonInteractive",
		"-ExecutionPolicy", "Bypass",
		"-Command", script,
	)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func isNetshAlreadyExists(output string) bool {
	s := strings.ToLower(output)
	// netsh messages vary; cover common cases.
	if strings.Contains(s, "already exists") {
		return true
	}
	if strings.Contains(s, "an object with that name already exists") {
		return true
	}
	return false
}

// AddHostsEntry is deprecated - NRPT rules are used instead
// Kept for compatibility
func AddHostsEntry(domain string) error { return nil }

// RemoveHostsEntry is deprecated - NRPT rules are used instead
// Kept for compatibility
func RemoveHostsEntry(domain string) error { return nil }

// GetHostsFilePath is deprecated
// Kept for compatibility
func GetHostsFilePath() string { return "" }
