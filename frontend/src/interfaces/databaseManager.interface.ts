import { Response } from "./common.interface"

export interface DatabaseProperties {
  id:number
  name:string
  username:string
}
export interface DatabaseResponse extends Response{
  id:number
}