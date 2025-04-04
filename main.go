// Filename: main.go
// Purpose: This program demonstrates how to create a TCP network connection using Go

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ScanResult struct {
	Target string `json:"target"`
	Port   int    `json:"port"`
	Open   bool   `json:"open"`
	Banner string `json:"banner,omitempty"`
}

func worker(wg *sync.WaitGroup, tasks chan int, results chan ScanResult, target string, timeout time.Duration, mutex *sync.Mutex, progress *int) {
	defer wg.Done()
	for port := range tasks {
		result := ScanResult{Target: target, Port: port}
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(target, strconv.Itoa(port)), timeout)
		if err == nil {
			result.Open = true
			// Banner grabbing
			conn.SetReadDeadline(time.Now().Add(timeout))
			banner := make([]byte, 1024)
			n, _ := conn.Read(banner)
			if n > 0 {
				result.Banner = strings.TrimSpace(string(banner[:n]))
			}
			conn.Close()
		}
		results <- result
		mutex.Lock()
		*progress++
		mutex.Unlock()
	}
}

func main() {
	// CLI flags
	target := flag.String("target", "scanme.nmap.org", "Target hostname or IP")
	targets := flag.String("targets", "", "Comma-separated list of targets")
	startPort := flag.Int("start", 1, "Starting port")
	endPort := flag.Int("end", 1024, "Ending port")
	workers := flag.Int("workers", 100, "Number of concurrent workers")
	timeout := flag.Int("timeout", 5, "Connection timeout in seconds")
	jsonOutput := flag.Bool("json", false, "Output results in JSON format")
	flag.Parse()

	// Prepare targets
	var targetsToScan []string
	if *targets != "" {
		targetsToScan = strings.Split(*targets, ",")
	} else {
		targetsToScan = []string{*target}
	}

	// Validate port range
	if *startPort < 1 || *endPort > 65535 || *startPort > *endPort {
		fmt.Println("Invalid port range. Use 1-65535 with start <= end")
		return
	}

	startTime := time.Now()
	var allResults []ScanResult

	for _, target := range targetsToScan {
		// Setup for each target
		totalPorts := *endPort - *startPort + 1
		tasks := make(chan int, *workers)
		results := make(chan ScanResult, totalPorts)
		var wg sync.WaitGroup
		var mutex sync.Mutex
		progress := 0
		openPorts := 0

		// Start workers
		for i := 0; i < *workers; i++ {
			wg.Add(1)
			go worker(&wg, tasks, results, target, time.Duration(*timeout)*time.Second, &mutex, &progress)
		}

		// Progress monitor
		go func() {
			for {
				mutex.Lock()
				current := progress
				mutex.Unlock()
				fmt.Printf("\rScanning %s: %d/%d (%.1f%%)", target, current, totalPorts, float64(current)/float64(totalPorts)*100)
				if current >= totalPorts {
					return
				}
				time.Sleep(200 * time.Millisecond)
			}
		}()

		// Generate tasks
		go func() {
			for port := *startPort; port <= *endPort; port++ {
				tasks <- port
			}
			close(tasks)
		}()

		// Process results
		var scanResults []ScanResult
		for i := 0; i < totalPorts; i++ {
			result := <-results
			if result.Open {
				openPorts++
				scanResults = append(scanResults, result)
			}
		}
		wg.Wait()

		allResults = append(allResults, scanResults...)
		fmt.Printf("\n%s: %d open ports found\n", target, openPorts)
	}

	// Output results
	if *jsonOutput {
		jsonData, err := json.MarshalIndent(allResults, "", "  ")
		if err != nil {
			fmt.Println("Error generating JSON output:", err)
		} else {
			fmt.Println(string(jsonData))
		}
	} else {
		fmt.Println("\nOpen ports:")
		for _, res := range allResults {
			output := fmt.Sprintf("%s:%d - Open", res.Target, res.Port)
			if res.Banner != "" {
				output += fmt.Sprintf(" (Banner: %s)", res.Banner)
			}
			fmt.Println(output)
		}
	}

	// Summary
	fmt.Printf("\nScan completed in %v\n", time.Since(startTime))
	fmt.Printf("Total open ports found: %d\n", len(allResults))
}