import PageBreadcrumb from "../../components/common/PageBreadCrumb";
import ComponentCard from "../../components/common/ComponentCard";
import PageMeta from "../../components/common/PageMeta";
import Form from "../../components/form/Form";
import Button from "../../components/ui/button/Button";
import Input from "../../components/form/input/InputField";
import Label from "../../components/form/Label";
import { useState, FormEvent, useEffect } from "react";

import Select from "../../components/form/Select";
import { ArrowRightIcon, MoreDotIcon } from "../../icons";
import { databaseManager } from "../../services/databaseManager.service";
import DataTable from "../../components/tables/BasicTables/DataTable";
import { Database, DBUser } from "../../interfaces/databaseManager.interface";
import { Dropdown } from "../../components/ui/dropdown/Dropdown";
import { DropdownItem } from "../../components/ui/dropdown/DropdownItem";
import { Modal } from "../../components/ui/modal";
import AlertModal from "../../components/ui/modal/AlertModal";
interface Option {
  value: string;
  label: string;
}

export default function MySQL() {
  const [databaseName, setDatabaseName] = useState("");
  const [username, setUsername] = useState("");
  const [databaseNameError, setDatabaseNameError] = useState("");
  const [userNameError, setUsernameError] = useState("");
  const [databases, setDatabases] = useState<Database[]>([]);
  const [users, setUsers] = useState<DBUser[]>([]);
  const [loading, setLoading] = useState(false);
  const [openAlertModal, setOpenAlertModal] = useState(false);
  const [alertModalTitle, setAlertModalTitle] = useState("");
  const [alertModalMessage, setAlertModalMessage] =
    useState<React.ReactNode>("");
  const [alertModalType, setAlertModalType] = useState<
    "success" | "error" | "warning"
  >("success");

  // State to track which row's dropdown is open
  const [openDropdownId, setOpenDropdownId] = useState<string | "">("");
  const [deleteTargetDatabase, setDeleteTargetDatabase] =
    useState<Database | null>(null);

  // Toggle a specific row's dropdown
  function toggleDropdown(rowId: string) {
    setOpenDropdownId((prev) => (prev === rowId ? "" : rowId));
  }

  // Close a specific row's dropdown
  function closeDropdown() {
    setOpenDropdownId("");
  }

  //handleDelete shows a waring to the user and then delete the database
  function openDeleteDatabaseModal(db: Database) {
    setDeleteTargetDatabase(db);
  }
  function closeDeleteDatabaseModal() {
    setDeleteTargetDatabase(null);
  }

  const handleConfirmDeleteDatabase = async () => {
    if (!deleteTargetDatabase) return;

    setLoading(true);

    try {
      const resp = await databaseManager.deleteMySQLDB(
        deleteTargetDatabase.dbName
      );
      if (resp?.error == true) {
        <AlertModal
          isOpen={deleteTargetDatabase != null}
          onClose={closeDeleteDatabaseModal}
          title="Delete Database"
          type="error"
          message={
            <>
              You are about to delete the database{" "}
              <span className="font-semibold text-red-600">
                {deleteTargetDatabase?.dbName}
              </span>
              . This action is irreversible. All tables and data will be lost.
            </>
          }
          primaryAction={{
            label: "Delete Database",
            onClick: handleConfirmDeleteDatabase,
            className: "bg-red-600 hover:bg-red-700 text-white",
          }}
          secondaryAction={{
            label: "Cancel",
            onClick: closeDeleteDatabaseModal,
          }}
        />;
        return;
      }
      // Refresh list
      await fetchDatabases();
    } catch (err) {
      console.error("Failed to delete database:", err);

      // Optionally show a toast:
      // toast.error("Failed to delete database.");
    } finally {
      setLoading(false);
      // Close the modal after success
      closeDeleteDatabaseModal();
    }
  };

  // Fetch available users
  const fetchUsers = async () => {
    setLoading(true);
    try {
      const res = await databaseManager.listMySQLUsers();
      // res here is the whole API response, so extract the databases array
      setUsers(Array.isArray(res.dbUsers) ? res.dbUsers : []);
    } catch (err) {
      console.error("Failed to fetch database users:", err);
      setUsers([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchUsers();
  }, []);
  // Initialize options array
  const options: Option[] = users.map((user) => ({
    value: user.username,
    label: user.username,
  }));

  const createNewDatabase = (e: FormEvent) => {
    e.preventDefault();

    let hasError = false;

    // Validate database name
    if (!databaseName) {
      setDatabaseNameError("Please enter a unique database name");
      hasError = true;
    } else if (!databaseName.trim()) {
      setDatabaseNameError("Database name cannot be blank");
      hasError = true;
    } else {
      setDatabaseNameError("");
    }

    // Validate username
    if (!username) {
      setUsernameError("Please select a database user");
      hasError = true;
    } else {
      setUsernameError("");
    }

    if (hasError) return;

    (async () => {
      try {
        const data = await databaseManager.createMySQLDB(
          databaseName,
          username
        );
        console.log(data);

        // Set modal content
        setAlertModalTitle(data?.error ? "Error Creating Database" : "Success");
        setAlertModalMessage(data?.message || "Database created");
        setAlertModalType(data?.error ? "error" : "success");
        setOpenAlertModal(true);

        // If success, clear form
        if (!data?.error) {
          setDatabaseName("");
          setUsername("");
        }

        fetchDatabases(); // Refresh list
      } catch (err: any) {
        console.error(err);
        setAlertModalTitle("Error Creating Database");
        setAlertModalMessage(err.message || String(err));
        setAlertModalType("error");
        setOpenAlertModal(true);
      }
    })();
  };

  const fetchDatabases = async () => {
    setLoading(true);
    try {
      const res = await databaseManager.listMySQLDB();
      // res here is the whole API response, so extract the databases array
      setDatabases(Array.isArray(res.databases) ? res.databases : []);
    } catch (err) {
      console.error("Failed to fetch databases:", err);
      setDatabases([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDatabases();
  }, []);

  // Database table columns
  const columns = [
    {
      key: "dbName",
      label: "Database Name",
      className: "font-medium text-gray-800 text-theme-sm dark:text-white/90",
      render: (row: Database) => row.dbName,
    },
    {
      key: "users",
      label: "Users",
      className: "py-3 text-gray-500 text-theme-sm dark:text-gray-400",
      render: (row: Database) => row?.user?.username,
    },
    {
      key: "tableCount",
      label: "Tables",
      className: "py-3 text-gray-500 text-theme-sm dark:text-gray-400",
      render: (row: Database) => row.tableCount,
    },
    {
      key: "databaseSizeMB",
      label: "Size (MB)",
      className: "py-3 text-gray-500 text-theme-sm dark:text-gray-400",
      render: (row: Database) => row.databaseSizeMB?.toFixed(2),
    },
    {
      key: "createdAt",
      label: "Created At",
      className: "py-3 text-gray-500 text-theme-sm dark:text-gray-400",
      render: (row: Database) =>
        row.createdAt ? new Date(row.createdAt).toLocaleString() : "-",
    },
    {
      key: "updatedAt",
      label: "Last Updated",
      className: "py-3 text-gray-500 text-theme-sm dark:text-gray-400",
      render: (row: Database) =>
        row.updatedAt ? new Date(row.updatedAt).toLocaleString() : "-",
    },
    {
      key: "actions",
      label: "Action",
      className: "py-3 text-gray-500 text-theme-sm dark:text-gray-400",
      render: (row: Database) => (
        <div className="relative inline-block">
          <button
            className="dropdown-toggle"
            onClick={() => toggleDropdown(row.dbName)}
          >
            <MoreDotIcon className="text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 size-6" />
          </button>
          <Dropdown
            isOpen={openDropdownId === row.dbName}
            onClose={closeDropdown}
            className="w-40 p-2"
          >
            <DropdownItem
              onItemClick={closeDropdown}
              className="flex w-full font-normal text-left text-gray-500 rounded-lg hover:bg-gray-100 hover:text-gray-700 dark:text-gray-400 dark:hover:bg-white/5 dark:hover:text-gray-300"
            >
              Import
            </DropdownItem>
            <DropdownItem
              onItemClick={() => openDeleteDatabaseModal(row)}
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
  const searchOptions = [
    { value: "user.username", label: "User" },
    { value: "dbName", label: "Database" },
  ];
  return (
    <>
      <PageMeta title="MySQL" description="Manage MySQL Database" />
      <PageBreadcrumb pageTitle="Manage MySQL Database" />
      {loading && ""}

      <div className="space-y-6">
        <ComponentCard title="Create New Database">
          <Form onSubmit={createNewDatabase}>
            <div className="space-y-6">
              <div>
                <Label>Database Name</Label>
                <div className="relative">
                  <Input
                    placeholder="Enter database name"
                    type="text"
                    onChange={(e) => {
                      const value = e.target.value.trim();
                      setDatabaseName(value);

                      // Check uniqueness against databases list
                      if (databases.some((db) => db.dbName === value)) {
                        setDatabaseNameError("Database name already exists");
                      } else {
                        setDatabaseNameError(""); // no error
                      }
                    }}
                    error={databaseNameError !== ""}
                    hint={databaseNameError}
                  />
                </div>
              </div>
              <div>
                <Label>Database User</Label>
                <div className="relative">
                  <Select
                    options={options}
                    placeholder="Select User"
                    onChange={(value) => {
                      setUsername(value);
                      setUsernameError("");

                    }}
                    className="dark:bg-dark-900"
                    error={userNameError != ""}
                    hint={userNameError}
                  />
                </div>
              </div>
              <div className="flex justify-center">
                <Button
                  size="sm"
                  variant="primary"
                  endIcon={<ArrowRightIcon />}
                >
                  Create Table
                </Button>
              </div>
            </div>
          </Form>
        </ComponentCard>
        <ComponentCard title="List of Database">
          <DataTable
            data={databases}
            columns={columns}
            searchOptions={searchOptions}
          />
        </ComponentCard>
        <Modal
          isOpen={deleteTargetDatabase != null}
          onClose={closeDeleteDatabaseModal}
          className="max-w-[500px] m-4"
        >
          <div className="no-scrollbar relative w-full max-w-[500px] rounded-3xl bg-white p-6 dark:bg-gray-900 lg:p-8">
            {/* Header */}
            <div className="px-2">
              <h4 className="mb-3 text-2xl font-semibold text-gray-800 dark:text-white/90">
                Delete Database
              </h4>
              <p className="text-sm text-gray-500 dark:text-gray-400">
                You are about to delete the database{" "}
                <span className="font-semibold text-red-600">
                  {deleteTargetDatabase?.dbName}
                </span>
                . <br />
                This action is irreversible. All tables and data will be lost.
              </p>
            </div>

            {/* Warning Section */}
            <div className="mt-6 px-2">
              <div className="rounded-xl border border-red-300 bg-red-50 p-4 dark:border-red-700 dark:bg-red-900/20">
                <h5 className="mb-2 text-lg font-medium text-red-700 dark:text-red-400">
                  ⚠️ Warning
                </h5>
                <p className="text-sm text-red-600 dark:text-red-300">
                  Deleting this database cannot be undone. Make sure you have
                  backups before proceeding.
                </p>
              </div>
            </div>

            {/* Footer Buttons */}
            <div className="mt-8 flex items-center gap-3 px-2 lg:justify-end">
              <Button
                size="sm"
                variant="outline"
                onClick={closeDeleteDatabaseModal}
              >
                Cancel
              </Button>
              <Button
                size="sm"
                variant="primary"
                className="bg-red-600 hover:bg-red-700 text-white"
                onClick={handleConfirmDeleteDatabase}
              >
                Delete Database
              </Button>
            </div>
          </div>
        </Modal>
        <AlertModal
          isOpen={openAlertModal}
          onClose={() => setOpenAlertModal(false)}
          title={alertModalTitle}
          type={alertModalType}
          message={alertModalMessage}
          primaryAction={{
            label: "Close",
            onClick: () => setOpenAlertModal(false),
          }}
        />
      </div>
    </>
  );
}
