import { message, Modal } from 'antd'

let errorMessageTimer: NodeJS.Timeout | null = null
let isShowingError: boolean = false
export const showRequestErrorOnce = (msg: string) => {
  if (!isShowingError) {
    isShowingError = true
    message.error(msg)
    if (errorMessageTimer) {
      clearTimeout(errorMessageTimer)
    }
    errorMessageTimer = setTimeout(() => {
      isShowingError = false
    }, 2000)
  }
}

export const showNotification = (msg: string) => {
  message.success(msg)
}

export const handleConfirm = (
  title: string,
  content: string,
  onOk: () => void,
) => {
  Modal.confirm({
    title,
    content,
    onOk,
  })
}
