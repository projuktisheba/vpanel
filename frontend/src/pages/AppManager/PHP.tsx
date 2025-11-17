import { useEffect, useState } from "react";
import { UploadProgress } from "../../interfaces/common.interface";
import { projectService } from "../../services/projectManager.service";
import { FileIcon } from "../../icons";
import ComponentCard from "../../components/common/ComponentCard";
import PageBreadcrumb from "../../components/common/PageBreadCrumb";
import PageMeta from "../../components/common/PageMeta";
import Label from "../../components/form/Label";
import Button from "../../components/ui/button/Button";
import { ArrowRight, Loader } from "lucide-react";
import DropzoneComponent from "../../components/form/form-elements/DropZone";
import Select from "../../components/form/Select";
import { Domain } from "../../interfaces/domain.interface";
import { domainManager } from "../../services/domainManager.service";
import { databaseManager } from "../../services/databaseManager.service";
interface Option {
  value: string;
  label: string;
}
export default function ProjectUploader() {
  //project name(domain)
  const [projectName, setProjectName] = useState<string>("");
  const [projectNameError, setProjectNameError] = useState<string>("");
  const [domainList, setDomainList] = useState<Option[]>([]);

  //databases
  const [databaseList, setDatabaseList] = useState<Option[]>([]);
  const [databaseName, setDatabaseName] = useState<string>("");
  const [databaseNameError, setDatabaseNameError] = useState<string>("");

  //framework
  const [projectFramework, setProjectFramework] = useState<string>("");
  const [projectFrameworkError, setProjectFrameworkError] =
    useState<string>("");
  //file
  const [file, setFile] = useState<File | null>(null);
  const [fileError, setFileError] = useState<string>("");

  //status
  const [uploading, setUploading] = useState(false);
  const [progress, setProgress] = useState<UploadProgress | null>(null);
  const [uploadSuccess, setUploadSuccess] = useState<string>("");
  const [uploadError, setUploadError] = useState<string>("");

  const fetchDomains = async () => {
    try {
      const res = await domainManager.listDomains();

      const domainsArray = Array.isArray(res.domains) ? res.domains : [];

      // Use domainsArray, not domains
      const list: Option[] = domainsArray.map((domain: Domain) => ({
        value: domain.domain,
        label: domain.domain,
      }));

      setDomainList(list);
      console.log(list); // now it will log correctly
    } catch (err) {
      console.error("Failed to fetch domains:", err);
    }
  };

  const fetchDatabases = async () => {
    try {
      const res = await databaseManager.listMySQLDB();

      const databaseArray = Array.isArray(res.databases) ? res.databases : [];

      const list: Option[] = databaseArray.map((database: { dbName: any }) => ({
        value: database.dbName,
        label: database.dbName,
      }));

      setDatabaseList(list);
      console.log(list); // now it will log correctly
    } catch (err) {
      console.error("Failed to fetch databases:", err);
    }
  };

  //call backend funcs
  useEffect(() => {
    fetchDomains();
    fetchDatabases();
  }, []);

  // Supported PHP Frameworks
  const frameworks = [
    { value: "Laravel", label: "Laravel" },
    { value: "CodeIgniter", label: "CodeIgniter" },
  ];

  //upload handler
  const handleProjectDeployment = async () => {
    // Step 1: Validate form fields
    if (!projectName) {
      setProjectNameError("Please enter a valid domain name.");
      return;
    }
    if (!projectFramework) {
      setProjectFrameworkError("Please select a valid framework.");
      return;
    }
    if (!databaseName) {
      setDatabaseNameError("Please select a valid database.");
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
    setUploadError("");
    setUploadSuccess("");

    try {
      // Step 2: Upload project folder in chunks
      await projectService.uploadProjectFolder(
        projectName,
        projectFramework,
        file,
        (p) => setProgress(p)
      );

      const filename = file instanceof File ? file.name : "folder.zip";
      // Step 3: Call project creation service with form values (not the file)
      const createResp = await projectService.createProject(
        projectName,
        projectFramework,
        databaseName,
        filename,
      );

      // Step 4: Show API response
      if (createResp.success) {
        setUploadSuccess(createResp.message);
      } else {
        setUploadError(createResp.message);
      }
    } catch (err) {
      console.error(err);
      setUploadError("Upload or project creation failed. Check console.");
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
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Left Column: Domain, Framework, Database */}
          <div className="flex flex-col gap-4">
            {/* Domain Name */}
            <div className="w-full">
              <Label>
                Domain <span className="text-red-700 font-medium"> *</span>
              </Label>
              <Select
                options={domainList}
                placeholder="Select Option"
                onChange={(value) => setProjectName(value)}
                className="dark:bg-dark-900"
              />
              {projectNameError && (
                <div className="text-red-600 text-sm font-medium">
                  {projectNameError}
                </div>
              )}
            </div>

            {/* Project Framework */}
            <div className="w-full">
              <Label>
                Framework <span className="text-red-700 font-medium"> *</span>
              </Label>
              <Select
                options={frameworks}
                placeholder="Select Option"
                onChange={(value) => setProjectFramework(value)}
                className="dark:bg-dark-900"
              />
              {projectFrameworkError && (
                <div className="text-red-600 text-sm font-medium">
                  {projectFrameworkError}
                </div>
              )}
            </div>

            {/* Database */}
            <div className="w-full">
              <Label>
                Database <span className="text-red-700 font-medium"> *</span>
              </Label>
              <Select
                options={databaseList}
                placeholder="Select database"
                onChange={(value) => setDatabaseName(value)}
                className="dark:bg-dark-900"
              />
              {databaseNameError && (
                <div className="text-red-600 text-sm font-medium">
                  {databaseNameError}
                </div>
              )}
            </div>
          </div>

          {/* Right Column: File Uploader */}
          <div className="w-full">
            <DropzoneComponent
              onFileSelect={(f) => {
                if (f instanceof File) {
                  setFile(f);
                  setFileError("");
                }

                if (f && f.name.endsWith(".zip") === false) {
                  setFile(null);
                  setFileError("Only .zip files are allowed.");
                }
              }}
            />
            {fileError && (
              <div className="text-red-600 text-sm font-medium mt-2">
                {fileError}
              </div>
            )}
          </div>

          {/* Full Width: Upload Button + Progress */}
          <div className="col-span-1 md:col-span-2 flex flex-col gap-4 mt-2">
            <Button
              disabled={
                !projectName ||
                !projectFramework ||
                !databaseName ||
                !file ||
                uploading
              }
              size="sm"
              variant="primary"
              onClick={() => handleProjectDeployment()}
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
