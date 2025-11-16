import React, { useState } from "react";
import { UploadProgress } from "../../interfaces/common.interface";
import { projectService } from "../../services/projectHandler.service";
import { FileIcon } from "../../icons";
import ComponentCard from "../../components/common/ComponentCard";
import PageBreadcrumb from "../../components/common/PageBreadCrumb";
import PageMeta from "../../components/common/PageMeta";
import Label from "../../components/form/Label";
import Input from "../../components/form/input/InputField";
import Button from "../../components/ui/button/Button";
import FileInput from "../../components/form/input/FileInput";
import { ArrowRight, Loader } from "lucide-react";

export default function ProjectUploader() {
  const [file, setFile] = useState<File | null>(null);
  const [fileError, setFileError] = useState<string>("");
  const [projectName, setProjectName] = useState<string>("");
  const [projectNameError, setProjectNameError] = useState<string>("");
  const [uploading, setUploading] = useState(false);
  const [progress, setProgress] = useState<UploadProgress | null>(null);
  const [uploadSuccess, setUploadSuccess] = useState<string>("");
  const [uploadError, setUploadError] = useState<string>("");

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const selected = e.target.files?.[0] || null;
    setFile(selected);
    console.log(file);
  };

  const handleUpload = async () => {
    if (!projectName) {
      setProjectNameError("Please enter a valid domain name.");
      return;
    }
    if (!file) {
      setFileError("Please select a ZIP file.");
      return;
    }

    setUploading(true);
    setProgress({
      chunkSizeMB: 0,
      uploadedChunks: 0,
      totalChunks: 0,
      percentage: 0,
    });

    try {
      await projectService.uploadProjectFolder(projectName, file, (p) => {
        setProgress(p);
      });

      setUploadSuccess("Project folder uploaded successfully!");
    } catch (err) {
      console.error(err);
      setUploadError("Upload failed. Check console.");
    } finally {
      setUploading(false);
      setProgress(null);
    }
  };

  return (
    <>
      <PageMeta
        title="PHP Site Builder"
        description="Manage & Manage PHP Website"
      />
      <PageBreadcrumb pageTitle="Build PHP Website" />
      <ComponentCard>
        <div className="grid grid-cols-1 sm:grid-cols-1 md:grid-cols-2 lg:grid-cols-2 xl:grid-cols-2 gap-6">
          {/* Column 1: Domain Name */}
          <div className="w-full">
            <Label>
              Domain Name <span className="text-red-700 font-medium"> *</span>
            </Label>
            <Input
              placeholder="Enter domain name(e.g. example.abc)"
              type="text"
              value={projectName}
              onChange={(e) => setProjectName(e.target.value)}
              className="w-full"
            />
            {projectNameError && (
              <div className="text-red-600 text-sm font-medium">
                {projectNameError}
              </div>
            )}
          </div>

          {/* Column 2: Project File */}
          <div className="w-full">
            <Label>
              Project File <span className="text-red-700 font-medium"> *</span>
            </Label>
            <FileInput
              accept=".zip"
              onChange={handleFileSelect}
              className="w-full"
            />
            {fileError && (
              <div className="text-red-600 text-sm font-medium">
                {fileError}
              </div>
            )}
          </div>

          {/* Upload button + progress bar spans both columns */}
          <div className="col-span-1 sm:col-span-1 md:col-span-2 lg:col-span-2 xl:col-span-2 flex flex-col gap-4 mt-2">
            <Button
              disabled={uploading || !file || !projectName}
              size="sm"
              variant="primary"
              onClick={() => handleUpload()}
              endIcon={<ArrowRight />}
            >
              {uploading ? (
                <>
                  <Loader className="animate-spin w-4 h-4 mr-2" />
                  Uploading...
                </>
              ) : (
                "Upload"
              )}
            </Button>

            {/* Uploading progress bar */}
            {uploading && progress && (
              <div className="m-0">
                <div className="mb-2 flex justify-between items-center">
                  <div className="flex items-center gap-x-3">
                    <span className="w-6 h-6 flex justify-center items-center bg-gray-100 border border-blue-200 text-gray-800 rounded">
                      <FileIcon />
                    </span>
                    <div>
                      <p className="text-xs dark:text-gray-50">
                        {progress.uploadedChunks * progress.chunkSizeMB} MB of{" "}
                        {progress.totalChunks * progress.chunkSizeMB} MB
                      </p>
                    </div>
                  </div>
                  <div className="text-sm font-medium text-gray-800 dark:text-gray-50">
                    {progress.percentage.toFixed(1)}%
                  </div>
                </div>

                <div className="w-full bg-gray-200 dark:bg-gray-700 h-4 rounded-full overflow-hidden shadow-inner">
                  <div
                    className="bg-blue-500 h-4 rounded-full transition-all duration-500 ease-out"
                    style={{ width: `${progress.percentage}%` }}
                  ></div>
                </div>
              </div>
            )}

            {/* Final Status */}
            {uploadError && (
              <div className="text-center text-red-600 text-sm font-medium">
                {uploadError}
              </div>
            )}
            {uploadSuccess && (
              <div className="text-center text-green-600 text-sm font-medium">
                {uploadSuccess}
              </div>
            )}
          </div>
        </div>
      </ComponentCard>
    </>
  );
}
