export interface Response {
    error:boolean
    message:string
}

export interface UploadProgress {
  /** Number of chunks uploaded so far */
  uploadedChunks: number;

  /** Total number of chunks to be uploaded */
  totalChunks: number;

  /** Percentage of completion (0-100) */
  percentage: number;
}