// 获取微信登录二维码
export const getWxLoginCode = (appId, callback = '/callback') => {
  // let redirect
  // if (type === 'login') {
  //   redirect = config.domain + config.baseUrl + 'callback?url=' + url
  // } else if (type === 'bindWx') {
  //   redirect = config.domain + config.baseUrl + 'bindWx?url=' + url
  // } else {
  //   redirect = config.domain + config.baseUrl + 'changeWx?url=' + url
  // }
  const redirect = location.origin + callback
  // const redirect = 'https://account.example.com/callback'
  // console.log('redirect', appId, redirect)
  return new WxLogin({
    id: 'wxQrcode',
    appid: appId,
    scope: 'snsapi_login',
    redirect_uri: redirect,
    state: Math.ceil(Math.random() * 1000000),
    style: 'black',
    href: location.origin + '/code.css',
  })
}

// 加载微信js
export const loadWxLoginScript = () => {
  return new Promise((resolve, reject) => {
    const script = document.createElement('script')
    script.src = 'https://res.wx.qq.com/connect/zh_CN/htmledition/js/wxLogin.js'
    script.onload = () => {
      resolve()
    }
    script.onerror = (e) => {
      reject(e)
    }
    document.head.appendChild(script)
  })
}
