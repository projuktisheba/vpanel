import { Response } from "./common.interface";

export interface DatabaseProperties {
  id: number;
  name: string;
  username: string;
}
export interface DatabaseResponse extends Response {
  id: number;
}

export interface DBUser {
  id: number;
  username: string;
  password: string;
  createdAt?: string | null; // Use ISO string for Date or null
  updatedAt?: string | null; // Latest table update
  deletedAt?: string | null; // Latest table update
}
export interface Database {
  id: number;
  dbName: string;
  tableCount: number;
  databaseSizeMB: number;
  createdAt?: string | null; // Use ISO string for Date or null
  updatedAt?: string | null; // Latest table update
  deletedAt?: string | null; // Latest table update
  user?: DBUser; // Users with privileges (optional)
}
