import HttpClient from "../hooks/AxiosInstance";
import { UploadProgress } from "../interfaces/common.interface";

export const projectService = {
  uploadProjectFolder: async (
  projectName: string,
  file: Blob,
  onProgress?: (progress: UploadProgress) => void
): Promise<void> => {
  const CHUNK_SIZE = 5 * 1024 * 1024; // 5MB per chunk
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

    try {
      await HttpClient.post("/project/upload-project-folder", formData, {
        headers: { "Content-Type": "multipart/form-data" },
      });

      if (onProgress) {
        onProgress({
          uploadedChunks: i + 1,
          totalChunks,
          percentage: Math.round(((i + 1) / totalChunks) * 100),
        });
      }
    } catch (err) {
      console.error(`Upload failed at chunk ${i}:`, err);
      // Optional: retry the chunk
      // throw err; // or decide to stop
    }
  }
}

};
