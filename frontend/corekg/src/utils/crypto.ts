import CryptoJS from 'crypto-js'

// 加密密钥和初始向量
const key = CryptoJS.enc.Utf8.parse('G2HX+F+411lK/O8c')
const iv = key

/**
 * 加密函数，与 Go 的 AesEncrypt 兼容
 * @param plaintext - 待加密的明文字符串
 * @returns Base64 编码的密文字符串
 */
export const aesEncrypt = (plaintext: string): string => {
  // 加密数据
  const encrypted = CryptoJS.AES.encrypt(plaintext, key, {
    iv: iv,
    mode: CryptoJS.mode.CBC,
    padding: CryptoJS.pad.Pkcs7,
  })

  return 'keep-enc-' + encrypted.toString()
}

/**
 * 解密函数，与 Go 的 AesDecrypt 兼容
 * @param ciphertext - Base64 编码的密文字符串
 * @returns 解密后的明文字符串
 */
export const aesDecrypt = (ciphertext: string): string => {
  let actualCiphertext = ciphertext
  if (ciphertext.startsWith('keep-enc-')) {
    actualCiphertext = ciphertext.substring(10)
  }

  // 解密数据
  const decrypted = CryptoJS.AES.decrypt(actualCiphertext, key, {
    iv: iv,
    mode: CryptoJS.mode.CBC,
    padding: CryptoJS.pad.Pkcs7,
  })

  // 将解密后的数据转换为 UTF-8 字符串
  return decrypted.toString(CryptoJS.enc.Utf8)
}

/**
 * 加密密码的便捷函数
 * @param password - 明文密码
 * @returns 加密后的密码
 */
export const encryptPassword = (password: string): string => {
  return aesEncrypt(password)
}

/**
 * 解密密码的便捷函数
 * @param encryptedPassword - 加密后的密码
 * @returns 明文密码
 */
export const decryptPassword = (encryptedPassword: string): string => {
  return aesDecrypt(encryptedPassword)
}
