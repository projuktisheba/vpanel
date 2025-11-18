// src/services/projectManager.service.ts
import HttpClient from "../hooks/AxiosInstance";

export const wordpressService = {
  // --- Create project upload ---
  BuildProject: async (
    databaseName: string,
    domainName: string
  ): Promise<{ error: boolean; message: string }> => {
    try {
      const response = await HttpClient.post(`/project/wordpress/deploy`, {
        databaseName,
        domainName,
      });

      return {
        error: false,
        message: response.data?.message || "Project created successfully",
      };
    } catch (err: any) {
      console.error("Project creation failed:", err);
      return {
        error: true,
        message: err.response?.data?.message || "Failed to create project",
      };
    }
  },
};
