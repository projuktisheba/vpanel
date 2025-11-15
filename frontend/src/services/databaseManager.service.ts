import HttpClient from "../hooks/AxiosInstance";
import { Response } from "../interfaces/common.interface";
import { DatabaseResponse } from "../interfaces/databaseManager.interface";

export const databaseManager = {
  // Create MySQL database
  createMySQLDB: async (
    databaseName: string,
    databaseUserName: string
  ): Promise<DatabaseResponse> => {
    try {
      const response = await HttpClient.post<DatabaseResponse>(
        "/db/mysql/create-database",
        {
          database_name: databaseName,
          database_user: databaseUserName,
        }
      );

      // Backend response is always JSON, even for errors
      const data = response.data;

      // Optionally throw if error is true, or just return the object
      if (data?.error) {
        console.warn("Backend returned an error:", data.message);
        // You can either throw here or just return
        // throw new Error(data.message);
      }

      return data;
    } catch (err: any) {
      // Axios error: check if response exists
      if (err.response?.data) {
        const data = err.response.data;
        // data should be your JSON object
        console.error("Error response from backend:", data);
        return {
          id:0,
          error:data.error,
          message:data.message,
        }
      } else {
        console.error("Network or Axios error:", err.message);
        throw new Error(err.message || "Unknown error");
      }
    }
  },

  // Delete MySQL database
  deleteMySQLDB: async (databaseName: string): Promise<Response> => {
    try {
      const response = await HttpClient.delete<DatabaseResponse>(
        `/db/mysql/delete-database?db_name=${databaseName}`
      );

      console.log(response.data);
      return response.data;
    } catch (error: any) {
      console.error(
        "Error deleting MySQL database:",
        error.response?.data || error.message
      );
      throw new Error(
        error.response?.data?.message || "Database creation failed"
      );
    }
  },

  // Import MySQL database
  importMySQLDB: async (formData: FormData): Promise<Response> => {
    try {
      const response = await HttpClient.post<DatabaseResponse>(
        "/db/mysql/import-database",
        formData,
        {
          headers: {
            "Content-Type": "multipart/form-data",
          },
        }
      );

      console.log("Import MySQL database response:", response.data);
      return response.data;
    } catch (error: any) {
      console.error(
        "Error importing MySQL database:",
        error.response?.data || error.message
      );
      throw new Error(
        error.response?.data?.message || "Failed to import MySQL database"
      );
    }
  },

  listMySQLDB: async (): Promise<any> => {
    try {
      const response = await HttpClient.get("db/mysql/databases");
      console.log(response.data);
      return response.data; // return the full response
    } catch (error: any) {
      console.error(
        "Error listing MySQL database:",
        error.response?.data || error.message
      );
      throw new Error(error.response?.data?.message || "Something went wrong");
    }
  },

  // Create MySQL user
  createMySQLUser: async (
    username: string,
    password: string
  ): Promise<DatabaseResponse> => {
    try {
      const response = await HttpClient.post<DatabaseResponse>(
        "/db/mysql/create-user",
        {
          username,
          password,
        }
      );

      console.log("Create MySQL user response:", response.data);
      return response.data;
    } catch (error: any) {
      console.error(
        "Error creating MySQL user:",
        error.response?.data || error.message
      );
      throw new Error(
        error.response?.data?.message || "Failed to create MySQL user"
      );
    }
  },

  listMySQLUsers: async (): Promise<any> => {
    try {
      const response = await HttpClient.get("db/mysql/users");
      console.log(response.data);
      return response.data; // return the full response
    } catch (error: any) {
      console.error(
        "Error listing MySQL users:",
        error.response?.data || error.message
      );
      throw new Error(error.response?.data?.message || "Something went wrong");
    }
  },
};
