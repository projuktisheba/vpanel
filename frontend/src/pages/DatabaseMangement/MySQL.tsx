import PageBreadcrumb from "../../components/common/PageBreadCrumb";
import ComponentCard from "../../components/common/ComponentCard";
import PageMeta from "../../components/common/PageMeta";
import DatabaseTable from "../../components/tables/BasicTables/DatabaseTable";
import Form from "../../components/form/Form";
import Button from "../../components/ui/button/Button";
import Input from "../../components/form/input/InputField";
import Label from "../../components/form/Label";
import { useState, FormEvent } from "react";

import Select from "../../components/form/Select";
import { ArrowRightIcon } from "../../icons";
import { databaseManager } from "../../services/databaseManager.service";
interface Option {
  value: string;
  label: string;
}

export default function MySQL() {
  const [databaseName, setDatabaseName] = useState("");
  const [username, setUsername] = useState("");
  const [databaseNameError, setDatabaseNameError] = useState("");
  const [userNameError, setUsernameError] = useState("");

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
        const data =  await databaseManager.createMySQLDB(databaseName, username)
        console.log(data)
        // success - clear form and errors
        setDatabaseName("");
        setUsername("");
        setDatabaseNameError("");
        setUsernameError("");
        alert(data.message);
      } catch (err) {
        console.error(err);
        setDatabaseNameError("An unexpected error occurred while creating the database");
      }
    })();
  };
  return (
    <>
      <PageMeta title="MySQL" description="Manage MySQL Database" />
      <PageBreadcrumb pageTitle="Manage MySQL Database" />

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
          <DatabaseTable />
        </ComponentCard>
      </div>
    </>
  );
}
