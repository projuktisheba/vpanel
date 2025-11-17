// src/services/projectManager.service.ts
import HttpClient from "../hooks/AxiosInstance";
import { UploadProgress } from "../interfaces/common.interface";

export const projectService = {
  // --- Upload project folder in chunks ---
  uploadProjectFolder: async (
    projectName: string,
    projectFramework: string,
    file: Blob,
    onProgress?: (progress: UploadProgress) => void,
    retries = 2
  ): Promise<void> => {
    const CHUNK_SIZE_IN_MB = 5; // 5MB per chunk
    const CHUNK_SIZE = CHUNK_SIZE_IN_MB * 1024 * 1024;
    const totalChunks = Math.ceil(file.size / CHUNK_SIZE);

    for (let i = 0; i < totalChunks; i++) {
      const start = i * CHUNK_SIZE;
      const end = Math.min(start + CHUNK_SIZE, file.size);
      const chunk = file.slice(start, end);

      const formData = new FormData();
      formData.append("chunk", chunk);
      formData.append("filename", file instanceof File ? file.name : "folder.zip");
      formData.append("chunkIndex", String(i));
      formData.append("totalChunks", String(totalChunks));
      formData.append("projectName", projectName);
      formData.append("projectFramework", projectFramework);

      let attempt = 0;
      let uploaded = false;

      while (!uploaded && attempt <= retries) {
        try {
          await HttpClient.post("/project/upload-project-folder", formData, {
            headers: { "Content-Type": "multipart/form-data" },
          });

          uploaded = true;

          if (onProgress) {
            onProgress({
              chunkSizeMB: CHUNK_SIZE_IN_MB,
              uploadedChunks: i + 1,
              totalChunks,
              percentage: Math.round(((i + 1) / totalChunks) * 100),
            });
          }
        } catch (err) {
          attempt++;
          console.error(`Chunk ${i} upload failed (attempt ${attempt}):`, err);

          if (attempt > retries) {
            throw new Error(`Failed to upload chunk ${i} after ${retries + 1} attempts`);
          }
        }
      }
    }
  },

  // --- Create project after file upload ---
  createProject: async (
    projectName: string,
    projectFramework: string,
    databaseName: string,
    filename:string,
  ): Promise<{ success: boolean; message: string }> => {
    try {
      const response = await HttpClient.post(`/project/create?filename=${filename}`, {
        projectName: projectName,
        domainName: projectName,
        projectFramework,
        databaseName,
      });

      return {
        success: true,
        message: response.data?.message || "Project created successfully",
      };
    } catch (err: any) {
      console.error("Project creation failed:", err);
      return {
        success: false,
        message: err.response?.data?.message || "Failed to create project",
      };
    }
  },
};
