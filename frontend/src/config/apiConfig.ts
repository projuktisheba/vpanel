// src/config/apiConfig.ts

// Centralized API configuration
// Default to localhost if VITE_API_BASE_URL is not defined
export const API_BASE_URL: string =
  import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8888/api/v1";
