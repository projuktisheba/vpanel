import React, { useState, useEffect, useRef } from "react";
import {
  Activity,
  Cpu,
  HardDrive,
  Network,
  Clock,
  TrendingUp,
  Server,
  ArrowDown,
  ArrowUp,
} from "lucide-react";
// Assuming you have an API base URL defined for the REST endpoint
import { API_BASE_URL, WEB_SOCKET_URL } from "../../config/apiConfig"; 

// --- 2. SERVER STATS INTERFACES AND UTILITIES ---

interface ServerStats {
  timestamp: number;
  cpu_usage: number;
  memory_used: number;
  memory_total: number;
  memory_perc: number;
  disk_used: number;
  disk_total: number;
  disk_perc: number;
  network_rx: number;
  network_tx: number;
  load_avg_1: number;
  load_avg_5: number;
  load_avg_15: number;
  uptime: number;
}

const formatBytes = (bytes: number) => {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return (bytes / Math.pow(k, i)).toFixed(2) + " " + sizes[i];
};

const formatUptime = (seconds: number) => {
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const mins = Math.floor((seconds % 3600) / 60);
  return `${days}d ${hours}h ${mins}m`;
};

// --- Stat Card Component (Theme-Aware) ---
// (StatCard component remains unchanged)

const StatCard = ({
  icon: Icon,
  title,
  value,
  subtitle,
  color,
  percentage,
}: {
  icon: React.ElementType;
  title: string;
  value: string;
  subtitle?: string;
  color: string;
  percentage?: number;
}) => (
  // Card styling: bg-white light / bg-gray-800 dark
  <div
    className="rounded-xl border border-gray-200 bg-white p-6 shadow-md transition-all duration-300 hover:shadow-lg 
             dark:border-gray-700 dark:bg-gray-800 dark:shadow-xl dark:hover:shadow-2xl"
  >
    <div className="flex items-center justify-between">
      <div>
        <p className="text-sm font-medium text-gray-500 dark:text-gray-400">
          {title}
        </p>
        <h3 className="mt-1 text-3xl font-extrabold text-gray-900 dark:text-white">
          {value}
        </h3>
        {subtitle && (
          <p className="mt-1 text-xs text-gray-500 dark:text-gray-500">
            {subtitle}
          </p>
        )}
      </div>
      <div
        className="flex h-12 w-12 items-center justify-center rounded-full bg-opacity-20"
        style={{ backgroundColor: `${color}25` }}
      >
        <Icon className="h-6 w-6" style={{ color }} />
      </div>
    </div>

    {/* Progress Bar (Theme-Aware) */}
    {percentage !== undefined && (
      <div className="mt-5">
        <div className="h-2 w-full rounded-full bg-gray-200 dark:bg-gray-700">
          <div
            className="h-2 rounded-full transition-all duration-500"
            style={{ width: `${percentage}%`, backgroundColor: color }}
          />
        </div>
      </div>
    )}

    {/* Placeholder for chart area (Theme-Aware) */}
    {percentage !== undefined && (
      <div className="mt-4 h-10 w-full rounded-lg bg-gray-100 dark:bg-gray-700/50 animate-pulse"></div>
    )}
  </div>
);

// --- 4. MAIN DASHBOARD COMPONENT (Updated) ---

export function ServerStatsDashboard() {
  const [stats, setStats] = useState<ServerStats | null>(null);
  const [connected, setConnected] = useState(false);
  // NEW STATE: To store the server's IP address
  const [ipAddress, setIpAddress] = useState<string>("Fetching IP..."); 
  const ws = useRef<WebSocket | null>(null);

  useEffect(() => {
    connectWebSocket();
    return () => {
      if (ws.current) {
        ws.current.close();
      }
    };
  }, []);
  
  // NEW FUNCTION: To fetch the IP address via REST API
  const fetchIpAddress = async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/ping`);
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const data = await response.json();
      // Assuming the ping API returns an object like { ip: "192.168.1.1" }
      setIpAddress(data.server_ip || "IP Unknown"); 
    } catch (error) {
      console.error("Error fetching IP address:", error);
      setIpAddress("IP Error");
    }
  };

  const connectWebSocket = () => {
    ws.current = new WebSocket(WEB_SOCKET_URL + "/ws");

    ws.current.onopen = () => {
      setConnected(true);
      console.log("WebSocket connected. Fetching IP...");
      // ACTION: Trigger IP fetch on successful connection
      fetchIpAddress(); 
    };

    ws.current.onmessage = (event) => {
      const data = JSON.parse(event.data);
      setStats(data);
    };

    ws.current.onerror = (error) => {
      console.error("WebSocket error:", error);
      setConnected(false);
    };

    ws.current.onclose = () => {
      setConnected(false);
      // Reset IP status on disconnect
      setIpAddress("Disconnected"); 
      console.log("WebSocket disconnected, reconnecting in 3s...");
      setTimeout(connectWebSocket, 3000);
    };
  };

  
  if (!stats) {
    return (
      // Light background default, dark background when 'dark' is present
      <div className="flex h-screen items-center justify-center bg-gray-50 dark:bg-gray-900">
        <div className="text-center">
          <div className="mx-auto mb-4 h-16 w-16 animate-spin rounded-full border-b-4 border-t-4 border-blue-500"></div>
          <p className="text-lg font-medium text-gray-600 dark:text-gray-300">
            Connecting to server...
          </p>
        </div>
      </div>
    );
  }

  return (
    // Outer Container: Sets the overall background theme color
    <div className="max-h-screen bg-gray-50 dark:bg-gray-900 transition-colors duration-500">
      {/* Header (Theme-Aware) */}
      <div className="border-b border-gray-200 bg-white dark:border-gray-700 dark:bg-gray-800">
        <div className="mx-auto max-w-7xl px-4 py-4 sm:px-6 lg:px-8">
          <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <div className="flex items-center gap-3">
              <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-blue-600">
                <Server className="h-6 w-6 text-white" />
              </div>
              <div>
                <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
                  VPS Monitor
                </h1>
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  Real-time server statistics
                </p>
              </div>
            </div>
            {/* Theme Toggle and Status (Updated) */}
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-3 rounded-lg border border-gray-300 bg-gray-100 px-4 py-2 dark:border-gray-700 dark:bg-gray-700/50">
                <div
                  className={`h-2.5 w-2.5 rounded-full ${
                    connected ? "bg-green-500" : "bg-red-500"
                  } ${connected && ipAddress !== "Fetching IP..." ? "animate-pulse" : ""}`}
                ></div>
                {/* DISPLAY IP ADDRESS/STATUS HERE */}
                <span className="text-sm font-medium text-gray-700 dark:text-white">
                  {connected ? ipAddress : "Disconnected"}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Main Content Area */}
      <div className="mx-auto max-w-7xl px-4 py-4 sm:px-6 sm:py-6 lg:px-8">
        {/* Stats Grid */}
        <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-4 mb-8">
          <StatCard
            icon={Cpu}
            title="CPU Usage"
            value={`${stats.cpu_usage.toFixed(1)}%`}
            color="#3b82f6"
            percentage={stats.cpu_usage}
          />

          <StatCard
            icon={Activity}
            title="Memory"
            value={`${stats.memory_perc.toFixed(1)}%`}
            subtitle={`${formatBytes(stats.memory_used)} / ${formatBytes(
              stats.memory_total
            )}`}
            color="#10b981"
            percentage={stats.memory_perc}
          />

          <StatCard
            icon={HardDrive}
            title="Disk Space"
            value={`${stats.disk_perc.toFixed(1)}%`}
            subtitle={`${formatBytes(stats.disk_used)} / ${formatBytes(
              stats.disk_total
            )}`}
            color="#f59e0b"
            percentage={stats.disk_perc}
          />

          <StatCard
            icon={Clock}
            title="Uptime"
            value={formatUptime(stats.uptime).split(" ")[0]}
            subtitle={formatUptime(stats.uptime)}
            color="#8b5cf6"
          />
        </div>

        {/* Network & System Load Grid */}
        <div className="grid grid-cols-1 gap-6 lg:grid-cols-2 mb-8">
          {/* Network Card (Theme-Aware) */}
          <div className="rounded-xl border border-gray-200 bg-white p-6 shadow-md dark:border-gray-700 dark:bg-gray-800">
            <div className="mb-6 flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-purple-600/10 dark:bg-purple-600/20">
                <Network className="h-5 w-5 text-purple-600 dark:text-purple-400" />
              </div>
              <div>
                <h3 className="text-xl font-semibold text-gray-900 dark:text-white">
                  Network Traffic
                </h3>
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  Real-time bandwidth usage (Bytes/s)
                </p>
              </div>
            </div>

            <div className="space-y-6">
              {/* Download (RX) */}
              <div className="rounded-lg bg-gray-100 p-4 dark:bg-gray-700/50">
                <div className="mb-2 flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <ArrowDown className="h-4 w-4 text-purple-600 dark:text-purple-400" />
                    <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                      Download (RX)
                    </span>
                  </div>
                  <span className="text-lg font-bold text-gray-900 dark:text-white">
                    {formatBytes(stats.network_rx)}/s
                  </span>
                </div>
                <div className="h-10 w-full rounded-lg bg-gray-200 dark:bg-gray-700/70 animate-pulse"></div>
              </div>

              {/* Upload (TX) */}
              <div className="rounded-lg bg-gray-100 p-4 dark:bg-gray-700/50">
                <div className="mb-2 flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <ArrowUp className="h-4 w-4 text-pink-600 dark:text-pink-400" />
                    <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                      Upload (TX)
                    </span>
                  </div>
                  <span className="text-lg font-bold text-gray-900 dark:text-white">
                    {formatBytes(stats.network_tx)}/s
                  </span>
                </div>
                <div className="h-10 w-full rounded-lg bg-gray-200 dark:bg-gray-700/70 animate-pulse"></div>
              </div>
            </div>
          </div>

          {/* System Load Card (Theme-Aware) */}
          <div className="rounded-xl border border-gray-200 bg-white p-6 shadow-md dark:border-gray-700 dark:bg-gray-800">
            <div className="mb-6 flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-cyan-600/10 dark:bg-cyan-600/20">
                <TrendingUp className="h-5 w-5 text-cyan-600 dark:text-cyan-400" />
              </div>
              <div>
                <h3 className="text-xl font-semibold text-gray-900 dark:text-white">
                  System Load
                </h3>
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  Average load metrics
                </p>
              </div>
            </div>

            <div className="space-y-4">
              <div className="flex items-center justify-between rounded-lg bg-gray-100 p-4 dark:bg-gray-700">
                <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                  1 Minute Average
                </span>
                <span className="text-2xl font-bold text-gray-900 dark:text-white">
                  {stats.load_avg_1.toFixed(2)}
                </span>
              </div>
              <div className="flex items-center justify-between rounded-lg bg-gray-100 p-4 dark:bg-gray-700">
                <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                  5 Minutes Average
                </span>
                <span className="text-2xl font-bold text-gray-900 dark:text-white">
                  {stats.load_avg_5.toFixed(2)}
                </span>
              </div>
              <div className="flex items-center justify-between rounded-lg bg-gray-100 p-4 dark:bg-gray-700">
                <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                  15 Minutes Average
                </span>
                <span className="text-2xl font-bold text-gray-900 dark:text-white">
                  {stats.load_avg_15.toFixed(2)}
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* Footer (Theme-Aware) */}
        <div className="rounded-xl border border-gray-200 bg-white px-4 py-3 text-center shadow-md dark:border-gray-700 dark:bg-gray-800">
          <p className="text-sm text-gray-500 dark:text-gray-400">
            Server Time:{" "}
            <span className="font-medium text-gray-900 dark:text-white mr-4">
              {new Date(stats.timestamp * 1000).toUTCString()}
            </span>
            Last updated:{" "}
            <span className="font-medium text-gray-900 dark:text-white">
              {new Date(stats.timestamp * 1000).toLocaleTimeString()}
            </span>
          </p>
        </div>
      </div>
    </div>
  );
}