import HttpClient from "../hooks/AxiosInstance";
import { Response } from "../interfaces/common.interface";

export const domainManager = {
  // add a new domain
  addNewDomain: async (domainName: string, domainNameProvider: string): Promise<Response> => {
    try {
      const response = await HttpClient.post<Response>(
        "/domain/create", // replace with your actual backend endpoint
        {
          domain: domainName,
          domainProvider: domainNameProvider,
        }
      );

      const data = response.data;

      // Log backend errors if any
      if (data?.error) {
        console.warn("Backend returned an error:", data.message);
        // optionally throw here: throw new Error(data.message);
      }

      return data;
    } catch (err: any) {
      if (err.response?.data) {
        const data = err.response.data;
        console.error("Error response from backend:", data);
        return {
          error: data.error,
          message: data.message,
        };
      } else {
        console.error("Network or Axios error:", err.message);
        throw new Error(err.message || "Unknown error");
      }
    }
  },


  listDomains: async (): Promise<any> => {
    try {
      const response = await HttpClient.get("/domain/list");
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
};
