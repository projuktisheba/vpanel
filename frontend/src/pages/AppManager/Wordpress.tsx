import { useEffect, useState } from "react";
import ComponentCard from "../../components/common/ComponentCard";
import PageBreadcrumb from "../../components/common/PageBreadCrumb";
import PageMeta from "../../components/common/PageMeta";
import Label from "../../components/form/Label";
import Button from "../../components/ui/button/Button";
import { ArrowRight, Copy, Loader } from "lucide-react";
import Select from "../../components/form/Select";
import { Domain } from "../../interfaces/domain.interface";
import { domainManager } from "../../services/domainManager.service";
import { databaseManager } from "../../services/databaseManager.service";
import { wordpressService } from "../../services/wordpress.service";
import { Database } from "../../interfaces/database.interface";
import { Project } from "../../interfaces/project.interface";
import moment from "moment";
interface Option {
  value: string;
  label: string;
}
export default function WordpressSiteBuilder() {
  //domain
  const [domainName, setDomainName] = useState<string>("");
  const [domainNameError, setDomainNameError] = useState<string>("");
  const [domainList, setDomainList] = useState<Domain[]>([]);

  //databases
  const [databaseName, setDatabaseName] = useState<string>("");
  const [databaseNameError, setDatabaseNameError] = useState<string>("");
  const [databaseList, setDatabaseList] = useState<Database[]>([]);

  //status
  const [deploying, setDeploying] = useState(false);
  const [buildSuccess, setBuildSuccess] = useState<string>("");

  const [summary, setSummary] = useState<Project | null>();
  const [buildError, setBuildError] = useState<string>("");
  
  //fetch databases
  const fetchDatabaseList = async () => {
    try {
      const res = await databaseManager.listMySQLDB();
      // res here is the whole API response, so extract the databases array
      setDatabaseList(Array.isArray(res.databases) ? res.databases : []);
    } catch (err) {
      console.error("Failed to fetch databases:", err);
      setDatabaseList([]);
    } finally {
      // setLoading(false);
    }
  };
  //fetch domains

  const fetchDomains = async () => {
    try {
      const res = await domainManager.listDomains();
      // res here is the whole API response, so extract the domains array
      setDomainList(Array.isArray(res.domains) ? res.domains : []);
    } catch (err) {
      console.error("Failed to fetch domains:", err);
      setDomainList([]);
    }
  };
  useEffect(() => {
    fetchDatabaseList();
    fetchDomains();
  }, []);

  //selectable domains
  const databases: Option[] = databaseList.map((db) => ({
    value: db.dbName,
    label: db.dbName,
  }));
  const domains: Option[] = domainList.map((d) => ({
    value: d.domain,
    label: d.domain,
  }));
  //upload handler
  const handleProjectDeployment = async () => {
    // Step 1: Validate form fields
    if (domainName == "") {
      setDomainNameError("Please enter a valid domain name.");
      return;
    }

    if (databaseName == "") {
      setDatabaseNameError("Please select a valid database.");
      return;
    }

    setDeploying(true);
    setBuildError("");
    setBuildSuccess("");

    try {
      // Step 2: Call project creation service with form values (not the file)
      const createResp = await wordpressService.BuildProject(
        databaseName,
        domainName
      );

      // Step 4: Show API response
      if (createResp.error) {
        setBuildError(createResp.message);
      } else {
        setBuildSuccess(createResp.message);
        setSummary(createResp.summary ? createResp.summary : null);
      }
    } catch (err) {
      console.error(err);
      setBuildError("Upload or project creation failed. Check console.");
    } finally {
      setDeploying(false);
    }
  };

  return (
    <>
      <PageMeta
        title="PHP Site Builder"
        description="Manage & Build PHP Website"
      />
      <PageBreadcrumb pageTitle="Build PHP Website" />
      <ComponentCard>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Left Half: Form */}
          <div className="space-y-6">
            {/* Domain Name */}
            <div className="w-full">
              <Label>
                Domain <span className="text-red-700 font-medium"> *</span>
              </Label>
              <Select
                options={domains}
                placeholder="Select a domain"
                onChange={(value) => setDomainName(value)}
                className="dark:bg-dark-900"
              />
              {domainNameError && (
                <div className="text-red-600 text-sm font-medium">
                  {domainNameError}
                </div>
              )}
            </div>

            {/* Database */}
            <div className="w-full">
              <Label>
                Database <span className="text-red-700 font-medium"> *</span>
              </Label>
              <Select
                options={databases}
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

            {/* Deploy Button */}
            <div className="flex flex-col gap-4 mt-2">
              <Button
                disabled={domainName === "" || databaseName === "" || deploying}
                size="sm"
                variant="primary"
                onClick={() => handleProjectDeployment()}
                endIcon={<ArrowRight />}
              >
                {deploying ? (
                  <>
                    <Loader className="animate-spin w-4 h-4 mr-2" />
                    Deploying...
                  </>
                ) : (
                  "Deploy Website"
                )}
              </Button>
            </div>
          </div>

          {/* Right Half: Compact & Short Workflow */}
          <div className="space-y-4 p-4 bg-gray-50 dark:bg-dark-800 rounded-md">
            <h3 className="font-semibold text-lg">Deployment Steps</h3>
            <ol className="list-none space-y-2">
              <li className="flex items-start">
                <span className="flex-shrink-0 flex items-center justify-center w-6 h-6 rounded-full bg-blue-500 text-white text-xs font-bold">
                  1
                </span>
                <span className="ml-2 text-gray-700 dark:text-gray-300 text-sm">
                  Create MySQL database & user.
                </span>
              </li>
              <li className="flex items-start">
                <span className="flex-shrink-0 flex items-center justify-center w-6 h-6 rounded-full bg-blue-500 text-white text-xs font-bold">
                  2
                </span>
                <span className="ml-2 text-gray-700 dark:text-gray-300 text-sm">
                  Add domain in Domain Management.
                </span>
              </li>
              <li className="flex items-start">
                <span className="flex-shrink-0 flex items-center justify-center w-6 h-6 rounded-full bg-blue-500 text-white text-xs font-bold">
                  3
                </span>
                <span className="ml-2 text-gray-700 dark:text-gray-300 text-sm">
                  Select domain & database for deployment.
                </span>
              </li>
              <li className="flex items-start">
                <span className="flex-shrink-0 flex items-center justify-center w-6 h-6 rounded-full bg-blue-500 text-white text-xs font-bold">
                  4
                </span>
                <span className="ml-2 text-gray-700 dark:text-gray-300 text-sm">
                  Click <strong>"Deploy Website"</strong>.
                </span>
              </li>
              <li className="flex items-start">
                <span className="flex-shrink-0 flex items-center justify-center w-6 h-6 rounded-full bg-blue-500 text-white text-xs font-bold">
                  5
                </span>
                <span className="ml-2 text-gray-700 dark:text-gray-300 text-sm">
                  Wait for completion & check messages.
                </span>
              </li>
            </ol>
          </div>
        </div>
        <div>
          {/* Deployment Info Card */}
          {buildError && (
            <div className="mt-6 p-4 bg-red-50 dark:bg-red-900 border border-red-300 dark:border-red-700 rounded-md shadow-sm">
              <h4 className="font-semibold text-lg text-red-800 dark:text-red-200 mb-2">
                Build Failed !!
              </h4>

              <div className="grid grid-cols-1 sm:grid-cols-1 gap-2 text-sm text-gray-700 dark:text-gray-300">
                <div>{buildError}</div>
              </div>
            </div>
          )}

          {buildSuccess && summary && (
            <div className="mt-6 p-4 bg-green-50 dark:bg-green-900 border border-green-300 dark:border-green-700 rounded-md shadow-sm">
              <h4 className="font-semibold text-lg text-green-800 dark:text-green-200 mb-2">
                Congratulations! Site Deployed Successfully
              </h4>

              <div className="grid grid-cols-1 sm:grid-cols-1 gap-2 text-sm text-gray-700 dark:text-gray-300">
                <div>
                  <span className="font-medium">Domain:</span>{" "}
                  <a
                    href={`https://${summary.domainName}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-600 dark:text-blue-400 underline"
                  >
                    https://{summary.domainName}
                  </a>
                </div>
                <div>
                  <span className="font-medium">Project Directory:</span>{" "}
                  {summary.projectDirectory}
                </div>
                <div>
                  <span className="font-medium">Database Name:</span>{" "}
                  {summary.databaseInfo.dbName}
                </div>
                <div>
                  <span className="font-medium">Database User:</span>{" "}
                  {summary.databaseInfo.user?.username}
                </div>
                <div>
                  <span className="font-medium">Database Password:</span>{" "}
                  {summary.databaseInfo.user?.password}
                </div>
                <div>
                  <span className="font-medium">Framework:</span>{" "}
                  {summary.projectFramework}
                </div>
                <div>
                  <span className="font-medium">Status:</span> {summary.status}
                </div>
                <div>
                  <span className="font-medium">Created At:</span>{" "}
                  {moment(summary.createdAt).format("DD-MMM-YYYY")}
                </div>
              </div>

              {/* Copy Button */}
              <div className="mt-4 flex justify-end">
                <button
                  className="px-2 py-1 bg-green-600 text-white text-sm rounded hover:bg-green-700"
                  onClick={() => {
                    const textToCopy = `
                          Domain: https://${summary.domainName}
                          Database: ${summary.databaseInfo.dbName}
                          User: ${summary.databaseInfo.user?.username}
                          Password: ${summary.databaseInfo.user?.password}
                          Framework: ${summary.projectFramework}
                          Status: ${summary.status}
                          Created At: ${moment(summary.createdAt).format(
                            "DD-MMM-YYYY"
                          )}`;
                    navigator.clipboard.writeText(textToCopy);
                    alert("Project info copied to clipboard!");
                  }}
                >
                  <Copy />
                </button>
              </div>
            </div>
          )}
        </div>
      </ComponentCard>
    </>
  );
}
