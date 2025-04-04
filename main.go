// Filename: main.go
// Purpose: This program demonstrates how to create a TCP network connection using Go

package main

import (
	"flag"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

func worker(wg *sync.WaitGroup, tasks chan int, results chan int, target string, timeout time.Duration) {
	defer wg.Done()
	for port := range tasks {
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(target, strconv.Itoa(port)), timeout)
		if err == nil {
			conn.Close()
			results <- port
		} else {
			results <- 0
		}
	}
}

func main() {
	// CLI flags
	target := flag.String("target", "scanme.nmap.org", "Target hostname or IP")
	startPort := flag.Int("start", 1, "Starting port")
	endPort := flag.Int("end", 1024, "Ending port")
	workers := flag.Int("workers", 100, "Number of concurrent workers")
	timeout := flag.Int("timeout", 5, "Connection timeout in seconds")
	flag.Parse()

	// Validate port range
	if *startPort < 1 || *endPort > 65535 || *startPort > *endPort {
		fmt.Println("Invalid port range. Use 1-65535 with start <= end")
		return
	}

	// Setup
	totalPorts := *endPort - *startPort + 1
	tasks := make(chan int, *workers)
	results := make(chan int, totalPorts)
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go worker(&wg, tasks, results, *target, time.Duration(*timeout)*time.Second)
	}

	// Generate tasks
	go func() {
		for port := *startPort; port <= *endPort; port++ {
			tasks <- port
		}
		close(tasks)
	}()

	// Process results
	openPorts := 0
	for i := 0; i < totalPorts; i++ {
		if port := <-results; port > 0 {
			fmt.Printf("Port %d is open\n", port)
			openPorts++
		}
	}
	wg.Wait()

	// Summary
	fmt.Printf("\nScan completed: %d open ports found\n", openPorts)
	fmt.Printf("Scanned %d ports in total\n", totalPorts)
}