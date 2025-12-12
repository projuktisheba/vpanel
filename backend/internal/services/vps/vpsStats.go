package vps

import (
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type ServerStats struct {
	
	Timestamp   int64   `json:"timestamp"`
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsed  uint64  `json:"memory_used"`
	MemoryTotal uint64  `json:"memory_total"`
	MemoryPerc  float64 `json:"memory_perc"`
	DiskUsed    uint64  `json:"disk_used"`
	DiskTotal   uint64  `json:"disk_total"`
	DiskPerc    float64 `json:"disk_perc"`
	NetworkRx   uint64  `json:"network_rx"`
	NetworkTx   uint64  `json:"network_tx"`
	LoadAvg1    float64 `json:"load_avg_1"`
	LoadAvg5    float64 `json:"load_avg_5"`
	LoadAvg15   float64 `json:"load_avg_15"`
	Uptime      uint64  `json:"uptime"`
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
	broadcast = make(chan ServerStats)

	prevRx   uint64
	prevTx   uint64
	prevTime time.Time
	netMu    sync.Mutex
)

func RunWebsocketServer() {
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/api/stats", handleHTTPStats)

	go handleBroadcast()
	go collectStats()

	log.Println("Server started on :8889")
	log.Fatal(http.ListenAndServe(":8889", nil))

}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()
	log.Println("Client connected")

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			clientsMu.Lock()
			delete(clients, conn)
			clientsMu.Unlock()
			log.Println("Client disconnected")
			break
		}
	}

}

func handleHTTPStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	stats := collectCurrentStats()
	json.NewEncoder(w).Encode(stats)

}

func handleBroadcast() {
	for stats := range broadcast {
		clientsMu.Lock()
		for client := range clients {
			if err := client.WriteJSON(stats); err != nil {
				log.Println("Write error:", err)
				client.Close()
				delete(clients, client)
			}
		}
		clientsMu.Unlock()
	}
}

func collectStats() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stats := collectCurrentStats()
		broadcast <- stats
	}

}

func collectCurrentStats() ServerStats {
	return ServerStats{
		Timestamp:   time.Now().Unix(),
		CPUUsage:    getCPUUsage(),
		MemoryUsed:  getMemoryUsed(),
		MemoryTotal: getMemoryTotal(),
		MemoryPerc:  getMemoryPercent(),
		DiskUsed:    getDiskUsed(),
		DiskTotal:   getDiskTotal(),
		DiskPerc:    getDiskPercent(),
		NetworkRx:   getNetworkRxSpeed(),
		NetworkTx:   getNetworkTxSpeed(),
		LoadAvg1:    getLoadAvg1(),
		LoadAvg5:    getLoadAvg5(),
		LoadAvg15:   getLoadAvg15(),
		Uptime:      getUptime(),
	}
}


func getCPUUsage() float64 {
	if runtime.GOOS != "linux" {
		return 0.0
	}

	cmd := exec.Command("sh", "-c", "top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\\([0-9.]*\\)%* id.*/\\1/' | awk '{print 100 - $1}'")
	output, err := cmd.Output()
	if err != nil {
		return 0.0
	}

	usage, _ := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	return usage
}

func getMemoryUsed() uint64 {
	if runtime.GOOS != "linux" {
		return 0
	}
	cmd := exec.Command("sh", "-c", "free -b | grep Mem | awk '{print $3}'")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	used, _ := strconv.ParseUint(strings.TrimSpace(string(output)), 10, 64)
	return used
}

func getMemoryTotal() uint64 {
	if runtime.GOOS != "linux" {
		return 0
	}
	cmd := exec.Command("sh", "-c", "free -b | grep Mem | awk '{print $2}'")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	total, _ := strconv.ParseUint(strings.TrimSpace(string(output)), 10, 64)
	return total
}

func getMemoryPercent() float64 {
	total := getMemoryTotal()
	if total == 0 {
		return 0.0
	}
	return float64(getMemoryUsed()) / float64(total) * 100
}

func getDiskUsed() uint64 {
	if runtime.GOOS != "linux" {
		return 0
	}
	cmd := exec.Command("sh", "-c", "df / | tail -1 | awk '{print $3}'")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	used, _ := strconv.ParseUint(strings.TrimSpace(string(output)), 10, 64)
	return used * 1024
}

func getDiskTotal() uint64 {
	if runtime.GOOS != "linux" {
		return 0
	}
	cmd := exec.Command("sh", "-c", "df / | tail -1 | awk '{print $2}'")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	total, _ := strconv.ParseUint(strings.TrimSpace(string(output)), 10, 64)
	return total * 1024
}

func getDiskPercent() float64 {
	if runtime.GOOS != "linux" {
		return 0.0
	}
	cmd := exec.Command("sh", "-c", "df / | tail -1 | awk '{print $5}' | sed 's/%//'")
	output, err := cmd.Output()
	if err != nil {
		return 0.0
	}
	perc, _ := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	return perc
}

// --- NETWORK FUNCTIONS WITH SPEED ---
func getNetworkRx() uint64 {
	if runtime.GOOS != "linux" {
		return 0
	}
	cmd := exec.Command("sh", "-c", `for iface in /sys/class/net/*; do
if [ "$(basename $iface)" != "lo" ]; then
    cat $iface/statistics/rx_bytes
    break
fi
done`)
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	rx, _ := strconv.ParseUint(strings.TrimSpace(string(output)), 10, 64)
	return rx
}

func getNetworkTx() uint64 {
	if runtime.GOOS != "linux" {
		return 0
	}
	cmd := exec.Command("sh", "-c", `for iface in /sys/class/net/*; do
if [ "$(basename $iface)" != "lo" ]; then
    cat $iface/statistics/tx_bytes
    break
fi
done`)
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	tx, _ := strconv.ParseUint(strings.TrimSpace(string(output)), 10, 64)
	return tx
}

func getNetworkRxSpeed() uint64 {
	netMu.Lock()
	defer netMu.Unlock()
	currentRx := getNetworkRx()
	now := time.Now()
	var speed uint64
	if !prevTime.IsZero() {
		interval := now.Sub(prevTime).Seconds()
		if interval > 0 {
			speed = uint64(float64(currentRx-prevRx) / interval)
		}
	}
	prevRx = currentRx
	prevTime = now
	return speed
}

func getNetworkTxSpeed() uint64 {
	netMu.Lock()
	defer netMu.Unlock()
	currentTx := getNetworkTx()
	now := time.Now()
	var speed uint64
	if !prevTime.IsZero() {
		interval := now.Sub(prevTime).Seconds()
		if interval > 0 {
			speed = uint64(float64(currentTx-prevTx) / interval)
		}
	}
	prevTx = currentTx
	prevTime = now
	return speed
}

// --- LOAD AND UPTIME ---
func getLoadAvg1() float64 {
	if runtime.GOOS != "linux" {
		return 0.0
	}
	cmd := exec.Command("sh", "-c", "cat /proc/loadavg | awk '{print $1}'")
	output, err := cmd.Output()
	if err != nil {
		return 0.0
	}
	load, _ := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	return load
}

func getLoadAvg5() float64 {
	if runtime.GOOS != "linux" {
		return 0.0
	}
	cmd := exec.Command("sh", "-c", "cat /proc/loadavg | awk '{print $2}'")
	output, err := cmd.Output()
	if err != nil {
		return 0.0
	}
	load, _ := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	return load
}

func getLoadAvg15() float64 {
	if runtime.GOOS != "linux" {
		return 0.0
	}
	cmd := exec.Command("sh", "-c", "cat /proc/loadavg | awk '{print $3}'")
	output, err := cmd.Output()
	if err != nil {
		return 0.0
	}
	load, _ := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	return load
}

func getUptime() uint64 {
	if runtime.GOOS != "linux" {
		return 0
	}
	cmd := exec.Command("sh", "-c", "cat /proc/uptime | awk '{print $1}' | cut -d. -f1")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	uptime, _ := strconv.ParseUint(strings.TrimSpace(string(output)), 10, 64)
	return uptime
}
