import { send, send2 } from './request'

export const checkLicense = (body: any) => send('corekg.CheckLicense', body)

export const registerLicense = (data: {
  license: string
  company_name: string
  company_logo: string
  website_info: {
    website_logo: string
    website_name: string
  }
}) => send('corekg.RegisterLicense', data)

export const getLicenseInfo = (body: any) =>
  send('corekg.GetLicenseInfo', body) as Promise<{
    meta?: {
      id?: number
      serial?: string
      issuer?: string
      created_at?: string
      expired_at?: string
    }
    status: number
    valid_days: number
    modules?: CustomModule[]
  }>

export const getClusterID = (body: any) => send2('status.GetClusterID', body)
