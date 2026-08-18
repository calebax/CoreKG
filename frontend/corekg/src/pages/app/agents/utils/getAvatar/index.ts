import { BasicAgentInfo } from 'Agent'
import { getDefaultAvatar } from '../../index/getDefaultAvatar'

export const getAvatar = (src: string, type: BasicAgentInfo['type']) => {
  if (src === 'default') return getDefaultAvatar(type)
  return src
}
