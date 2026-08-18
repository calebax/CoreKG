import { send } from './request'

// 获取所有账户可绑定列表
export async function loadAccountBindingList() {
  return send('account.Bindings', {})
}

// 获取绑定平台的凭证
export async function getBindAccountState(data: {
  provider: string
  redirectUrl: string
}) {
  return send('account.PreConnect', data)
}

// 解除绑定
export async function unBindAccount(id: number) {
  return send('account.Unbind', { id })
}
