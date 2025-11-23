import { useEffect, useState } from "react";
import { UploadProgress } from "../../interfaces/common.interface";
import { projectService } from "../../services/PHP.service";
import { FileIcon } from "../../icons";
import ComponentCard from "../../components/common/ComponentCard";
import PageBreadcrumb from "../../components/common/PageBreadCrumb";
import PageMeta from "../../components/common/PageMeta";
import Label from "../../components/form/Label";
import Button from "../../components/ui/button/Button";
import { ArrowRight } from "lucide-react";
import DropzoneComponent from "../../components/form/form-elements/DropZone";
import Select from "../../components/form/Select";
import { Domain } from "../../interfaces/domain.interface";
import {
  domainManager,} from "../../services/domainManager.service";
import { databaseManager } from "../../services/databaseManager.service";
import ProjectProgress, {
  Step,
} from "../../components/ui/progress/ProjectProgress";
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

  //file
  const [file, setFile] = useState<File | null>(null);
  const [fileError, setFileError] = useState<string>("");

  //status
  const [uploading, setUploading] = useState(false);
  const [steps, setSteps] = useState<Step[]>([
    { title: "Initialize", description: "Setup project", hasError: false },
    { title: "Upload", description: "Upload files", hasError: false }, // error here
    { title: "Deploy", description: "Deploy to server", hasError: false },
    // {
    //   title: "SSL Setup",
    //   description: "Install SSL certificate and verify",
    //   hasError: false,
    // },
  ]);

  // Function to mark/unmark error for a specific step
  const setStepError = (index: number, error: boolean) => {
    setSteps((prev: Step[]) =>
      prev.map((step: Step, i: number) =>
        i === index ? { ...step, hasError: error } : step
      )
    );
  };

  const [currentStep, setCurrentStep] = useState(0);

  const [processing, setProcessing] = useState<boolean>(false);
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

  //upload handler
  const handleProjectDeployment = async () => {
    // Step 0: Validate form fields
    if (!projectName) {
      setProjectNameError("Please enter a valid domain name.");
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
    setProcessing(true);
    setProgress({
      chunkSizeMB: 0,
      uploadedChunks: 0,
      totalChunks: 0,
      percentage: 0,
    });
    setUploadError("");
    setUploadSuccess("");
    setCurrentStep(0);

    try {
      // ====== Step 0: Initiate Project ======
      const initResponse = await projectService.InitiateProject(
        projectName,
        databaseName
      );

      if (initResponse.success) {
        setUploadSuccess(initResponse.message);
        setStepError(0, false);
        setCurrentStep(1);
      } else {
        setUploadError(initResponse.message);
        setStepError(0, true);
        return;
      }

      // ====== Step 1: Upload Project Folder ======
      try {
        const uploadResult = await projectService.uploadProjectFolder(
          initResponse.project.id,
          projectName,
          file,
          (done, progress) => {
            setProgress(progress);
            if (done) {
              setCurrentStep(2);
              for (let step = 0; step <= 1; step++) setStepError(step, false);
            }
          }
        );

        if (!uploadResult.success) {
          setStepError(1, true);
          setUploadError(uploadResult.message);
          return;
        }
      } catch (err: any) {
        console.error(err);
        setStepError(1, true);
        setUploadError(
          "Unexpected error during upload: " + (err?.message || err)
        );
        return;
      }

      // ====== Step 2: Deploy Project ======
      try {
        const deployResult = await projectService.DeployProject(
          initResponse.project.id,
          projectName
        );

        if (deployResult.success) {
          setStepError(2, false);
          setCurrentStep(3);
          setUploadSuccess("Project deployed successfully!");
        } else {
          setStepError(2, true);
          setUploadError(deployResult.message);
          return;
        }
      } catch (err: any) {
        console.error(err);
        setStepError(2, true);
        setUploadError(
          "Unexpected error during deployment: " + (err?.message || err)
        );
        return;
      }

      // ====== Step 4: Issue SSL ======
      // try {
        
      //   const sslResult = await sslManager.issueSSL(projectName);

      //   if (sslResult.error) {
      //     setStepError(4, true);
          
      //     setUploadError("SSL setup failed: " + sslResult.message);
      //   } else {
      //     setStepError(4, false);
      //     setCurrentStep(4);
      //     setUploadSuccess(
      //       (prev) => prev + " SSL setup completed successfully!"
      //     );
      //   }
      // } catch (err: any) {
      //   console.error(err);
      //   setStepError(4, true);
      //   setUploadError(
      //     "Unexpected error during SSL setup: " + (err?.message || err)
      //   );
      // }
    } catch (err) {
      console.error(err);
      setUploadError("Project creation failed. Check console.");
    } finally {
      setUploading(false);
      setProgress(null);
    }
  };

  return (
    <>
      <PageMeta
        title="PHP Site Builder"
        description="Manage & Deploy PHP Website"
      />
      <PageBreadcrumb pageTitle="Build PHP Website" />

      {/* Main Layout Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* LEFT COLUMN: Form & Dropzone (Takes 2/3 width on large screens) */}
        <div className="lg:col-span-2">
          <ComponentCard
            title="Deploy Your WordPress Site"
            desc="Fill out the details below and upload your project zip file."
          >
            <div className="space-y-6">
              {/* Row 1: Inputs */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
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
                    <div className="text-red-600 text-sm font-medium mt-1">
                      {projectNameError}
                    </div>
                  )}
                </div>

                {/* Database */}
                <div className="w-full">
                  <Label>
                    Database{" "}
                    <span className="text-red-700 font-medium"> *</span>
                  </Label>
                  <Select
                    options={databaseList}
                    placeholder="Select database"
                    onChange={(value) => setDatabaseName(value)}
                    className="dark:bg-dark-900"
                  />
                  {databaseNameError && (
                    <div className="text-red-600 text-sm font-medium mt-1">
                      {databaseNameError}
                    </div>
                  )}
                </div>
              </div>

              <hr className="border-gray-200 dark:border-gray-700" />

              {/* Row 2: File Uploader */}
              <div className="w-full">
                <Label className="mb-2 block">Project Files (.zip)</Label>
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

                {/* File Errors */}
                {fileError && (
                  <div className="text-red-600 text-sm font-medium mt-2">
                    {fileError}
                  </div>
                )}

                {/* Upload Progress Bar */}
                {currentStep > 0 && progress && (
                  <div className="mt-4 p-4 bg-gray-50 dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700">
                    <div className="mb-2 flex justify-between items-center">
                      {/* File Icon + Text */}
                      <div className="flex items-center gap-x-3">
                        <span className="w-8 h-8 flex justify-center items-center bg-white dark:bg-gray-800 border border-blue-200 dark:border-gray-700 text-blue-500 rounded shadow-sm">
                          <FileIcon />
                        </span>

                        <div>
                          <p className="text-xs font-medium text-gray-800 dark:text-gray-200">
                            Uploading...
                          </p>
                          <p className="text-[10px] text-gray-500 dark:text-gray-400">
                            {progress.uploadedChunks * progress.chunkSizeMB} MB
                            of {progress.totalChunks * progress.chunkSizeMB} MB
                          </p>
                        </div>
                      </div>

                      {/* Percent */}
                      <div className="text-sm font-bold text-blue-600 dark:text-blue-400">
                        {progress.percentage.toFixed(1)}%
                      </div>
                    </div>

                    {/* TailAdmin Dark Compatible Progress Bar */}
                    <div className="w-full bg-gray-200 dark:bg-gray-700 h-2 rounded-full overflow-hidden">
                      <div
                        className="bg-blue-500 h-2 rounded-full transition-all duration-500 ease-out"
                        style={{ width: `${progress.percentage}%` }}
                      ></div>
                    </div>
                  </div>
                )}

                {processing && (
                  <ProjectProgress steps={steps} currentStep={currentStep} />
                )}

                {/* Error Message */}
                {uploadError && (
                  <div className="mt-3 p-3 bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400 text-sm font-medium rounded border border-red-200 dark:border-red-800">
                    Error: {uploadError}
                  </div>
                )}

                {/* Success Message */}
                {uploadSuccess && currentStep != 4 && (
                  <div className="mt-3 p-3 bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 text-sm font-medium rounded border border-green-200 dark:border-green-800">
                    {uploadSuccess}
                  </div>
                )}
                {/* Success Message */}
                {currentStep==4 && (
                  <div className="mt-3 p-3 bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 text-sm font-medium rounded border border-green-200 dark:border-green-800">
                    {uploadSuccess}
                    visit at <a
                    href={`https://${projectName}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-600 dark:text-blue-400 underline"
                  >
                    https://{projectName}
                  </a>
                  </div>
                )}
              </div>

              {/* Row 3: Action Button */}
              <div className="w-full pt-2">
                <Button
                  disabled={!projectName || !databaseName || !file || uploading}
                  isHidden={uploading}
                  size={"md"}
                  variant="primary"
                  className="w-full md:w-auto"
                  onClick={() => handleProjectDeployment()}
                  endIcon={!uploading && <ArrowRight />}
                >
                  Upload & Deploy
                </Button>
              </div>
            </div>
          </ComponentCard>
        </div>

        {/* RIGHT COLUMN: Instructions (Takes 1/3 width on large screens) */}
        <div className="lg:col-span-1">
          <ComponentCard
            title="Deployment Steps"
            desc="Follow this guide for success."
          >
            <div className="p-2">
              <ol className="relative border-l border-gray-200 dark:border-gray-700 ml-3 space-y-6">
                {[
                  "Create MySQL database & user.",
                  "Add domain in Domain Management.",
                  "Select domain & database.",
                  "Upload your .zip project file.",
                  'Click "Upload & Deploy".',
                  "Wait for completion & verify the site.",
                ].map((step, index) => (
                  <li key={index} className="ml-6">
                    <span className="absolute flex items-center justify-center w-6 h-6 bg-blue-100 rounded-full -left-3 ring-8 ring-white dark:ring-gray-900 dark:bg-blue-900">
                      <span className="text-xs font-bold text-blue-600 dark:text-blue-300">
                        {index + 1}
                      </span>
                    </span>
                    <p className="text-sm font-medium text-gray-900 dark:text-gray-300">
                      {step}
                    </p>
                  </li>
                ))}
              </ol>

              <div className="mt-6 bg-blue-50 dark:bg-blue-900/20 p-4 rounded-lg border border-blue-100 dark:border-blue-800">
                <h4 className="text-blue-800 dark:text-blue-300 text-lg font-bold uppercase mb-2">
                  Note
                </h4>
                <div className="space-y-1">
                  <p className="flex items-start text-sm text-blue-700 dark:text-blue-200">
                    <span className="flex-shrink-0 mr-2 font-bold text-blue-600 dark:text-blue-300">
                      1.
                    </span>
                    <span>
                      Choosing wrong domain or database can break other sites
                    </span>
                  </p>
                  <p className="flex items-start text-sm text-blue-700 dark:text-blue-200">
                    <span className="flex-shrink-0 mr-2 font-bold text-blue-600 dark:text-blue-300">
                      2.
                    </span>
                    <span>Don't close this page during project deployment</span>
                  </p>
                  <p className="flex items-start text-sm text-blue-700 dark:text-blue-200">
                    <span className="flex-shrink-0 mr-2 font-bold text-blue-600 dark:text-blue-300">
                      3.
                    </span>
                    <span>
                      Ensure your zip file contains the <code>index.php</code>{" "}
                      at the root level, not inside a subfolder.
                    </span>
                  </p>
                </div>
              </div>
            </div>
          </ComponentCard>
        </div>
      </div>
    </>
  );
}
