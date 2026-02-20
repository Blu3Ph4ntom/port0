//go:build darwin

package setup

import (
	"fmt"
	"os"
	"os/exec"
)

const resolverDir = "/etc/resolver"
const resolverFile = "/etc/resolver/web"
const resolverContent = `nameserver 127.0.0.1
port 53
`

const launchdPlistPath = "/Library/LaunchDaemons/com.port0.daemon.plist"

func launchdPlist() string {
	bin, _ := os.Executable()
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.port0.daemon</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
        <string>daemon</string>
        <string>start</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/port0-daemon.stdout.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/port0-daemon.stderr.log</string>
</dict>
</plist>
`, bin)
}

func Setup() error {
	if err := os.MkdirAll(resolverDir, 0755); err != nil {
		return fmt.Errorf("setup: create resolver dir: %w", err)
	}

	if err := os.WriteFile(resolverFile, []byte(resolverContent), 0644); err != nil {
		return fmt.Errorf("setup: write resolver: %w", err)
	}

	plist := launchdPlist()
	if err := os.WriteFile(launchdPlistPath, []byte(plist), 0644); err != nil {
		return fmt.Errorf("setup: write plist: %w", err)
	}

	fmt.Println("Setup complete.")
	fmt.Println("  Resolver for .web written to /etc/resolver/web")
	fmt.Println("  LaunchDaemon plist written to", launchdPlistPath)
	fmt.Println()
	fmt.Println("To start the daemon now:")
	fmt.Println("  sudo launchctl load", launchdPlistPath)

	return nil
}

func Teardown() error {
	os.Remove(resolverFile)
	exec.Command("launchctl", "unload", launchdPlistPath).Run()
	os.Remove(launchdPlistPath)
	fmt.Println("Teardown complete. Resolver and LaunchDaemon removed.")
	return nil
}
