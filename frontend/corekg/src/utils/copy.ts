export const copyToClipboard = (value: string) => {
  return new Promise<void>((resolve, reject) => {
    try {
      const textarea = document.createElement('textarea')
      textarea.readOnly = true
      textarea.style.position = 'absolute'
      textarea.style.left = '-9999px'
      textarea.value = value
      document.body.appendChild(textarea)
      textarea.select()
      textarea.setSelectionRange(0, textarea.value.length)
      const successful = document.execCommand('copy')
      document.body.removeChild(textarea)
      if (successful) {
        resolve()
      } else {
        reject(new Error('Failed to copy text.'))
      }
    } catch (error) {
      reject(error)
    }
  })
}
