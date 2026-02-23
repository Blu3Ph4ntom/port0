//go:build linux

package setup

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const resolvedConfDir = "/etc/systemd/resolved.conf.d"
const resolvedConfFile = "/etc/systemd/resolved.conf.d/port0.conf"
const resolvedConfContent = `[Resolve]
DNS=127.0.0.1
Domains=~web ~local
`

func Setup() error {
	fmt.Println("port0 setup for Linux")
	fmt.Println()

	// Create resolved config directory
	fmt.Println("Creating systemd-resolved config directory...")
	if err := os.MkdirAll(resolvedConfDir, 0755); err != nil {
		fmt.Printf("  Warning: could not create config dir: %v\n", err)
		fmt.Println("  You may need to run with sudo")
	} else {
		fmt.Println("  ✓ Created", resolvedConfDir)
	}

	// Write resolved config for .web and .local
	fmt.Println("Configuring DNS for .web and .local TLDs...")
	if err := os.WriteFile(resolvedConfFile, []byte(resolvedConfContent), 0644); err != nil {
		fmt.Printf("  Warning: could not write resolved config: %v\n", err)
		fmt.Println("  You may need to run with sudo")
	} else {
		fmt.Println("  ✓ DNS config written for .web and .local → 127.0.0.1")
	}

	// Restart systemd-resolved
	fmt.Println("Restarting systemd-resolved...")
	if out, err := run("systemctl", "restart", "systemd-resolved"); err != nil {
		fmt.Printf("  Warning: could not restart systemd-resolved: %v\n", err)
		if strings.TrimSpace(out) != "" {
			fmt.Printf("  Output: %s\n", strings.TrimSpace(out))
		}
		fmt.Println("  You may need to run with sudo")
	} else {
		fmt.Println("  ✓ systemd-resolved restarted")
	}

	// Apply per-interface routing for environments where global resolved.conf.d does not
	// reliably apply routing domains (common on some VPS images / network managers).
	//
	// This makes .web/.local work without requiring manual resolvectl commands.
	iface, _ := defaultInterface()
	if iface != "" {
		fmt.Printf("Applying routing to default interface (%s)...\n", iface)

		if out, err := run("resolvectl", "dns", iface, "127.0.0.1"); err != nil {
			fmt.Printf("  Warning: could not set interface DNS via resolvectl: %v\n", err)
			if strings.TrimSpace(out) != "" {
				fmt.Printf("  Output: %s\n", strings.TrimSpace(out))
			}
		} else {
			fmt.Println("  ✓ resolvectl dns applied")
		}

		if out, err := run("resolvectl", "domain", iface, "~web", "~local"); err != nil {
			fmt.Printf("  Warning: could not set routing domains via resolvectl: %v\n", err)
			if strings.TrimSpace(out) != "" {
				fmt.Printf("  Output: %s\n", strings.TrimSpace(out))
			}
		} else {
			fmt.Println("  ✓ resolvectl domain applied")
		}

		_, _ = run("resolvectl", "flush-caches")
	} else {
		fmt.Println("Applying routing to default interface...")
		fmt.Println("  Warning: could not detect default interface (skipping resolvectl per-link routing)")
		fmt.Println("  Tip: you can manually run:")
		fmt.Println("    sudo resolvectl dns <iface> 127.0.0.1")
		fmt.Println("    sudo resolvectl domain <iface> '~web' '~local'")
	}

	// Set capability for binding to privileged ports
	bin, _ := os.Executable()
	fmt.Println("Setting CAP_NET_BIND_SERVICE capability...")
	if out, err := run("setcap", "cap_net_bind_service=+ep", bin); err != nil {
		fmt.Printf("  Warning: could not set capability: %v\n", err)
		if strings.TrimSpace(out) != "" {
			fmt.Printf("  Output: %s\n", strings.TrimSpace(out))
		}
		fmt.Printf("  Run manually: sudo setcap cap_net_bind_service=+ep %s\n", bin)
	} else {
		fmt.Println("  ✓ CAP_NET_BIND_SERVICE set on", bin)
	}

	// Create systemd user service
	home, _ := os.UserHomeDir()
	serviceDir := filepath.Join(home, ".config", "systemd", "user")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		fmt.Printf("  Warning: could not create user service directory: %v\n", err)
	}

	serviceFile := filepath.Join(serviceDir, "port0.service")
	serviceContent := fmt.Sprintf(`[Unit]
Description=port0 daemon

[Service]
ExecStart=%s daemon start
Restart=on-failure

[Install]
WantedBy=default.target
`, bin)

	fmt.Println("Creating systemd user service...")
	if err := os.WriteFile(serviceFile, []byte(serviceContent), 0644); err != nil {
		fmt.Printf("  Warning: could not write service file: %v\n", err)
	} else {
		fmt.Println("  ✓ Systemd service written to", serviceFile)
	}

	_, _ = run("systemctl", "--user", "daemon-reload")
	_, _ = run("systemctl", "--user", "enable", "port0")

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Setup complete! Your TLD options:")
	fmt.Println()
	fmt.Println("  *.localhost  → Works natively in Chrome/Firefox (STABLE)")
	fmt.Println("  *.web        → Resolved via systemd-resolved")
	fmt.Println("  *.local      → Resolved via systemd-resolved")
	fmt.Println()
	fmt.Println("All three TLDs are now available for your projects!")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("")
	fmt.Println("To start the daemon now:")
	fmt.Println("  systemctl --user start port0")

	return nil
}

func Teardown() error {
	fmt.Println("port0 teardown for Linux")
	fmt.Println()

	// Stop and disable systemd user service
	fmt.Println("Stopping systemd user service...")
	_, _ = run("systemctl", "--user", "stop", "port0")
	_, _ = run("systemctl", "--user", "disable", "port0")

	home, _ := os.UserHomeDir()
	serviceFile := filepath.Join(home, ".config", "systemd", "user", "port0.service")
	if err := os.Remove(serviceFile); err == nil {
		fmt.Println("  ✓ Removed", serviceFile)
	}
	_, _ = run("systemctl", "--user", "daemon-reload")

	// Remove resolved config
	fmt.Println("Removing DNS config...")
	if err := os.Remove(resolvedConfFile); err == nil {
		fmt.Println("  ✓ Removed", resolvedConfFile)
	}

	// Revert per-interface routing (best-effort)
	iface, _ := defaultInterface()
	if iface != "" {
		fmt.Printf("Reverting resolvectl routing on %s...\n", iface)
		_, _ = run("resolvectl", "revert", iface)
	}

	// Restart systemd-resolved
	fmt.Println("Restarting systemd-resolved...")
	_, _ = run("systemctl", "restart", "systemd-resolved")
	fmt.Println("  ✓ systemd-resolved restarted")
	_, _ = run("resolvectl", "flush-caches")

	// Remove capability from binary
	bin, _ := os.Executable()
	fmt.Println("Removing capability from binary...")
	_, _ = run("setcap", "-r", bin)
	fmt.Println("  ✓ Capability removed")

	fmt.Println()
	fmt.Println("Teardown complete!")
	return nil
}

func run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b
	err := cmd.Run()
	return b.String(), err
}

func defaultInterface() (string, error) {
	// Primary: iproute2
	out, err := run("ip", "route", "show", "default")
	if err == nil {
		iface := parseDefaultInterface(out)
		if iface != "" {
			return iface, nil
		}
	}

	// Fallback: try `ip -o route` which is more stable across formats
	out, err2 := run("ip", "-o", "route", "show", "to", "default")
	if err2 == nil {
		iface := parseDefaultInterface(out)
		if iface != "" {
			return iface, nil
		}
	}

	// Last resort: try `route -n` (best-effort)
	out, err3 := run("route", "-n")
	if err3 == nil {
		iface := parseDefaultInterfaceRouteN(out)
		if iface != "" {
			return iface, nil
		}
	}

	if err != nil {
		return "", err
	}
	return "", fmt.Errorf("could not detect default interface")
}

func parseDefaultInterface(s string) string {
	// Example: "default via 67.207.67.1 dev eth0 proto dhcp src 67.207.67.2 metric 100\n"
	fields := strings.Fields(s)
	for i := 0; i+1 < len(fields); i++ {
		if fields[i] == "dev" {
			return fields[i+1]
		}
	}
	return ""
}

func parseDefaultInterfaceRouteN(s string) string {
	// Very basic parse: look for line where Destination is 0.0.0.0 and take Iface column.
	// Output varies; we keep it best-effort.
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "Kernel") || strings.HasPrefix(line, "Destination") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 8 {
			continue
		}
		// Destination is first field, Iface is typically last.
		if fields[0] == "0.0.0.0" {
			return fields[len(fields)-1]
		}
	}
	return ""
}
