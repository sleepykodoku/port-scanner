# 🚀 TCP Port Scanner in Go 🐹

This program demonstrates how to create a TCP network connection using Go.

## ✨ Features

- 🎯 Scan single or multiple targets
- 🔢 Custom port ranges or specific ports
- 🚦 Concurrent workers for fast scanning
- 🎨 Banner grabbing capability
- 📊 JSON output option
- 📈 Real-time progress tracking
- 🐹 Written in pure Go (no dependencies!)

🏃 Running
Basic scan:
go run main.go -target scanme.nmap.org
Advanced scan:
go run main.go -targets "scanme.nmap.org,example.com" -ports "22,80,443,8080" -workers 200 -timeout 3 -json

🎨 Sample Output
Regular Output

Scan Summary:
Targets: [scanme.nmap.org]
Open ports: 2
Total ports scanned: 1024
Time taken: 5.07 seconds
Workers: 100
Timeout: 5 seconds

JSON Output

[
 {
  "results": [
    {
      "target": "scanme.nmap.org",
      "port": 22,
      "open": true,
      "banner": "SSH-2.0-OpenSSH_6.6.1p1 Ubuntu-2ubuntu2.13"
    },
    {
      "target": "scanme.nmap.org",
      "port": 80,
      "open": true
    }
  ],
  "summary": {
    "targets": [
      "scanme.nmap.org"
    ],
    "open_ports": 2,
    "total_ports": 1024,
    "time_taken": "5.07 seconds",
    "workers": 100,
    "timeout_seconds": 5
  }
}
]