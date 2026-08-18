import { send, upload } from './request'

export interface OrganizationProfile {
  id: string
  name: string
  logo: string
  company_quota?: number
}

export interface UpdateOrganizationPayload {
  name: string
}

// 上传组织头像
export const uploadAvatarImg = (
  data: { file: File; purpose: string } | FormData,
  config?: any,
) => upload('account.UploadOrganizeLogo', data, config)

// 上传website头像
export const uploadWebsiteLogo = (
  data: { file: File; purpose: string } | FormData,
  config?: any,
) => upload('account.UploadWebSiteLogo', data, config)

// 获取组织信息
export const fetchOrganizationProfile = () =>
  send('account.GetCompanyInfo', {}) as Promise<OrganizationProfile>

// 更新组织信息
export const updateOrganizationProfile = (payload: UpdateOrganizationPayload) =>
  send('account.EditCompanyInfo', payload)

export interface GlobalInfo {
  website_info: { website_logo: string; website_name: string }
}

// 获取全局信息（网站logo和名称）
export const getGlobalInfo = () =>
  send('account.GetGlobalInfo', {}) as Promise<GlobalInfo>
