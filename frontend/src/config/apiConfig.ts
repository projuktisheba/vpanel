// src/config/apiConfig.ts

// Centralized API configuration
// Default to localhost if VITE_API_BASE_URL is not defined
export const API_BASE_URL: string =
  import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8888/api/v1";
  
  // Default to localhost if API_WEB_SOCKET_URL is not defined
export const WEB_SOCKET_URL: string =
  import.meta.env.VITE_WEB_SOCKET_URL ?? "ws://localhost:8889";
