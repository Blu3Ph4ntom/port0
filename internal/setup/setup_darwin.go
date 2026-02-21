//go:build darwin

package setup

import (
	"fmt"
	"os"
	"os/exec"
)

const resolverDir = "/etc/resolver"
const resolverFileWeb = "/etc/resolver/web"
const resolverFileLocal = "/etc/resolver/local"
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
	fmt.Println("port0 setup for macOS")
	fmt.Println()

	// Create resolver directory
	fmt.Println("Creating resolver directory...")
	if err := os.MkdirAll(resolverDir, 0755); err != nil {
		fmt.Printf("  Warning: could not create resolver dir: %v\n", err)
		fmt.Println("  You may need to run with sudo")
	} else {
		fmt.Println("  ✓ Created /etc/resolver")
	}

	// Create resolver for .web
	fmt.Println("Configuring DNS for .web TLD...")
	if err := os.WriteFile(resolverFileWeb, []byte(resolverContent), 0644); err != nil {
		fmt.Printf("  Warning: could not write .web resolver: %v\n", err)
		fmt.Println("  You may need to run with sudo")
	} else {
		fmt.Println("  ✓ Resolver created for .web → 127.0.0.1")
	}

	// Create resolver for .local
	fmt.Println("Configuring DNS for .local TLD...")
	if err := os.WriteFile(resolverFileLocal, []byte(resolverContent), 0644); err != nil {
		fmt.Printf("  Warning: could not write .local resolver: %v\n", err)
		fmt.Println("  You may need to run with sudo")
	} else {
		fmt.Println("  ✓ Resolver created for .local → 127.0.0.1")
	}

	// Create launchd plist
	fmt.Println("Creating LaunchDaemon...")
	plist := launchdPlist()
	if err := os.WriteFile(launchdPlistPath, []byte(plist), 0644); err != nil {
		fmt.Printf("  Warning: could not write plist: %v\n", err)
		fmt.Println("  You may need to run with sudo")
	} else {
		fmt.Println("  ✓ LaunchDaemon plist written to", launchdPlistPath)
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Setup complete! Your TLD options:")
	fmt.Println()
	fmt.Println("  *.localhost  → Works natively in Chrome/Firefox/Safari (STABLE)")
	fmt.Println("  *.web        → Resolved via /etc/resolver/web")
	fmt.Println("  *.local      → Resolved via /etc/resolver/local")
	fmt.Println()
	fmt.Println("All three TLDs are now available for your projects!")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("To start the daemon now:")
	fmt.Println("  sudo launchctl load", launchdPlistPath)

	return nil
}

func Teardown() error {
	fmt.Println("port0 teardown for macOS")
	fmt.Println()

	// Remove resolver files
	fmt.Println("Removing resolver files...")
	if err := os.Remove(resolverFileWeb); err == nil {
		fmt.Println("  ✓ Removed", resolverFileWeb)
	}
	if err := os.Remove(resolverFileLocal); err == nil {
		fmt.Println("  ✓ Removed", resolverFileLocal)
	}

	// Unload and remove launchd plist
	fmt.Println("Removing LaunchDaemon...")
	exec.Command("launchctl", "unload", launchdPlistPath).Run()
	if err := os.Remove(launchdPlistPath); err == nil {
		fmt.Println("  ✓ Removed", launchdPlistPath)
	}

	fmt.Println()
	fmt.Println("Teardown complete!")
	return nil
}
