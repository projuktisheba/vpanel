import PageBreadcrumb from "../../../components/common/PageBreadCrumb";
import ComponentCard from "../../../components/common/ComponentCard";
import PageMeta from "../../../components/common/PageMeta";
import Button from "../../../components/ui/button/Button";
import Input from "../../../components/form/input/InputField";
import Label from "../../../components/form/Label";
import { useState, FormEvent, useEffect } from "react";

import Select from "../../../components/form/Select";
import { EyeCloseIcon, MoreDotIcon } from "../../../icons";
import { databaseManager } from "../../../services/databaseManager.service";
import DataTable from "../../../components/tables/BasicTables/DataTable";
import {
  Database,
  DBUser,
} from "../../../interfaces/databaseManager.interface";
import { Dropdown } from "../../../components/ui/dropdown/Dropdown";
import { DropdownItem } from "../../../components/ui/dropdown/DropdownItem";
import { Modal } from "../../../components/ui/modal";
import AlertModal from "../../../components/ui/modal/AlertModal";
import moment from "moment";
import FileInput from "../../../components/form/input/FileInput";

import { EyeIcon, Loader } from "lucide-react";
import Tabs from "../../UiElements/Tabs";
import Form from "../../../components/form/Form";
import { Preloader } from "../../../components/preloaders/Preloader";
interface Option {
  value: string;
  label: string;
}

export default function MySQL() {
  // Database
  const [databaseName, setDatabaseName] = useState("");
  const [username, setUsername] = useState("");
  const [databaseNameError, setDatabaseNameError] = useState("");
  const [userNameError, setUsernameError] = useState("");
  const [isDBCreating, setIsDBCreating] = useState(false);
  const [dbCreateError, setDbCreateError] = useState("");
  const [dbCreateSuccess, setDbCreateSuccess] = useState("");

  // List of Database
  const [databases, setDatabases] = useState<Database[]>([]);
  const [importDatabase, setImportDatabase] = useState<Database | null>(null);

  // Database Users
  const [users, setUsers] = useState<DBUser[]>([]);
  const [loading, setLoading] = useState(true);
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
  const [openCreateDBModal, setOpenCreateDBModal] = useState(false);
  const [openCreateUserModal, setOpenCreateUserModal] = useState(false);
  const [openImportDatabaseModal, setOpenImportDatabaseModal] = useState(false);
  const [newUsername, setNewUsername] = useState("");
  const [newUsernameError, setNewUsernameError] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [newPasswordError, setNewPasswordError] = useState("");
  const [isUserCreating, setIsUserCreating] = useState(false);
  const [newUserCreateSuccess, setNewUserCreateSuccess] = useState("");
  const [newUserCreateError, setNewUserCreateError] = useState("");
  const [importDatabaseSuccess, setImportDatabaseSuccess] = useState("");
  const [importDatabaseError, setImportDatabaseError] = useState("");
  const [sqlFile, setSqlFile] = useState<File | null>(null);

  // Fetch available users
  const fetchUsers = async () => {
    
    if (!openCreateDBModal && openCreateUserModal ) setLoading(true);
    try {
      const res = await databaseManager.listMySQLUsers();
      // res here is the whole API response, so extract the databases array
      setUsers(Array.isArray(res.users) ? res.users : []);
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

  // Database table columns
  const userListColumns = [
    {
      key: "username",
      label: "Username",
      className: "font-medium text-gray-800 text-theme-sm dark:text-white/90",
      render: (row: DBUser) => row?.username,
    },
    {
      key: "password",
      label: "Password",
      className: "py-3 text-gray-500 text-theme-sm dark:text-gray-400",
      render: (row: DBUser) => row?.password,
    },

    {
      key: "createdAt",
      label: "Created At",
      className: "py-3 text-gray-500 text-theme-sm dark:text-gray-400",
      render: (row: DBUser) =>
        row.createdAt
          ? moment.utc(row.createdAt).local().format("DD-MMM-YYYY")
          : "-",
    },
    {
      key: "actions",
      label: "Action",
      className: "py-3 text-gray-500 text-theme-sm dark:text-gray-400",
      noPrint: true,
      render: (row: DBUser) => (
        <div className="relative inline-block">
          <button
            className="dropdown-toggle"
            onClick={() => toggleDropdown(row.username)}
          >
            <MoreDotIcon className="text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 size-6" />
          </button>
          <Dropdown
            isOpen={openDropdownId === row.username}
            onClose={closeDropdown}
            className="w-40 p-2"
          >
            <DropdownItem
              onItemClick={() => setOpenCreateUserModal(true)}
              className="flex w-full font-normal text-left text-gray-500 rounded-lg hover:bg-gray-100 hover:text-gray-700 dark:text-gray-400 dark:hover:bg-white/5 dark:hover:text-gray-300"
            >
              Edit
            </DropdownItem>
            {/* <DropdownItem
              onItemClick={() => openDeleteDatabaseModal(row)}
              className="flex w-full font-normal text-left text-gray-500 rounded-lg hover:bg-gray-100 hover:text-gray-700 dark:text-gray-400 dark:hover:bg-white/5 dark:hover:text-gray-300"
            >
              Delete
            </DropdownItem> */}
          </Dropdown>
        </div>
      ),
    },
  ];

  // Table search options
  const userSearchOptions = [{ value: "username", label: "Username" }];

  // Initialize options array
  const options: Option[] = users.map((user) => ({
    value: user.username,
    label: user.username,
  }));

  const fetchDatabases = async () => {
    if (!openCreateDBModal && openCreateUserModal ) setLoading(true);
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
  const databaseListColumns = [
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
        row.createdAt
          ? moment.utc(row.createdAt).local().format("DD-MMM-YYYY")
          : "-",
    },
    {
      key: "actions",
      label: "Action",
      className: "py-3 text-gray-500 text-theme-sm dark:text-gray-400",
      noPrint: true,
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
              onItemClick={() => {
                setImportDatabase(row);
                setOpenImportDatabaseModal(true);
              }}
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
  const dbSearchOptions = [
    { value: "user.username", label: "User" },
    { value: "dbName", label: "Database" },
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

      if (resp?.error) {
        
        setAlertModalTitle("Error Deleting Database");
        setAlertModalMessage(resp.message || "Failed to delete the database.");
        setAlertModalType("error");
        setOpenAlertModal(true);
        return; // stop here
      }

      // Success
      setAlertModalTitle("Database Deleted");
      setAlertModalMessage(
        resp.message ||
          `Database "${deleteTargetDatabase.dbName}" deleted successfully.`
      );
      setAlertModalType("success");
      setOpenAlertModal(true);

      await fetchDatabases(); // refresh list
    } catch (err: any) {
      console.error("Failed to delete database:", err);
      setAlertModalTitle("Error Deleting Database");
      setAlertModalMessage(err.message || String(err));
      setAlertModalType("error");
      setOpenAlertModal(true);
    } finally {
      setLoading(false);
      closeDeleteDatabaseModal();
    }
  };

  //Create Database
  // handlers for creating new database
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
      setIsDBCreating(true);

      const data = await databaseManager.createMySQLDB(databaseName, username);
      console.log(data);

      // If error, show error message on the form
      if (data?.error == true) {
        setDbCreateError(data?.message);
        setDbCreateSuccess("");
      } else {
        setDbCreateError("");
        setDbCreateSuccess(data?.message);
        fetchDatabases(); // Refresh list
      }
      setIsDBCreating(false);
    })();
  };

  const closeCreateDatabaseModal = () => {
    setDatabaseName("");
    setDatabaseNameError("");
    setUsername("");
    setUsernameError("");
    setDbCreateError("");
    setDbCreateSuccess("");
    setIsDBCreating(false);
    setOpenCreateDBModal(false);
  };

  // Create User
  // handle onclose modal
  const closeCreateUserModal = () => {
    setOpenCreateUserModal(false);
    setNewUsername("");
    setNewUsernameError("");
    setNewPassword("");
    setNewPasswordError("");
    setShowPassword(false);
    setNewUserCreateSuccess("");
    setNewUserCreateError("");
  };

  const createNewUser = (e: FormEvent) => {
    e.preventDefault();

    let hasError = false;

    // Validate Username
    if (!newUsername) {
      setNewUsernameError("Please enter a unique username");
      hasError = true;
    } else if (!newUsername.trim()) {
      setNewUsernameError("Username cannot be blank");
      hasError = true;
    } else {
      setNewUsernameError("");
    }

    // Validate password
    if (!newPassword.trim()) {
      setNewPasswordError("Username cannot be blank");
      hasError = true;
    } else {
      setNewPasswordError("");
    }

    if (hasError) return;

    (async () => {
      setIsUserCreating(true);
      const data = await databaseManager.createMySQLUser(
        newUsername,
        newPassword
      );
      console.log(data);

      // If error, show error message on the form
      if (data?.error == true) {
        setNewUserCreateError(data?.message);
        setNewUserCreateSuccess("");
      } else {
        setNewUserCreateError("");
        setNewUserCreateSuccess(data?.message);
        fetchUsers(); // Refresh list
      }
      setIsUserCreating(false);
    })();
  };

  // Import Database
  // Close modal and reset state
  const closeImportDatabaseModal = () => {
    setOpenImportDatabaseModal(false);
    setImportDatabaseError("");
    setImportDatabase(null);
    setSqlFile(null);
  };

  // Handle file selection
  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      const file = e.target.files[0];

      // Validate .sql extension
      if (!file.name.toLowerCase().endsWith(".sql")) {
        setImportDatabaseError("Only .sql files are allowed.");
        setSqlFile(null);
        return;
      }

      setSqlFile(file);
      setImportDatabaseError(""); // clear previous errors
    }
  };

  // Import SQL file
  const handleImportDatabase = async () => {
    if (!sqlFile) {
      setImportDatabaseError("Please select a SQL file to import.");
      return;
    }

    setImportDatabaseError(""); // clear previous errors

    if (!importDatabase) {
      setImportDatabaseError("Please select a database to import into.");
      return;
    }

    try {
      const formData = new FormData();
      formData.append("dbName", importDatabase.dbName); // database name
      formData.append("sqlFile", sqlFile); // upload file

      // Call backend
      const response = await databaseManager.importMySQLDB(formData);

      if (!response.error) {
        setImportDatabaseSuccess(
          response.message || "Database imported successfully!"
        );
      } else {
        setImportDatabaseError(response.message || "Import failed.");
      }
    } catch (err: any) {
      console.error("Import failed:", err);
      setImportDatabaseError(err?.message || "An unexpected error occurred.");
    }
  };

  if (loading) {
    return <Preloader/>
  }

  return (
    <>
      <PageMeta title="MySQL" description="Manage MySQL Database" />
      <PageBreadcrumb pageTitle="Manage MySQL Database" />
      <ComponentCard>
        <div className="space-y-6">
          <Tabs
            tabs={[
              {
                label: "Database",
                content: (
                  <DataTable
                    data={databases}
                    columns={databaseListColumns}
                    searchOptions={dbSearchOptions}
                    title={"List of MySQL Database"}
                    extraActions={
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => setOpenCreateDBModal(true)}
                      >
                        Add Database
                      </Button>
                    }
                  />
                ),
              },
              {
                label: "Users",
                content: (
                  <DataTable
                    data={users}
                    columns={userListColumns}
                    searchOptions={userSearchOptions}
                    title={"List of MySQL User"}
                    extraActions={
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => setOpenCreateUserModal(true)}
                      >
                        Add User
                      </Button>
                    }
                  />
                ),
              },
            ]}
          />

          {/* Create MySQL Database Modal */}
          <Modal
            isOpen={openCreateDBModal}
            onClose={closeCreateDatabaseModal}
            className="max-w-[500px] m-4"
          >
            <div className="no-scrollbar relative w-full max-w-[500px] rounded-3xl bg-white p-6 dark:bg-gray-900 lg:p-8">
              {/* Header */}
              <div className="px-2">
                <h4 className="mb-3 text-2xl font-semibold text-gray-800 dark:text-white/90">
                  Create New Database
                </h4>
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  Enter a database name and assign a user to create a new MySQL
                  database.
                </p>
              </div>

              {/* Form */}
              <form
                onSubmit={createNewDatabase}
                className="mt-6 flex flex-col gap-4"
              >
                {/* Database Name */}
                <div>
                  <Label>
                    Database Name{" "}
                    <span className="text-red-700 font-medium"> *</span>
                  </Label>
                  <Input
                    placeholder="Enter database name"
                    type="text"
                    value={databaseName}
                    onChange={(e) => {
                      const value = e.target.value.trim();
                      setDatabaseName(value);

                      if (databases.some((db) => db.dbName === value)) {
                        setDatabaseNameError("Database name already exists");
                      } else {
                        setDatabaseNameError("");
                      }
                    }}
                    error={databaseNameError !== ""}
                    hint={databaseNameError}
                  />
                </div>

                {/* Database User */}
                <div>
                  <Label>
                    Database User{" "}
                    <span className="text-red-700 font-medium"> *</span>
                  </Label>
                  <Select
                    options={options}
                    placeholder="Select User"
                    onChange={(value) => {
                      setUsername(value);
                      setUsernameError("");
                    }}
                    className="dark:bg-dark-900"
                    error={userNameError !== ""}
                    hint={userNameError}
                  />
                </div>

                {/* Success message for DB creation Display (Optional) */}
                {dbCreateSuccess && (
                  <div className="text-green-600 text-sm font-medium">
                    {dbCreateSuccess}
                  </div>
                )}

                {/* Error message for DB creation Display (Optional)*/}
                {dbCreateError && (
                  <div className="text-red-600 text-sm font-medium">
                    {dbCreateError}
                  </div>
                )}

                {/* Footer Buttons */}
                <div className="mt-6 flex items-center gap-3 lg:justify-end">
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={closeCreateDatabaseModal}
                  >
                    Cancel
                  </Button>

                  <Button
                    size="sm"
                    variant="primary"
                    disabled={
                      !databaseName ||
                      !username ||
                      databaseNameError !== "" ||
                      userNameError !== ""
                    }
                  >
                    {isDBCreating ? (
                      <>
                        <Loader className="animate-spin w-4 h-4 mr-2" />
                        Processing...
                      </>
                    ) : (
                      "Create Database"
                    )}
                  </Button>
                </div>
              </form>
            </div>
          </Modal>

          {/* Modal for import database */}
          <Modal
            isOpen={openImportDatabaseModal}
            onClose={closeImportDatabaseModal}
            className="max-w-[500px] m-4"
          >
            <div className="no-scrollbar relative w-full max-w-[500px] rounded-3xl bg-white p-6 dark:bg-gray-900 lg:p-8">
              {/* Header */}
              <div className="px-2">
                <h4 className="mb-3 text-2xl font-semibold text-gray-800 dark:text-white/90">
                  Import Database
                </h4>
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  Upload a .SQL file to import data into the database.
                </p>
              </div>

              {/* Form */}
              <div className="mt-6 flex flex-col gap-4">
                <div>
                  <Label>
                    Upload SQL File{" "}
                    <span className="text-red-700 font-medium"> *</span>
                  </Label>
                  <FileInput accept=".sql" onChange={handleFileChange} />
                </div>

                {importDatabaseError && (
                  <div className="text-red-600 text-sm font-medium">
                    {importDatabaseError}
                  </div>
                )}
                {importDatabaseSuccess && (
                  <div className="text-green-600 text-sm font-medium">
                    {importDatabaseSuccess}
                  </div>
                )}

                {/* Footer Buttons */}
                <div className="mt-6 flex items-center gap-3 lg:justify-end">
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={closeImportDatabaseModal}
                  >
                    Cancel
                  </Button>

                  <Button
                    disabled={!sqlFile} // disable if no file selected
                    size="sm"
                    variant="primary"
                    onClick={() => handleImportDatabase()} // wrap in arrow to fix TS type
                  >
                    Import
                  </Button>
                </div>
              </div>
            </div>
          </Modal>

          {/* Confirmation dialog for deleting database */}
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

          {/* Create MySQL User Modal */}
          <Modal
            isOpen={openCreateUserModal}
            onClose={closeCreateUserModal}
            className="max-w-[500px] m-4"
          >
            <div className="no-scrollbar relative w-full max-w-[500px] rounded-3xl bg-white p-6 dark:bg-gray-900 lg:p-8">
              {/* Header */}
              <div className="px-2">
                <h4 className="mb-3 text-2xl font-semibold text-gray-800 dark:text-white/90">
                  Create Database User
                </h4>
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  Fill in the details to create a new MySQL database user. Make
                  sure the username is unique.
                </p>
              </div>

              {/* Form */}
              <Form
                onSubmit={createNewUser}
                className="mt-6 flex flex-col gap-4"
              >
                {/* Username */}
                <div>
                  <Label>
                    Username{" "}
                    <span className="text-red-700 font-medium"> *</span>
                  </Label>
                  <Input
                    placeholder="Enter username"
                    type="text"
                    value={newUsername}
                    onChange={(e) => {
                      const value = e.target.value.trim();
                      setNewUsername(value);

                      if (users.some((u) => u.username === value)) {
                        setNewUsernameError("Username already exists");
                      } else {
                        setNewUsernameError("");
                      }
                    }}
                    error={newUsernameError !== ""}
                    hint={newUsernameError}
                  />
                </div>

                {/* Password */}
                <div>
                  <Label>
                    Password{" "}
                    <span className="text-red-700 font-medium"> *</span>
                  </Label>
                  <div className="relative">
                    <Input
                      type={showPassword ? "text" : "password"}
                      placeholder="Enter user password"
                      value={newPassword}
                      onChange={(e) => {
                        const value = e.target.value.trim();
                        setNewPassword(value);
                      }}
                      error={newPasswordError !== ""}
                      hint={newPasswordError}
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      className="absolute z-30 -translate-y-1/2 cursor-pointer right-4 top-1/2"
                    >
                      {showPassword ? (
                        <EyeIcon className="fill-white-200 dark:fill-gray-400 size-5" />
                      ) : (
                        <EyeCloseIcon className="fill-gray-400 dark:fill-gray-400 size-5" />
                      )}
                    </button>
                  </div>
                </div>

                {/* Success message for user creation Display (Optional) */}
                {newUserCreateSuccess && (
                  <div className="text-green-600 text-sm font-medium">
                    {newUserCreateSuccess}
                  </div>
                )}

                {/* Error message for user creation Display (Optional)*/}
                {newUserCreateError && (
                  <div className="text-red-600 text-sm font-medium">
                    {newUserCreateError}
                  </div>
                )}

                {/* Footer Buttons */}
                <div className="mt-6 flex items-center gap-3 lg:justify-end">
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={closeCreateUserModal}
                  >
                    Cancel
                  </Button>

                  <Button
                    size="sm"
                    variant="primary"
                    disabled={
                      newUsername == "" ||
                      newPassword == "" ||
                      newUsernameError != "" ||
                      newPasswordError != ""
                    }
                  >
                    {isUserCreating ? (
                      <>
                        <Loader className="animate-spin w-4 h-4 mr-2" />
                        Processing...
                      </>
                    ) : (
                      "Create User"
                    )}
                  </Button>
                </div>
              </Form>
            </div>
          </Modal>
          {/* Alerts */}
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
      </ComponentCard>
    </>
  );
}
