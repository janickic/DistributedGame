package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("No server IP given, would you like to start a new game? (y/n)")
		var newGame string
		fmt.Scanf("%s", &newGame)
		if newGame == "y" {
			startServerMode()
		} else {
			fmt.Println("Please insert server IP (x:x:x:x)")
			var ip string
			fmt.Scanf("%s", &ip)
			ip = validateIP(ip)
			startClientMode(ip)
		}

	} else {
		ip := os.Args[1]
		fmt.Println(ip)
		ip = validateIP(ip)
		fmt.Printf("Connecting to server at IP %s\n", ip)
		startClientMode(ip)
	}
}

func validateIP(ip string) string {
	validIP := net.ParseIP(ip)
	for validIP == nil {
		fmt.Println("Please insert server IP in valid format (x:x:x:x)")
		var ip string
		fmt.Scanf("%s", &ip)
		validIP = net.ParseIP(ip)
	}
	return validIP.String()
}
