// src/services/projectManager.service.ts
import HttpClient from "../hooks/AxiosInstance";
import { UploadProgress } from "../interfaces/common.interface";
import { Project } from "../interfaces/project.interface";

export const projectService = {
  // --- Create project after file upload ---
  InitiateProject: async (
    projectName: string,
    databaseName: string,
    projectFramework: string
  ): Promise<{ success: boolean; message: string; project: Project }> => {
    try {
      const response = await HttpClient.post(`/project/php/init`, {
        domainName: projectName,
        dbName: databaseName,
        projectFramework: projectFramework,
      });

      return {
        success: true,
        message: response.data?.message || "Project created successfully",
        project: response.data?.summary,
      };
    } catch (err: any) {
      console.error("Project creation failed:", err);
      return {
        success: false,
        message: err.response?.data?.message || "Failed to create project",
        project: err.response?.data?.summary,
      };
    }
  },
  // --- Upload project folder in chunks ---
  uploadProjectFolder: async (
    projectID: number,
    projectName: string,
    file: Blob,
    onProgress?: (completed: boolean, progress: UploadProgress) => void,
    retries = 2
  ): Promise<{ success: boolean; message: string }> => {
    const CHUNK_SIZE_IN_MB = 5; // 5MB per chunk
    const CHUNK_SIZE = CHUNK_SIZE_IN_MB * 1024 * 1024;
    const totalChunks = Math.ceil(file.size / CHUNK_SIZE);

    for (let currentChunk = 0; currentChunk < totalChunks; currentChunk++) {
      const start = currentChunk * CHUNK_SIZE;
      const end = Math.min(start + CHUNK_SIZE, file.size);
      const chunk = file.slice(start, end);

      const formData = new FormData();
      formData.append("chunk", chunk);
      formData.append(
        "filename",
        file instanceof File ? file.name : "folder.zip"
      );
      formData.append("chunkIndex", String(currentChunk));
      formData.append("totalChunks", String(totalChunks));
      formData.append("projectID", String(projectID));
      formData.append("projectName", projectName);

      let attempt = 0;
      let uploaded = false;

      while (!uploaded && attempt <= retries) {
        try {
          await HttpClient.post("/project/php/upload-project-file", formData, {
            headers: { "Content-Type": "multipart/form-data" },
          });

          uploaded = true;

          if (onProgress) {
            onProgress(currentChunk === totalChunks - 1, {
              chunkSizeMB: CHUNK_SIZE_IN_MB,
              uploadedChunks: currentChunk + 1,
              totalChunks,
              percentage: Math.round(((currentChunk + 1) / totalChunks) * 100),
            });
          }
        } catch (err: any) {
          attempt++;
          console.error(
            `Chunk ${currentChunk} upload failed (attempt ${attempt}):`,
            err
          );

          if (attempt > retries) {
            return {
              success: false,
              message: `Failed to upload chunk ${currentChunk} after ${
                retries + 1
              } attempts: ${err?.message || err}`,
            };
          }
        }
      }
    }

    return { success: true, message: "Project file uploaded successfully" };
  },

  // --- Deploy PHP Project ---
  DeployProject: async (
    projectID: number,
    projectName: string,
    projectFramework: string
  ): Promise<{ success: boolean; message: string }> => {
    try {
      const formData = new FormData();
      formData.append("projectID", String(projectID));
      formData.append("projectName", projectName);
      formData.append("projectFramework", projectFramework);

      const response = await HttpClient.post(`/project/php/deploy`, formData, {
        headers: { "Content-Type": "multipart/form-data" },
      });

      return {
        success: true,
        message: response.data?.message || "Project deployed successfully",
      };
    } catch (err: any) {
      console.error("Project deployment failed:", err);
      return {
        success: false,
        message: err.response?.data?.message || "Failed to deploy project",
      };
    }
  },
};
