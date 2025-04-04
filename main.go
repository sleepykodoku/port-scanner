package main  // Declares this as an executable program

import (
	"encoding/json"  // For JSON encoding/decoding
	"flag"           // For command-line flag parsing
	"fmt"            // For formatted I/O
	"net"            // For network operations
	"strconv"        // For string conversions
	"strings"        // For string manipulation
	"sync"           // For synchronization primitives
	"time"           // For time-related operations
)

type ScanResult struct {  // Stores results for each port scan
	Target string `json:"target"`  // Target hostname/IP
	Port   int    `json:"port"`    // Port number
	Open   bool   `json:"open"`    // Whether port is open
	Banner string `json:"banner,omitempty"`  // Service banner if available
}

type ScanSummary struct {  // Stores summary of the scan
	Targets        []string `json:"targets"`         // List of targets scanned
	OpenPorts      int      `json:"open_ports"`      // Count of open ports
	TotalPorts     int      `json:"total_ports"`     // Total ports scanned
	TimeTaken      string   `json:"time_taken"`      // Duration of scan
	Workers        int      `json:"workers"`        // Number of workers used
	TimeoutSeconds int      `json:"timeout_seconds"` // Timeout setting
}

func worker(wg *sync.WaitGroup, tasks chan int, results chan ScanResult,target string, timeout time.Duration, mutex *sync.Mutex, progress *int) {
	defer wg.Done()  // Signal completion when worker exits
	for port := range tasks {  // Process ports from task channel
		result := ScanResult{Target: target, Port: port}
		// Try TCP connection with timeout
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(target, strconv.Itoa(port)), timeout)
		if err == nil {  // If connection succeeded
			result.Open = true
			conn.SetReadDeadline(time.Now().Add(timeout))  // Set read timeout
			banner := make([]byte, 1024)  // Buffer for banner
			n, _ := conn.Read(banner)  // Read initial response
			if n > 0 {
				result.Banner = strings.TrimSpace(string(banner[:n]))  // Store banner
			}
			conn.Close()  // Close connection
		}
		results <- result  // Send result to output channel
		mutex.Lock()  // Safely update progress counter
		*progress++
		mutex.Unlock()
	}
}

func main() {
	// Command-line flag definitions
	target := flag.String("target", "scanme.nmap.org", "Target hostname or IP")
	targets := flag.String("targets", "", "Comma-separated list of targets")
	startPort := flag.Int("start", 1, "Starting port")
	endPort := flag.Int("end", 1024, "Ending port")
	workers := flag.Int("workers", 100, "Number of concurrent workers")
	timeout := flag.Int("timeout", 5, "Connection timeout in seconds")
	portsList := flag.String("ports", "", "Comma-separated list of ports")
	jsonOutput := flag.Bool("json", false, "Output results in JSON format")
	flag.Parse()  // Parse command-line flags

	// Process target list
	var targetsToScan []string
	if *targets != "" {
		targetsToScan = strings.Split(*targets, ",")
	} else {
		targetsToScan = []string{*target}
	}

	// Process port list/range
	var portsToScan []int
	if *portsList != "" {
		portStrs := strings.Split(*portsList, ",")
		for _, p := range portStrs {
			port, err := strconv.Atoi(p)
			if err != nil {
				fmt.Printf("Invalid port: %s\n", p)
				continue
			}
			if port < 1 || port > 102 {
				fmt.Printf("Port %d out of range (1-65535)\n", port)
				continue
			}
			portsToScan = append(portsToScan, port)
		}
	} else {  // Use port range if specific ports not provided
		if *startPort < 1 || *endPort > 65535 || *startPort > *endPort {
			fmt.Println("Invalid port range. Use 1-65535 with start <= end")
			return
		}
		for port := *startPort; port <= *endPort; port++ {
			portsToScan = append(portsToScan, port)
		}
	}

	startTime := time.Now()  // Record start time
	var allResults []ScanResult
	openPorts := 0  // Counter for open ports

	// Process each target
	for _, target := range targetsToScan {
		totalPorts := len(portsToScan)
		tasks := make(chan int, *workers)  // Buffered channel for ports to scan
		results := make(chan ScanResult, totalPorts)  // Buffered channel for results
		var wg sync.WaitGroup  // WaitGroup to track workers
		var mutex sync.Mutex  // Mutex for progress counter
		progress := 0  // Progress counter

		// Launch worker goroutines
		for i := 0; i < *workers; i++ {
			wg.Add(1)
			go worker(&wg, tasks, results, target, time.Duration(*timeout)*time.Second, &mutex, &progress)
		}

		// Feed ports to workers
		go func() {
			for _, port := range portsToScan {
				tasks <- port
			}
			close(tasks)  // Close channel when done
		}()

		// Collect results
		var scanResults []ScanResult
		for i := 0; i < totalPorts; i++ {
			result := <-results
			if result.Open {
				openPorts++
				scanResults = append(scanResults, result)
			}
		}
		wg.Wait()  // Wait for all workers to finish
		allResults = append(allResults, scanResults...)
	}

	// Output results
	if *jsonOutput {
		summary := ScanSummary{
			Targets:        targetsToScan,
			OpenPorts:      openPorts,
			TotalPorts:     len(portsToScan),
			TimeTaken:      fmt.Sprintf("%.2f seconds", time.Since(startTime).Seconds()),
			Workers:        *workers,
			TimeoutSeconds: *timeout,
		}

		output := struct {
			Results []ScanResult `json:"results"`
			Summary ScanSummary  `json:"summary"`
		}{
			Results: allResults,
			Summary: summary,
		}

		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			fmt.Println("Error generating JSON output:", err)
		} else {
			fmt.Println(string(jsonData))
		}
	} else {  // Human-readable output
		fmt.Println("\nScan Summary:")
		fmt.Printf("Targets: %v\n", targetsToScan)
		fmt.Printf("Open ports: %d\n", openPorts)
		fmt.Printf("Total ports scanned: %d\n", len(portsToScan))
		fmt.Printf("Time taken: %.2f seconds\n", time.Since(startTime).Seconds())
		fmt.Printf("Workers: %d\n", *workers)
		fmt.Printf("Timeout: %d seconds\n", *timeout)
	}
}