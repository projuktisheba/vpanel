
import { Loader } from "lucide-react";
import moment from "moment";
import { useState, useEffect, FormEvent } from "react";
import ComponentCard from "../../components/common/ComponentCard";
import PageBreadcrumb from "../../components/common/PageBreadCrumb";
import PageMeta from "../../components/common/PageMeta";
import Input from "../../components/form/input/InputField";
import Label from "../../components/form/Label";
import { Preloader } from "../../components/preloader/Preloader";
import DataTable from "../../components/tables/BasicTables/DataTable";
import Button from "../../components/ui/button/Button";
import { Dropdown } from "../../components/ui/dropdown/Dropdown";
import { DropdownItem } from "../../components/ui/dropdown/DropdownItem";
import { Modal } from "../../components/ui/modal";
import { MoreDotIcon } from "../../icons";
import { domainManager } from "../../services/domainManager.service";
import { Domain } from "../../interfaces/domain.interface";

export default function DomainManager() {
  // Domain
  const [domainName, setDomainName] = useState("");
  const [domainNameError, setDomainNameError] = useState("");
  const [domainNameProvider, setDomainNameProvider] = useState("");
  const [domainNameProviderError, setDomainNameProviderError] = useState("");
  const [isDomainCreating, setIsDomainCreating] = useState(false);
  const [domainCreateError, setDomainCreateError] = useState("");
  const [domainCreateSuccess, setDomainCreateSuccess] = useState("");

  // List of Domain
  const [domains, setDomains] = useState<Domain[]>([]);

  // Domain Users
  const [loading, setLoading] = useState(true);

  // State to track which row's dropdown is open
  const [openDropdownId, setOpenDropdownId] = useState<string | "">("");
  const [openAddDomainModal, setOpenAddDomainModal] = useState(false);

  const fetchDomains = async () => {
    if (!openAddDomainModal) setLoading(true);
    try {
      const res = await domainManager.listDomains();
      // res here is the whole API response, so extract the domains array
      setDomains(Array.isArray(res.domains) ? res.domains : []);
    } catch (err) {
      console.error("Failed to fetch domains:", err);
      setDomains([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDomains();
  }, []);

  // Domain table columns
  const domainListColumns = [
    {
      key: "domain",
      label: "Domain Name",
      className: "font-medium text-gray-800 text-theme-sm dark:text-white/90",
      render: (row: Domain) => row.domain,
    },
    {
      key: "domainProvider",
      label: "Provider",
      className: "py-3 text-gray-500 text-theme-sm dark:text-gray-400",
      render: (row: Domain) => row.domainProvider,
    },
    {
      key: "updatedAt",
      label: "Last Update",
      className: "py-3 text-gray-500 text-theme-sm dark:text-gray-400",
      render: (row: Domain) =>
        row.updatedAt
          ? moment.utc(row.updatedAt).local().format("DD-MMM-YYYY")
          : "-",
    },
    {
      key: "createdAt",
      label: "Created At",
      className: "py-3 text-gray-500 text-theme-sm dark:text-gray-400",

      render: (row: Domain) =>
        row.createdAt
          ? moment.utc(row.createdAt).local().format("DD-MMM-YYYY")
          : "-",
    },
    {
      key: "actions",
      label: "Action",
      className: "py-3 text-gray-500 text-theme-sm dark:text-gray-400",
      noPrint: true,
      render: (row: Domain) => (
        <div className="relative inline-block">
          <button
            className="dropdown-toggle"
            onClick={() => toggleDropdown(row.domain)}
          >
            <MoreDotIcon className="text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 size-6" />
          </button>
          <Dropdown
            isOpen={openDropdownId === row.domain}
            onClose={closeDropdown}
            className="w-40 p-2"
          >
            <DropdownItem
              onItemClick={() => {
                alert("Incomplete Editing section");
              }}
              className="flex w-full font-normal text-left text-gray-500 rounded-lg hover:bg-gray-100 hover:text-gray-700 dark:text-gray-400 dark:hover:bg-white/5 dark:hover:text-gray-300"
            >
              Edit
            </DropdownItem>
            <DropdownItem
              onItemClick={() => alert("Incomplete Deleting section")}
              className="flex w-full font-normal text-left text-gray-500 rounded-lg hover:bg-gray-100 hover:text-gray-700 dark:text-gray-400 dark:hover:bg-white/5 dark:hover:text-gray-300"
            >
              Delete
            </DropdownItem>
          </Dropdown>
        </div>
      ),
    },
  ];

  // Table search options
  const domainSearchOptions = [
    { value: "domain", label: "Domain" },
    { value: "domainProvider", label: "Provider" },
  ];

  // ===================== Handlers =====================

  // Toggle a specific row's dropdown
  function toggleDropdown(rowId: string) {
    setOpenDropdownId((prev) => (prev === rowId ? "" : rowId));
  }

  // Close a specific row's dropdown
  function closeDropdown() {
    setOpenDropdownId("");
  }

  const closeAddDomainModal = () => {
    setDomainName("");
    setDomainNameError("");
    setDomainCreateError("");
    setDomainCreateSuccess("");
    setIsDomainCreating(false);
    setOpenAddDomainModal(false);
  };

  const addNewDomain = (e: FormEvent) => {
    e.preventDefault();

    let hasError = false;

    // Trim names
    const trimmedDomain = domainName.trim();
    const trimmedDomainNameProvider = domainNameProvider.trim();

    // Validate Domain Name
    if (!trimmedDomain) {
      setDomainNameError("Please enter a domain name");
      hasError = true;
    } else if (domains.some((d) => d.domain === trimmedDomain)) {
      setDomainNameError("Domain name already exists");
      hasError = true;
    } else {
      setDomainNameError("");
    }
    // Validate Domain Provider Name
    if (!trimmedDomainNameProvider) {
      setDomainNameProviderError("Please enter the domain provider name");
      hasError = true;
    } else {
      setDomainNameProviderError("");
    }

    if (hasError) return;

    (async () => {
      setIsDomainCreating(true);

      try {
        const data = await domainManager.addNewDomain(trimmedDomain, trimmedDomainNameProvider);
        console.log(data);

        if (data?.error) {
          // Show error from backend
          setDomainCreateError(data.message || "Failed to create domain");
          setDomainCreateSuccess("");
        } else {
          // Success
          setDomainCreateError("");
          setDomainCreateSuccess(data.message || "Domain created successfully");

          // Refresh domain list
          fetchDomains();
        }
      } catch (err: any) {
        setDomainCreateError(err.message || "Something went wrong");
        setDomainCreateSuccess("");
      } finally {
        setIsDomainCreating(false);
      }
    })();
  };

  if (loading) {
    return <Preloader />;
  }

  return (
    <>
      <PageMeta title="Domains" description="Manage Domains" />
      <PageBreadcrumb pageTitle="Domain Manager" />
      <ComponentCard>
        <div className="space-y-6">
          <DataTable
            data={domains}
            columns={domainListColumns}
            searchOptions={domainSearchOptions}
            title={"List of Connected Domain"}
            extraActions={
              <Button
                size="sm"
                variant="outline"
                onClick={() => setOpenAddDomainModal(true)}
              >
                Add Domain
              </Button>
            }
          />

          {/* ADD Domain Modal */}
          <Modal
            isOpen={openAddDomainModal}
            onClose={closeAddDomainModal}
            className="max-w-[500px] m-4"
          >
            <div className="no-scrollbar relative w-full max-w-[500px] rounded-3xl bg-white p-6 dark:bg-gray-900 lg:p-8">
              {/* Header */}
              <div className="px-2">
                <h4 className="mb-3 text-2xl font-semibold text-gray-800 dark:text-white/90">
                  Create New Domain
                </h4>
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  Enter a domain name and its provider name
                </p>
              </div>

              {/* Form */}
              <form
                onSubmit={addNewDomain}
                className="mt-6 flex flex-col gap-4"
              >
                {/* Domain Name */}
                <div>
                  <Label>
                    Domain Name {"  "}
                    <span className="text-red-700 font-medium"> *</span>
                  </Label>
                  <Input
                    placeholder="Enter domain name"
                    type="text"
                    value={domainName}
                    onChange={(e) => {
                      const value = e.target.value.trim();
                      setDomainName(value);

                      if (domains.some((d) => d.domain === value)) {
                        setDomainNameError("Domain name already exists");
                      } else {
                        setDomainNameError("");
                      }
                    }}
                    error={domainNameError !== ""}
                    hint={domainNameError}
                  />
                </div>

                {/* Domain Provider Name */}
                <div>
                  <Label>
                    Provider Name {"  "}
                    <span className="text-red-700 font-medium"> *</span>
                  </Label>
                  <Input
                    placeholder="Enter Domain Provider name"
                    type="text"
                    value={domainNameProvider}
                    onChange={(e) => {
                      const value = e.target.value.trim();
                      setDomainNameProvider(value);

                      if (domains.some((d) => d.domain === value)) {
                        setDomainNameProviderError("Domain name already exists");
                      } else {
                        setDomainNameProviderError("");
                      }
                    }}
                    error={domainNameProviderError !== ""}
                    hint={domainNameProviderError}
                  />
                </div>

                {/* Success message for domain creation Display (Optional) */}
                {domainCreateSuccess && (
                  <div className="text-green-600 text-sm font-medium">
                    {domainCreateSuccess}
                  </div>
                )}

                {/* Error message for DOMAIN creation Display (Optional)*/}
                {domainCreateError && (
                  <div className="text-red-600 text-sm font-medium">
                    {domainCreateError}
                  </div>
                )}

                {/* Footer Buttons */}
                <div className="mt-6 flex items-center gap-3 lg:justify-end">
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={closeAddDomainModal}
                  >
                    Cancel
                  </Button>

                  <Button
                    size="sm"
                    variant="primary"
                    disabled={!domainName || domainNameError !== "" || !domainNameProvider || domainNameProviderError !== ""}
                  >
                    {isDomainCreating ? (
                      <>
                        <Loader className="animate-spin w-4 h-4 mr-2" />
                        Adding...
                      </>
                    ) : (
                      "Add Domain"
                    )}
                  </Button>
                </div>
              </form>
            </div>
          </Modal>
        </div>
      </ComponentCard>
    </>
  );
}
