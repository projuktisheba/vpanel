import React, { useState } from "react";
import { UploadProgress } from "../../interfaces/common.interface";
import { projectService } from "../../services/projectHandler.service";

export default function ProjectUploader() {
  const [file, setFile] = useState<File | null>(null);
  const [projectName, setProjectName] = useState<string>("");
  const [uploading, setUploading] = useState(false);
  const [progress, setProgress] = useState<UploadProgress | null>(null);

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const selected = e.target.files?.[0] || null;
    setFile(selected);
  };

  const handleUpload = async () => {
    if (!file || !projectName) {
      alert("Please select a ZIP file and enter a project name.");
      return;
    }

    setUploading(true);
    setProgress({ uploadedChunks: 0, totalChunks: 0, percentage: 0 });

    try {
      await projectService.uploadProjectFolder(projectName, file, (p) => {
        setProgress(p);
      });

      alert("Project folder uploaded successfully!");
    } catch (err) {
      console.error(err);
      alert("Upload failed. Check console.");
    } finally {
      setUploading(false);
      setProgress(null);
    }
  };

  return (
    <div className="p-4 border rounded w-full max-w-md mx-auto">
      <input
        type="text"
        placeholder="Enter project name"
        value={projectName}
        onChange={(e) => setProjectName(e.target.value)}
        className="mb-2 p-1 border rounded w-full"
      />

      {/* Select only zip file */}
      <input
        type="file"
        accept=".zip"
        onChange={handleFileSelect}
        className="mb-2 w-full"
      />

      <button
        onClick={handleUpload}
        disabled={uploading || !file}
        className="px-4 py-2 bg-blue-500 text-white rounded w-full"
      >
        {uploading ? "Uploading..." : "Upload ZIP"}
      </button>

      {uploading && progress && (
        <div className="mt-2">
          <p>
            Uploaded {progress.uploadedChunks}/{progress.totalChunks} chunks (
            {progress.percentage.toFixed(2)}%)
          </p>
          <div className="w-full bg-gray-200 h-3 rounded">
            <div
              className="bg-blue-500 h-3 rounded"
              style={{ width: `${progress.percentage}%` }}
            />
          </div>
        </div>
      )}
    </div>
  );
}
