import { Response } from "./common.interface"

export interface DatabaseProperties {
  id:number
  name:string
  username:string
}
export interface DatabaseResponse extends Response{
  id:number
}

export interface DatabaseMeta {
  dbName: string;
  tableCount: number;
  databaseSizeMB: number;
  createdAt?: string | null;  // Use ISO string for Date or null
  updatedAt?: string | null;  // Latest table update
  users?: string[];           // Users with privileges (optional)
}
