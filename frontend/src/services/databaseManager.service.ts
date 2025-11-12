import HttpClient from "../hooks/AxiosInstance";
import { DatabaseMeta, DatabaseResponse } from "../interfaces/databaseManager.interface";
import { tokenUtils } from "../utils/tokenUtils"; // optional — if you use JWT for protected routes

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

      console.log(response.data);
      return response.data;
    } catch (error: any) {
      console.error(
        "Error creating MySQL database:",
        error.response?.data || error.message
      );
      throw new Error(
        error.response?.data?.message || "Database creation failed"
      );
    }
  },
  listMySQLDB: async (): Promise<any> => {
  try {
    const response = await HttpClient.get("db/mysql/databases");
    console.log(response.data);
    return response.data; // return the full response
  } catch (error: any) {
    console.error("Error listing MySQL database:", error.response?.data || error.message);
    throw new Error(error.response?.data?.message || "Something went wrong");
  }
},

};
