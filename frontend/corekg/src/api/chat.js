import config from '@/config'
import globalConfig from '@/config'
import useLocalStore from '@/stores/local'
import { toLogin } from './request'

const { withPrefix } = globalConfig
export const chat = (url, data, options = {}) => {
  const token = useLocalStore.getState().token
  if (token) {
    const apiUrl = config.apiUrl + withPrefix(url)
    const body = {
      cmd: withPrefix(url),
      env: config.apiEnv,
      version: config.version,
      request: data,
    }
    return fetch(apiUrl, {
      method: 'POST',
      headers: {
        Env: config.apiEnv,
        Authorization: `Bearer ${token}`,
        'Accept-Language': window.__LANG,
      },
      body: JSON.stringify(body),
      ...options,
    })
  } else {
    toLogin()
  }
}
export const chatFile = (url, data, options = {}) => {
  const token = useLocalStore.getState().token
  if (token) {
    const apiUrl = config.apiUrl + withPrefix(url)
    const formData = new FormData()
    for (const key in data) {
      formData.append(key, data[key])
    }
    return fetch(apiUrl, {
      method: 'POST',
      headers: {
        Env: config.apiEnv,
        Authorization: `Bearer ${token}`,
        'Accept-Language': window.__LANG,
      },
      body: formData,
      ...options,
    })
  } else {
    toLogin()
  }
}
export const chat1 = (url, data) => {
  const token = useLocalStore.getState().token
  const apiUrl = config.apiUrl + url
  return fetch(apiUrl, {
    method: 'POST',
    headers: {
      Env: config.apiEnv,
      Authorization: `Bearer ${token}`,
      // 'Content-Type': 'multipart/form-data',
      'Accept-Language': window.__LANG,
    },
    body: data,
  })
}
export const chat2 = async (url, data) => {
  const baseLocalStore = useBaseLocalStore()
  if (baseLocalStore.token) {
    try {
      const apiUrl = config.apiUrl + url
      const body = {
        cmd: url,
        env: config.apiEnv,
        request: data,
      }
      const response = await fetch(apiUrl, {
        method: 'POST',
        headers: {
          Env: config.apiEnv,
          Authorization: `Bearer ${baseLocalStore.token}`,
          'Accept-Language': window.__LANG,
        },
        body: JSON.stringify(body),
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const reader = response.body.getReader()

      return new ReadableStream({
        start(controller) {
          function pump() {
            return reader.read().then(({ done, value }) => {
              if (done) {
                controller.close()
                return
              }
              controller.enqueue(value)
              return pump()
            })
          }
          return pump()
        },
      })
    } catch (error) {
      console.error('Fetch error:', error)
      throw error
    }
  } else {
    toLogin()
  }
}

/** chat
try {
    const response = await writePatentDraft(body)
    const reader = response.body.getReader()
    const decoder = new TextDecoder()
    while (true) {
      const { done, value } = await reader.read()
      if (done) {
        sendLoading.value = false
        writeLocalStore.setSendLoading(false)
        scrollToEnd()
        break
      }
      const text = decoder.decode(value, { stream: true })
      if (isJsonString(text)) {
        const val = JSON.parse(text).content
        updateAnswerValue(item.id, val)
        scrollToEnd()
      }
    }
  } catch (error) {
    sendLoading.value = false
    writeLocalStore.setSendLoading(false)
  }
*/

/**  chat2
writePatentDraft(body).then(stream => {
  const reader = stream.getReader()
  const decoder = new TextDecoder()
  let count = 0
  reader.read().then(function processText ({ done, value }) {
    if (done) {
      sendLoading.value = false
      writeLocalStore.setSendLoading(false)
      scrollToEnd()
      return
    }
    const text = decoder.decode(value, { stream: true })
    console.log(count, text)
    count++
    if (isJsonString(text)) {
      const val = JSON.parse(text).content
      updateAnswerValue(item.id, val)
      scrollToEnd()
    }
    reader.read().then(processText)
  })
}).catch(() => {
  sendLoading.value = false
  writeLocalStore.setSendLoading(false)
})
*/
