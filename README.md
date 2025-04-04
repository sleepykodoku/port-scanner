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
./portscanner -target scanme.nmap.org
Advanced scan:
./portscanner -targets "scanme.nmap.org,example.com" -ports "22,80,443,8080" -workers 200 -timeout 3 -json

🎨 Sample Output
Regular Output

Scanning scanme.nmap.org: 1024/1024 (100.0%)
scanme.nmap.org: 2 open ports found

Open ports:
scanme.nmap.org:22 - Open (Banner: SSH-2.0-OpenSSH_7.6p1 Ubuntu-4ubuntu0.3)
scanme.nmap.org:80 - Open (Banner: HTTP/1.1 400 Bad Request...)

Scan completed in 4.21s
Total open ports found: 2
JSON Output

[
  {
    "target": "scanme.nmap.org",
    "port": 22,
    "open": true,
    "banner": "SSH-2.0-OpenSSH_7.6p1 Ubuntu-4ubuntu0.3"
  },
  {
    "target": "scanme.nmap.org",
    "port": 80,
    "open": true,
    "banner": "HTTP/1.1 400 Bad Request..."
  }
]