/*
 * @Descripttion: localStorage
 */

const TOKEN = 'token'

export default {
  setToken: (payload: string) => {
    window.localStorage.setItem(TOKEN, payload)
  },
  getToken: (): string | null => {
    return window.localStorage.getItem(TOKEN)
  },
  rmToken: () => {
    window.localStorage.removeItem(TOKEN)
  },
}
