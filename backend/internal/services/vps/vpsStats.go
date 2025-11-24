package vps

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/host"
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
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
	broadcast = make(chan ServerStats)
	
	// State variables needed for calculating delta-based stats
	prevNetStats net.IOCountersStat
    lastNetTime time.Time
)

func RunWebsocketServer() {
	// Initialize previous network stats for delta calculation.
	netStats, _ := net.IOCounters(false)
	if len(netStats) > 0 {
		prevNetStats = netStats[0]
	}
    // Set initial time for network rate calculation
    lastNetTime = time.Now()

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

	stats := collectCurrentStatsGopsutil()
	json.NewEncoder(w).Encode(stats)
}

func handleBroadcast() {
	for {
		stats := <-broadcast

		clientsMu.Lock()
		for client := range clients {
			err := client.WriteJSON(stats)
			if err != nil {
				log.Println("Write error:", err)
				client.Close()
				delete(clients, client)
			}
		}
		clientsMu.Unlock()
	}
}

func collectStats() {
	// The interval dictates how frequently new stats are collected
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stats := collectCurrentStatsGopsutil()
		broadcast <- stats
	}
}

// --- RESOURCE-OPTIMIZED STATS COLLECTION (Fixed for Accuracy) ---
func collectCurrentStatsGopsutil() ServerStats {
	var stats ServerStats
	stats.Timestamp = time.Now().Unix()

	var wg sync.WaitGroup

	// Helper function for concurrent data collection
	collect := func(f func() error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := f(); err != nil {
				log.Printf("Error collecting stat: %v", err) 
			}
		}()
	}

	// 1. CPU Usage: Block for 1 second to get accurate rate
	collect(func() error {
		// This blocks for 1 second, calculating the average CPU usage during that concurrent time.
		percentages, err := cpu.Percent(time.Second, false) 
		if err != nil {
			return err
		}
		if len(percentages) > 0 {
			stats.CPUUsage = percentages[0]
		}
		return nil
	})

	// 2. Memory Stats
	collect(func() error {
		v, err := mem.VirtualMemory()
		if err != nil {
			return err
		}
		stats.MemoryUsed = v.Used
		stats.MemoryTotal = v.Total
		stats.MemoryPerc = v.UsedPercent
		return nil
	})

	// 3. Disk Stats
	collect(func() error {
		d, err := disk.Usage("/")
		if err != nil {
			return err
		}
		stats.DiskUsed = d.Used
		stats.DiskTotal = d.Total
		stats.DiskPerc = d.UsedPercent
		return nil
	})

	// 4. Load Averages
	collect(func() error {
		l, err := load.Avg()
		if err != nil {
			return err
		}
		stats.LoadAvg1 = l.Load1
		stats.LoadAvg5 = l.Load5
		stats.LoadAvg15 = l.Load15
		return nil
	})
	
	// 5. Uptime
	collect(func() error {
		u, err := host.Uptime()
		if err != nil {
			return err
		}
		stats.Uptime = u
		return nil
	})
	
	// 6. Network: Use actual time delta for accurate Bytes/Second rate
	wg.Add(1)
	go func() {
		defer wg.Done()
        
        // Calculate the time difference since the last successful collection
        now := time.Now()
        timeDelta := now.Sub(lastNetTime).Seconds()
        lastNetTime = now // Update global state for next cycle

		netStats, err := net.IOCounters(false) 
		if err != nil {
			log.Printf("Error collecting network stats: %v", err)
			return
		}
		
		if len(netStats) == 0 || timeDelta == 0 {
			return
		}

		currentNetStats := netStats[0]
		
		if prevNetStats.BytesRecv > 0 || prevNetStats.BytesSent > 0 {
            // Calculate total bytes transferred
            rxBytesDelta := currentNetStats.BytesRecv - prevNetStats.BytesRecv
			txBytesDelta := currentNetStats.BytesSent - prevNetStats.BytesSent
            
            // Normalize by timeDelta (Bytes / Seconds)
            stats.NetworkRx = uint64(float64(rxBytesDelta) / timeDelta)
            stats.NetworkTx = uint64(float64(txBytesDelta) / timeDelta)

		} else {
			stats.NetworkRx = 0
			stats.NetworkTx = 0
		}
		
		prevNetStats = currentNetStats
	}()

	wg.Wait()

	return stats
}