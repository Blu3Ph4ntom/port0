//go:build linux

package setup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const resolvedConfDir = "/etc/systemd/resolved.conf.d"
const resolvedConfFile = "/etc/systemd/resolved.conf.d/port0-web.conf"
const resolvedConfContent = `[Resolve]
DNS=127.0.0.1
Domains=~web
`

func Setup() error {
	if err := os.MkdirAll(resolvedConfDir, 0755); err != nil {
		return fmt.Errorf("setup: create resolved conf dir: %w", err)
	}

	if err := os.WriteFile(resolvedConfFile, []byte(resolvedConfContent), 0644); err != nil {
		return fmt.Errorf("setup: write resolved conf: %w", err)
	}

	if err := exec.Command("systemctl", "restart", "systemd-resolved").Run(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not restart systemd-resolved: %v\n", err)
	}

	bin, _ := os.Executable()
	if err := exec.Command("setcap", "cap_net_bind_service=+ep", bin).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not set CAP_NET_BIND_SERVICE: %v\n", err)
		fmt.Fprintf(os.Stderr, "  run manually: sudo setcap cap_net_bind_service=+ep %s\n", bin)
	}

	home, _ := os.UserHomeDir()
	serviceDir := filepath.Join(home, ".config", "systemd", "user")
	os.MkdirAll(serviceDir, 0755)

	serviceFile := filepath.Join(serviceDir, "port0.service")
	serviceContent := fmt.Sprintf(`[Unit]
Description=port0 daemon

[Service]
ExecStart=%s daemon start
Restart=on-failure

[Install]
WantedBy=default.target
`, bin)

	if err := os.WriteFile(serviceFile, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("setup: write systemd service: %w", err)
	}

	exec.Command("systemctl", "--user", "daemon-reload").Run()
	exec.Command("systemctl", "--user", "enable", "port0").Run()

	fmt.Println("Setup complete.")
	fmt.Println("  DNS config written to", resolvedConfFile)
	fmt.Println("  Systemd user service installed at", serviceFile)
	fmt.Println()
	fmt.Println("To start the daemon now:")
	fmt.Println("  systemctl --user start port0")

	return nil
}

func Teardown() error {
	exec.Command("systemctl", "--user", "stop", "port0").Run()
	exec.Command("systemctl", "--user", "disable", "port0").Run()

	home, _ := os.UserHomeDir()
	serviceFile := filepath.Join(home, ".config", "systemd", "user", "port0.service")
	os.Remove(serviceFile)
	exec.Command("systemctl", "--user", "daemon-reload").Run()

	os.Remove(resolvedConfFile)
	exec.Command("systemctl", "restart", "systemd-resolved").Run()

	fmt.Println("Teardown complete. DNS config, systemd service, and capability removed.")
	return nil
}
