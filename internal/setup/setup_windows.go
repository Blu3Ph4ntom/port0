//go:build windows

package setup

import "fmt"

func Setup() error {
	fmt.Println("Windows setup is not yet fully supported.")
	fmt.Println()
	fmt.Println("Manual steps:")
	fmt.Println("  1. Run 'port0 daemon start' manually")
	fmt.Println("  2. Add a firewall rule to allow port 80:")
	fmt.Println("     netsh advfirewall firewall add rule name=\"port0\" dir=in action=allow protocol=tcp localport=80")
	fmt.Println("  3. For .web domains, configure your DNS to point *.web to 127.0.0.1")
	fmt.Println()
	fmt.Println("Note: *.localhost works in Chrome/Firefox without any setup.")
	return nil
}

func Teardown() error {
	fmt.Println("Windows teardown:")
	fmt.Println("  1. Stop the daemon: port0 daemon stop")
	fmt.Println("  2. Remove firewall rule:")
	fmt.Println("     netsh advfirewall firewall delete rule name=\"port0\"")
	return nil
}
