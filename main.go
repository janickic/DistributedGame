package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("No server IP given, would you like to start a new game? (y/n)")
		var newGame string
		fmt.Scanf("%s", &newGame)
		if newGame == "y" {
			fmt.Println("Enter N for board size between 2-6")
			var n string
			var N int
			fmt.Scanf("%s", &n)
			N, _ = strconv.Atoi(n)
			for N < 2 || N > 6 {
				fmt.Println("Enter N for board size between 2-6")
				fmt.Scanf("%s", &n)
				N, _ = strconv.Atoi(n)
			}
			fmt.Println("Enter minFill for cell between 0.1-0.9")
			var min string
			var minFill float32
			fmt.Scanf("%s", &min)
			minFill64, _ := strconv.ParseFloat(min, 32)
			minFill = float32(minFill64)
			for minFill < 0.1 || minFill > 0.9 {
				fmt.Println("Enter minFill for cell between 0.1-0.9")
				fmt.Scanf("%s", &min)
				minFill64, _ := strconv.ParseFloat(min, 32)
				minFill = float32(minFill64)
			}

			startServerMode(N, minFill)
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
