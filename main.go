package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		startServerMode()

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
