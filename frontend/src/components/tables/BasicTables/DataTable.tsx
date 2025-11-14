import {
  Table,
  TableBody,
  TableCell,
  TableHeader,
  TableRow,
} from "../../ui/table";
import { useState, useMemo, ReactNode } from "react";
import Select from "../../form/Select";
import Input from "../../form/input/InputField";
import Button from "../../ui/button/Button";

interface Column {
  key: string;
  label: string;
  className: string;
  render?: (row: any) => ReactNode; // Optional custom render
}

interface DataTableProps {
  data: any[];
  columns: Column[];
  searchOptions?: { value: string; label: string }[];
  defaultSearchColumn?: string;
  onAddClick?: () => void;
  addButtonLabel?: string;
}

export default function DataTable({
  data,
  columns,
  searchOptions,
  defaultSearchColumn,
}: DataTableProps) {
  const [searchColumn, setSearchColumn] = useState(
    defaultSearchColumn || searchOptions?.[0]?.value || ""
  );
  const [searchQuery, setSearchQuery] = useState("");

  // ============ Filtering ============
  // Helper to get nested value by path
const getValueByPath = (obj: any, path: string) => {
  return path.split(".").reduce((acc, key) => (acc ? acc[key] : undefined), obj);
};
const filteredData = useMemo(() => {
  if (!searchQuery.trim()) return data;
  const query = searchQuery.toLowerCase();

  return data.filter((row) => {
    const value = getValueByPath(row, searchColumn); // <-- updated

    if (value == null) return false;

    if (typeof value === "string") return value.toLowerCase().includes(query);
    if (typeof value === "number") return value.toString().includes(query);

    if (Array.isArray(value)) return value.some((item) => item?.toString().toLowerCase().includes(query));

    if (typeof value === "object") return Object.values(value).join(" ").toLowerCase().includes(query);

    return false;
  });
}, [searchQuery, searchColumn, data]);

  // ============ Print ============
 const handlePrint = () => {
  
  // Create hidden iframe
  const iframe = document.createElement("iframe");
  iframe.style.position = "fixed";
  iframe.style.top = "0";
  iframe.style.left = "0";
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

  // Dynamically build table headers
  const headers = columns.map((col) => `<th>${col.label}</th>`).join("");

  // Safe function to extract text from ReactNode
  const extractText = (node: ReactNode): string => {
    if (node == null) return "";
    if (typeof node === "string" || typeof node === "number" || typeof node === "boolean")
      return node.toString();
    if (Array.isArray(node)) return node.map(extractText).join(" ");
    if (
      typeof node === "object" &&
      "props" in (node as any) &&
      (node as any).props?.children
    ) {
      return extractText((node as any).props.children);
    }
    return "";
  };

  // Build table rows
  const tableRows = filteredData
    .map(
      (row) => `
        <tr>
          ${columns
            .map((col) => {
              let cellValue = "";

              if (col.render) {
                try {
                  const rendered = col.render(row);
                  cellValue = extractText(rendered);
                } catch {
                  cellValue = "-";
                }
              } else {
                const val = row[col.key];
                if (typeof val === "object" && val !== null)
                  cellValue = Object.values(val).join(" ");
                else cellValue = val ?? "-";
              }

              return `<td style="text-align:left;">${cellValue}</td>`;
            })
            .join("")}
        </tr>`
    )
    .join("");

  // Write HTML into iframe
  doc.open();
  doc.write(`
    <html>
      <head>
        <title>Print Table</title>
        <style>
          @page { margin: 15mm; }
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
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
          }
          th, td {
            padding: 10px;
            border-bottom: 1px solid #e5e7eb;
            font-size: 14px;
            text-align: left;
          }
          thead th {
            background-color: #1e293b;
            color: #f8fafc;
            font-weight: 600;
          }
          tbody tr:nth-child(even) { background-color: #f9fafb; }
          tbody tr:last-child td { border-bottom: none; }
          .footer {
            margin-top: 30px;
            text-align: center;
            color: #6b7280;
            font-size: 12px;
          }
        </style>
      </head>
      <body>
        <h2>Filtered Table Data</h2>
        <table>
          <thead><tr>${headers}</tr></thead>
          <tbody>${tableRows}</tbody>
        </table>
        <div class="footer">
          Generated by vPanel — ${new Date().toLocaleString()}
        </div>
      </body>
    </html>
  `);
  doc.close();

  // Give it time to render before printing
  setTimeout(() => {
    win.focus();
    win.print();
    setTimeout(() => iframe.remove(), 1000);
  }, 300);
};


  return (
    <div className="overflow-hidden rounded-xl border border-gray-200 bg-white dark:border-white/[0.05] dark:bg-white/[0.03]">
      {/* ======== Controls (Search + Print) ======== */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 px-5 py-4 border-b border-gray-200 dark:border-white/[0.05]">
        {searchOptions && (
          <div className="flex items-center gap-2">
            <Select
              options={searchOptions}
              placeholder="Select Option"
              onChange={setSearchColumn}
              defaultValue={searchColumn}
              className="dark:bg-dark-900"
            />
            <Input
              type="text"
              value={searchQuery}
              onChange={(e) => {
                setSearchQuery(e.target.value);
              }}
              placeholder={`Search by ${
                searchOptions?.find((opt) => opt.value === searchColumn)
                  ?.label || ""
              }...`}
              className="min-w-max"
            />
            <Button size="sm" variant="primary" onClick={handlePrint}>
              Print
            </Button>
          </div>
        )}
      </div>

      {/* Table */}
      <div className="max-w-full overflow-x-auto">
        <Table>
          {/* Table Header */}
          <TableHeader className="border-b border-gray-100 dark:border-white/[0.05]">
            <TableRow>
              {columns.map((col) => (
                <TableCell
                  key={col.key}
                  isHeader
                  className="px-5 py-3 font-medium text-gray-500 text-start text-theme-xs dark:text-gray-400"
                >
                  {col.label}
                </TableCell>
              ))}
            </TableRow>
          </TableHeader>

          {/* Table Body */}
          <TableBody className="divide-y divide-gray-100 dark:divide-white/[0.05]">
            {filteredData.map((row, idx) => (
              <TableRow key={idx}>
                {columns.map((col) => (
                  <TableCell
                    key={col.key}
                    className={`px-5 py-4 sm:px-6 text-start ${col.className}`}
                  >
                    {col.render ? col.render(row) : row[col.key]}
                  </TableCell>
                ))}
              </TableRow>
            ))}

            {filteredData.length === 0 && (
              <TableRow>
                <TableCell
                  colSpan={5}
                  className="py-5 text-center text-gray-500 dark:text-gray-400"
                >
                  No{" "}
                  {searchOptions?.find((opt) => opt.value === searchColumn)
                    ?.label || "results"}{" "}
                  found
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
