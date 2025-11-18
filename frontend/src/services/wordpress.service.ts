// src/services/projectManager.service.ts
import HttpClient from "../hooks/AxiosInstance";
import { Project } from "../interfaces/project.interface";

export const wordpressService = {
  // --- Create project upload ---
  BuildProject: async (
    databaseName: string,
    domainName: string
  ): Promise<{
    error: boolean;
    message: string;
    summary?: Project;
  }> => {
    try {
      const response = await HttpClient.post(`/project/wordpress/deploy`, {
        dbName: databaseName,
        domainName: domainName,
      });

      return {
        error: false,
        message: response.data?.message || "Project created successfully",
        summary: response.data?.summary,
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
