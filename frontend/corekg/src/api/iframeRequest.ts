import axios from 'axios'
import globalConfig from '@/config'

const iframeRequest = axios.create({
  baseURL: globalConfig.apiUrl,
  timeout: 600000000,
  headers: {
    Env: globalConfig.apiEnv,
    'Content-Type': 'application/json',
  },
})

export const iframeSend = async (
  url: string,
  data: any,
  token?: string,
): Promise<any> => {
  const body = {
    cmd: globalConfig.withPrefix(url),
    env: globalConfig.apiEnv,
    version: globalConfig.version,
    request: data,
  }

  const headers: any = {
    Env: globalConfig.apiEnv,
    'Content-Type': 'application/json',
  }

  if (token) {
    headers.Authorization = `Bearer ${token}`
  }

  try {
    const response = await iframeRequest.post(
      globalConfig.withPrefix(url),
      body,
      { headers },
    )
    const { data: responseData } = response
    if (responseData.code === 0) {
      return responseData.Response
    } else {
      throw new Error(responseData.message || 'API调用失败')
    }
  } catch (error) {
    console.error('iframe API调用失败:', error)
    throw error
  }
}

export const iframeChat = async (
  url: string,
  data: any,
  token: string,
  options = {},
) => {
  const apiUrl = globalConfig.apiUrl + globalConfig.withPrefix(url)
  const body = {
    cmd: globalConfig.withPrefix(url),
    env: globalConfig.apiEnv,
    version: globalConfig.version,
    request: data,
  }

  return fetch(apiUrl, {
    method: 'POST',
    headers: {
      Env: globalConfig.apiEnv,
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(body),
    ...options,
  })
}
