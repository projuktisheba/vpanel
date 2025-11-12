import PageBreadcrumb from "../../components/common/PageBreadCrumb";
import ComponentCard from "../../components/common/ComponentCard";
import PageMeta from "../../components/common/PageMeta";
import Form from "../../components/form/Form";
import Button from "../../components/ui/button/Button";
import Input from "../../components/form/input/InputField";
import Label from "../../components/form/Label";
import { useState, FormEvent, useEffect } from "react";

import Select from "../../components/form/Select";
import { ArrowRightIcon } from "../../icons";
import { databaseManager } from "../../services/databaseManager.service";
import DataTable from "../../components/tables/BasicTables/DataTable";
import { DatabaseMeta } from "../../interfaces/databaseManager.interface";
interface Option {
  value: string;
  label: string;
}

export default function MySQL() {
  const [databaseName, setDatabaseName] = useState("");
  const [username, setUsername] = useState("");
  const [databaseNameError, setDatabaseNameError] = useState("");
  const [userNameError, setUsernameError] = useState("");
  const [databases, setDatabases] = useState<DatabaseMeta[]>([]);
  const [loading, setLoading] = useState(false);

  // Fetch available users
  const users = [
    { id: 1, name: "psi_api_user" },
    { id: 2, name: "vpanel_user" },
  ];

  // Initialize options array
  const options: Option[] = users.map((user) => ({
    value: user.name,
    label: user.name,
  }));

  const createNewDatabase = (e: FormEvent) => {
    e.preventDefault();

    let hasError = false;

    if (!databaseName) {
      setDatabaseNameError("Please enter a unique database name");
      hasError = true;
    } else if (!databaseName.trim()) {
      setDatabaseNameError("Database name cannot be blank");
      hasError = true;
    } else {
      setDatabaseNameError("");
    }

    if (!username) {
      setUsernameError("Please select a database user");
      hasError = true;
    } else {
      setUsernameError("");
    }

    if (hasError) {
      return;
    }
    (async () => {
      try {
        const data = await databaseManager.createMySQLDB(
          databaseName,
          username
        );
        console.log(data);
        // success - clear form and errors
        setDatabaseName("");
        setUsername("");
        setDatabaseNameError("");
        setUsernameError("");
        alert(data.message);
        fetchDatabases()
      } catch (err) {
        console.error(err);
        setDatabaseNameError(
          "An unexpected error occurred while creating the database"
        );
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
    { key: "dbName", label: "Database Name", className:"font-medium text-gray-800 text-theme-sm dark:text-white/90", render: (row: DatabaseMeta) => row.dbName,},
    
    {
      key: "users",
      label: "Users",
       className:"py-3 text-gray-500 text-theme-sm dark:text-gray-400",
      render: (row: DatabaseMeta) => row?.users?.join(", "),
    },
    { key: "tableCount", label: "Tables",  className:"py-3 text-gray-500 text-theme-sm dark:text-gray-400" , render: (row: DatabaseMeta) => row.tableCount,},
    {
      key: "databaseSizeMB",
      label: "Size (MB)",
       className:"py-3 text-gray-500 text-theme-sm dark:text-gray-400",
      render: (row: DatabaseMeta) => row.databaseSizeMB?.toFixed(2),
    },
    {
      key: "createdAt",
      label: "Created At",
       className:"py-3 text-gray-500 text-theme-sm dark:text-gray-400",
      render: (row: DatabaseMeta) =>
        row.createdAt ? new Date(row.createdAt).toLocaleString() : "-",
    },
    {
      key: "updatedAt",
      label: "Last Updated",
       className:"py-3 text-gray-500 text-theme-sm dark:text-gray-400",
      render: (row: DatabaseMeta) =>
        row.updatedAt ? new Date(row.updatedAt).toLocaleString() : "-",
    },
  ];

  // Table search options
  const searchOptions = [
    { value: "users", label: "User" },
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
                      setDatabaseName(e.target.value);
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
                    onChange={setUsername}
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
          <DataTable data={databases} columns={columns} searchOptions={searchOptions} />
          {/* <DatabaseTable/> */}
        </ComponentCard>
      </div>
    </>
  );
}
