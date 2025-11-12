import {
  Table,
  TableBody,
  TableCell,
  TableHeader,
  TableRow,
} from "../../ui/table";

import Badge from "../../ui/badge/Badge";
import { useState, useMemo } from "react";
import Select from "../../form/Select";
import Input from "../../form/input/InputField";
import { PlusIcon } from "../../../icons";
import Button from "../../ui/button/Button";

interface Order {
  id: number;
  user: {
    image: string;
    name: string;
    role: string;
  };
  projectName: string;
  team: {
    images: string[];
  };
  status: string;
  budget: string;
}

// Define the table data using the interface
const tableData: Order[] = [
  {
    id: 1,
    user: {
      image: "/images/user/user-17.jpg",
      name: "Lindsey Curtis",
      role: "Web Designer",
    },
    projectName: "Agency Website",
    team: {
      images: [
        "/images/user/user-22.jpg",
        "/images/user/user-23.jpg",
        "/images/user/user-24.jpg",
      ],
    },
    budget: "3.9K",
    status: "Active",
  },
  {
    id: 2,
    user: {
      image: "/images/user/user-18.jpg",
      name: "Kaiya George",
      role: "Project Manager",
    },
    projectName: "Technology",
    team: {
      images: ["/images/user/user-25.jpg", "/images/user/user-26.jpg"],
    },
    budget: "24.9K",
    status: "Pending",
  },
  {
    id: 3,
    user: {
      image: "/images/user/user-17.jpg",
      name: "Zain Geidt",
      role: "Content Writing",
    },
    projectName: "Blog Writing",
    team: {
      images: ["/images/user/user-27.jpg"],
    },
    budget: "12.7K",
    status: "Active",
  },
  {
    id: 4,
    user: {
      image: "/images/user/user-20.jpg",
      name: "Abram Schleifer",
      role: "Digital Marketer",
    },
    projectName: "Social Media",
    team: {
      images: [
        "/images/user/user-28.jpg",
        "/images/user/user-29.jpg",
        "/images/user/user-30.jpg",
      ],
    },
    budget: "2.8K",
    status: "Cancel",
  },
  {
    id: 5,
    user: {
      image: "/images/user/user-21.jpg",
      name: "Carla George",
      role: "Front-end Developer",
    },
    projectName: "Website",
    team: {
      images: [
        "/images/user/user-31.jpg",
        "/images/user/user-32.jpg",
        "/images/user/user-33.jpg",
      ],
    },
    budget: "4.5K",
    status: "Active",
  },
];

export default function DatabaseTable() {
  const [searchColumn, setSearchColumn] = useState("database");
  const [searchQuery, setSearchQuery] = useState("");

  // ============ Filtering ============

  //selection options
  const options = [
    { value: "database", label: "Database" },
    { value: "user", label: "User" },
  ];
  const filteredData = useMemo(() => {
    if (!searchQuery.trim()) return tableData;
    const query = searchQuery.toLowerCase();

    return tableData.filter((order) => {
      switch (searchColumn) {
        case "user":
          return (
            order.user.name.toLowerCase().includes(query) ||
            order.user.role.toLowerCase().includes(query)
          );
        case "database":
          return order.projectName.toLowerCase().includes(query);
        default:
          return true;
      }
    });
  }, [searchQuery, searchColumn]);

  // ============ Print ============
  const handlePrint = () => {
    // Create a hidden iframe
    const iframe = document.createElement("iframe");
    iframe.style.position = "fixed";
    iframe.style.right = "0";
    iframe.style.bottom = "0";
    iframe.style.width = "0";
    iframe.style.height = "0";
    iframe.style.border = "0";
    document.body.appendChild(iframe);

    const win = iframe.contentWindow;
    if (!win) {
      iframe.remove();
      return;
    }
    const doc = win.document;

    // Generate table rows
    const tableRows = filteredData
      .map(
        (o, i) => `
        <tr>
          <td>${i + 1}</td>
          <td>
            <strong>${o.user.name}</strong><br/>
            <span style="font-size: 12px; color: #6b7280;">${o.user.role}</span>
          </td>
          <td>${o.projectName}</td>
          <td>
            <span style="
              padding: 4px 8px;
              border-radius: 6px;
              background-color: ${
                o.status === "Completed"
                  ? "#dcfce7"
                  : o.status === "Pending"
                  ? "#fef9c3"
                  : "#fee2e2"
              };
              color: ${
                o.status === "Completed"
                  ? "#15803d"
                  : o.status === "Pending"
                  ? "#854d0e"
                  : "#b91c1c"
              };
              font-weight: 500;
            ">
              ${o.status}
            </span>
          </td>
          <td style="text-align: right;">$${Number(
            o.budget
          ).toLocaleString()}</td>
        </tr>`
      )
      .join("");

    // Write HTML to iframe
    doc.open();
    doc.write(`
    <html>
      <head>
        <title>Print Orders</title>
        <style>
          @page { margin: 20mm; }
          body {
            font-family: 'Inter', sans-serif;
            background-color: #f9fafb;
            color: #111827;
            padding: 20px;
          }
          h2 {
            text-align: center;
            color: #0f172a;
            margin-bottom: 20px;
            font-size: 22px;
          }
          table {
            width: 100%;
            border-collapse: collapse;
            background-color: #ffffff;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
            border-radius: 8px;
            overflow: hidden;
          }
          thead th {
            background-color: #1e293b;
            color: #f8fafc;
            font-weight: 600;
            padding: 10px;
            font-size: 14px;
            text-align: left;
          }
          tbody td {
            border-bottom: 1px solid #e5e7eb;
            padding: 10px;
            font-size: 14px;
          }
          tbody tr:nth-child(even) {
            background-color: #f9fafb;
          }
          tbody tr:last-child td {
            border-bottom: none;
          }
          .footer {
            margin-top: 30px;
            text-align: center;
            color: #6b7280;
            font-size: 12px;
          }
        </style>
      </head>
      <body>
        <h2>Filtered Orders</h2>
        <table>
          <thead>
            <tr>
              <th>#</th>
              <th>User</th>
              <th>Project</th>
              <th>Status</th>
              <th style="text-align: right;">Budget</th>
            </tr>
          </thead>
          <tbody>${tableRows}</tbody>
        </table>
        <div class="footer">
          Generated by vPanel — ${new Date().toLocaleString()}
        </div>
      </body>
    </html>
  `);
    doc.close();

    // Wait for iframe content to load, then print
    iframe.onload = () => {
      const w = iframe.contentWindow ?? win;
      if (!w) {
        iframe.remove();
        return;
      }
      w.focus();
      w.print();
      setTimeout(() => iframe.remove(), 1000);
    };
  };

  return (
    <div className="overflow-hidden rounded-xl border border-gray-200 bg-white dark:border-white/[0.05] dark:bg-white/[0.03]">
      {/* ======== Controls (Search + Print) ======== */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 px-5 py-4 border-b border-gray-200 dark:border-white/[0.05]">
        <div className="flex items-center gap-2">
          <Select
            options={options}
            placeholder="Select Option"
            onChange={setSearchColumn}
            defaultValue={options[0].value}
            className="dark:bg-dark-900"
          />
          <Input
            type="text"
            value={searchQuery}
            onChange={(e) => {
              setSearchQuery(e.target.value);
            }}
            placeholder={`Search by ${searchColumn}...`}
            className="min-w-max"
          />
          <Button size="sm" variant="primary" onClick={handlePrint}>
            Print
          </Button>
        </div>
        <div className="flex items-center gap-2">
          <Button size="sm" variant="warning" startIcon={<PlusIcon />}>
            Add New Table
          </Button>
        </div>
      </div>

      <div className="max-w-full overflow-x-auto">
        <Table>
          {/* Table Header */}
          <TableHeader className="border-b border-gray-100 dark:border-white/[0.05]">
            <TableRow>
              <TableCell
                isHeader
                className="px-5 py-3 font-medium text-gray-500 text-start text-theme-xs dark:text-gray-400"
              >
                User
              </TableCell>
              <TableCell
                isHeader
                className="px-5 py-3 font-medium text-gray-500 text-start text-theme-xs dark:text-gray-400"
              >
                Project Name
              </TableCell>
              <TableCell
                isHeader
                className="px-5 py-3 font-medium text-gray-500 text-start text-theme-xs dark:text-gray-400"
              >
                Team
              </TableCell>
              <TableCell
                isHeader
                className="px-5 py-3 font-medium text-gray-500 text-start text-theme-xs dark:text-gray-400"
              >
                Status
              </TableCell>
              <TableCell
                isHeader
                className="px-5 py-3 font-medium text-gray-500 text-start text-theme-xs dark:text-gray-400"
              >
                Budget
              </TableCell>
            </TableRow>
          </TableHeader>

          {/* Table Body */}
          <TableBody className="divide-y divide-gray-100 dark:divide-white/[0.05]">
            {filteredData.map((order) => (
              <TableRow key={order.id}>
                <TableCell className="px-5 py-4 sm:px-6 text-start">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 overflow-hidden rounded-full">
                      <img
                        width={40}
                        height={40}
                        src={order.user.image}
                        alt={order.user.name}
                      />
                    </div>
                    <div>
                      <span className="block font-medium text-gray-800 text-theme-sm dark:text-white/90">
                        {order.user.name}
                      </span>
                      <span className="block text-gray-500 text-theme-xs dark:text-gray-400">
                        {order.user.role}
                      </span>
                    </div>
                  </div>
                </TableCell>
                <TableCell className="px-4 py-3 text-gray-500 text-start text-theme-sm dark:text-gray-400">
                  {order.projectName}
                </TableCell>
                <TableCell className="px-4 py-3 text-gray-500 text-start text-theme-sm dark:text-gray-400">
                  <div className="flex -space-x-2">
                    {order.team.images.map((teamImage, index) => (
                      <div
                        key={index}
                        className="w-6 h-6 overflow-hidden border-2 border-white rounded-full dark:border-gray-900"
                      >
                        <img
                          width={24}
                          height={24}
                          src={teamImage}
                          alt={`Team member ${index + 1}`}
                          className="w-full size-6"
                        />
                      </div>
                    ))}
                  </div>
                </TableCell>
                <TableCell className="px-4 py-3 text-gray-500 text-start text-theme-sm dark:text-gray-400">
                  <Badge
                    size="sm"
                    color={
                      order.status === "Active"
                        ? "success"
                        : order.status === "Pending"
                        ? "warning"
                        : "error"
                    }
                  >
                    {order.status}
                  </Badge>
                </TableCell>
                <TableCell className="px-4 py-3 text-gray-500 text-theme-sm dark:text-gray-400">
                  {order.budget}
                </TableCell>
              </TableRow>
            ))}

            {filteredData.length === 0 && (
              <TableRow>
                <TableCell
                  colSpan={5}
                  className="py-5 text-center text-gray-500 dark:text-gray-400"
                >
                  No {searchColumn} found
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
