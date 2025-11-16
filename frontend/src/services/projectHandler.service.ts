import HttpClient from "../hooks/AxiosInstance";
import { UploadProgress } from "../interfaces/common.interface";

export const projectService = {
  uploadProjectFolder: async (
    projectName: string,
    blob: Blob,
    onProgress?: (progress: UploadProgress) => void
  ): Promise<void> => {
    const CHUNK_SIZE = 5 * 1024 * 1024; // 5MB
    const totalChunks = Math.ceil(blob.size / CHUNK_SIZE);

    for (let i = 0; i < totalChunks; i++) {
      const start = i * CHUNK_SIZE;
      const end = Math.min(start + CHUNK_SIZE, blob.size);
      const chunk = blob.slice(start, end);

      const formData = new FormData();
      formData.append("chunk", chunk);
      formData.append("filename", "folder.zip");
      formData.append("chunkIndex", String(i));
      formData.append("totalChunks", String(totalChunks));
      formData.append("projectName", projectName);

      await HttpClient.post("/project/upload-project-folder", formData, {
        headers: { "Content-Type": "multipart/form-data" },
      });

      // Call progress callback
      if (onProgress) {
        onProgress({
          uploadedChunks: i + 1,
          totalChunks,
          percentage: ((i + 1) / totalChunks) * 100,
        });
      }
    }
  },
};
