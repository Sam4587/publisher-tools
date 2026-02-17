// Package main provides a CLI interface for the publisher tools
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	fmt.Println("Publisher Tools CLI")
	flag.Parse()
	os.Exit(0)
}
