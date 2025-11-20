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
        title="Wordpress Site Builder"
        description="Manage & Build Wordpress Website"
      />
      <PageBreadcrumb pageTitle="Build Wordpress Website" />
      <ComponentCard>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Left Half: Form */}
          <ComponentCard
            title="Deploy Your WordPress Site"
            desc="Fill out the domain and database information below, then click 'Deploy Website' to start."
          >
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
                  disabled={
                    domainName === "" || databaseName === "" || deploying
                  }
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
          </ComponentCard>
          {/* Right Half: Compact & Short Workflow */}
          <ComponentCard
            title="Deployment Steps"
            desc="Follow this guide for success."
          >
            <div className="p-2">
              <ol className="relative border-l border-gray-200 dark:border-gray-700 ml-3 space-y-6">
                {[
                  "Create MySQL database & user.",
                  "Add domain in Domain Management.",
                  "Select domain & database for deployment.",
                  'Click "Deploy Website".',
                  "Wait for completion & check messages.",
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
                <h4 className="text-blue-800 dark:text-blue-300 text-xs font-bold uppercase mb-1">
                  Note
                </h4>
                <p className="text-sm text-blue-600 dark:text-blue-300">
                  Make sure your domain
                  {domainName && (<span className="mx-1 px-1 py-0.25 bg-gray-200 text-red-600 dark:bg-gray-700 rounded">
                    {domainName}
                  </span>)}
                  has a DNS record pointing to this server:
                  <code className="px-1 py-0.5 bg-gray-100 dark:bg-gray-800 rounded">
                    203.161.48.179
                  </code>
                </p>
              </div>
            </div>
          </ComponentCard>
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
