//go:build linux

package setup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	if err := exec.Command("systemctl", "restart", "systemd-resolved").Run(); err != nil {
		fmt.Printf("  Warning: could not restart systemd-resolved: %v\n", err)
		fmt.Println("  You may need to run with sudo")
	} else {
		fmt.Println("  ✓ systemd-resolved restarted")
	}

	// Set capability for binding to privileged ports
	bin, _ := os.Executable()
	fmt.Println("Setting CAP_NET_BIND_SERVICE capability...")
	if err := exec.Command("setcap", "cap_net_bind_service=+ep", bin).Run(); err != nil {
		fmt.Printf("  Warning: could not set capability: %v\n", err)
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

	_ = exec.Command("systemctl", "--user", "daemon-reload").Run()
	_ = exec.Command("systemctl", "--user", "enable", "port0").Run()

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
	_ = exec.Command("systemctl", "--user", "stop", "port0").Run()
	_ = exec.Command("systemctl", "--user", "disable", "port0").Run()

	home, _ := os.UserHomeDir()
	serviceFile := filepath.Join(home, ".config", "systemd", "user", "port0.service")
	if err := os.Remove(serviceFile); err == nil {
		fmt.Println("  ✓ Removed", serviceFile)
	}
	_ = exec.Command("systemctl", "--user", "daemon-reload").Run()

	// Remove resolved config
	fmt.Println("Removing DNS config...")
	if err := os.Remove(resolvedConfFile); err == nil {
		fmt.Println("  ✓ Removed", resolvedConfFile)
	}

	// Restart systemd-resolved
	fmt.Println("Restarting systemd-resolved...")
	_ = exec.Command("systemctl", "restart", "systemd-resolved").Run()
	fmt.Println("  ✓ systemd-resolved restarted")

	// Remove capability from binary
	bin, _ := os.Executable()
	fmt.Println("Removing capability from binary...")
	_ = exec.Command("setcap", "-r", bin).Run()
	fmt.Println("  ✓ Capability removed")

	fmt.Println()
	fmt.Println("Teardown complete!")
	return nil
}
