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

type ScanSummary struct {
	Targets        []string `json:"targets"`
	OpenPorts      int      `json:"open_ports"`
	TotalPorts     int      `json:"total_ports"`
	TimeTaken      string   `json:"time_taken"`
	Workers        int      `json:"workers"`
	TimeoutSeconds int      `json:"timeout_seconds"`
}

func worker(wg *sync.WaitGroup, tasks chan int, results chan ScanResult, target string, timeout time.Duration, mutex *sync.Mutex, progress *int) {
	defer wg.Done()
	for port := range tasks {
		result := ScanResult{Target: target, Port: port}
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(target, strconv.Itoa(port)), timeout)
		if err == nil {
			result.Open = true
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
	target := flag.String("target", "scanme.nmap.org", "Target hostname or IP")
	targets := flag.String("targets", "", "Comma-separated list of targets")
	startPort := flag.Int("start", 1, "Starting port")
	endPort := flag.Int("end", 1024, "Ending port")
	workers := flag.Int("workers", 100, "Number of concurrent workers")
	timeout := flag.Int("timeout", 5, "Connection timeout in seconds")
	portsList := flag.String("ports", "", "Comma-separated list of ports")
	jsonOutput := flag.Bool("json", false, "Output results in JSON format")
	flag.Parse()

	var targetsToScan []string
	if *targets != "" {
		targetsToScan = strings.Split(*targets, ",")
	} else {
		targetsToScan = []string{*target}
	}

	var portsToScan []int
	if *portsList != "" {
		portStrs := strings.Split(*portsList, ",")
		for _, p := range portStrs {
			port, err := strconv.Atoi(p)
			if err != nil {
				fmt.Printf("Invalid port: %s\n", p)
				continue
			}
			if port < 1 || port > 65535 {
				fmt.Printf("Port %d out of range (1-65535)\n", port)
				continue
			}
			portsToScan = append(portsToScan, port)
		}
	} else {
		if *startPort < 1 || *endPort > 65535 || *startPort > *endPort {
			fmt.Println("Invalid port range. Use 1-65535 with start <= end")
			return
		}
		for port := *startPort; port <= *endPort; port++ {
			portsToScan = append(portsToScan, port)
		}
	}

	startTime := time.Now()
	var allResults []ScanResult
	openPorts := 0

	for _, target := range targetsToScan {
		totalPorts := len(portsToScan)
		tasks := make(chan int, *workers)
		results := make(chan ScanResult, totalPorts)
		var wg sync.WaitGroup
		var mutex sync.Mutex
		progress := 0

		for i := 0; i < *workers; i++ {
			wg.Add(1)
			go worker(&wg, tasks, results, target, time.Duration(*timeout)*time.Second, &mutex, &progress)
		}

		go func() {
			for _, port := range portsToScan {
				tasks <- port
			}
			close(tasks)
		}()

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
		fmt.Printf("\n%s: %d open ports found\n", target, len(scanResults))
	}

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
	} else {
		fmt.Println("\nOpen ports:")
		for _, res := range allResults {
			output := fmt.Sprintf("%s:%d - Open", res.Target, res.Port)
			if res.Banner != "" {
				output += fmt.Sprintf(" (Banner: %s)", res.Banner)
			}
			fmt.Println(output)
		}

		fmt.Println("\nScan Summary:")
		fmt.Printf("Targets: %v\n", targetsToScan)
		fmt.Printf("Open ports: %d\n", openPorts)
		fmt.Printf("Total ports scanned: %d\n", len(portsToScan))
		fmt.Printf("Time taken: %.2f seconds\n", time.Since(startTime).Seconds())
		fmt.Printf("Workers: %d\n", *workers)
		fmt.Printf("Timeout: %d seconds\n", *timeout)
	}
}