// src/config/apiConfig.ts

// Centralized API configuration
// Default to localhost if VITE_API_BASE_URL is not defined
export const API_BASE_URL: string =
  import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8888/api/v1";

// Example: Optional API key from env
export const API_KEY: string | undefined = import.meta.env.VITE_API_KEY;
