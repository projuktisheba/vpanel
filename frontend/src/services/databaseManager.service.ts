import axios, { AxiosInstance } from "axios";
import { DatabaseResponse } from "../interfaces/databaseManager.interface";
import { tokenUtils } from "../utils/tokenUtils"; // optional — if you use JWT for protected routes
import { API_BASE_URL } from "../config/apiConfig";

const api: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    "Content-Type": "application/json",
  },
  withCredentials: true, // keep if your backend uses cookies or refresh tokens
});

export const databaseManager = {
  // Create MySQL database
  createMySQLDB: async (
    databaseName: string,
    databaseUserName: string
  ): Promise<DatabaseResponse> => {
    try {
      const accessToken = tokenUtils.getAccessToken(); // optional
      const response = await api.post<DatabaseResponse>(
        "/db/mysql/create-database",
        {
          database_name: databaseName,
          database_user: databaseUserName,
        },
        {
          headers: {
            Authorization: `Bearer ${accessToken}`,
          },
        }
      );

      console.log(response.data);
      return response.data;
    } catch (error: any) {
      console.error("Error creating MySQL database:", error.response?.data || error.message);
      throw new Error(error.response?.data?.message || "Database creation failed");
    }
  },
};
