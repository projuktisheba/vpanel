import { Database } from "./database.interface";
import { Domain } from "./domain.interface";

export interface Project {
  id: number;
  projectName: string;
  domainName: string;
  dbName: string;
  projectFramework: string;
  templatePath: string;
  projectDirectory: string;
  status: string;
  createdAt: string;
  updatedAt: string;
  domainInfo: Domain;
  databaseInfo: Database
}