import default1 from './images/default1.svg'
import default2 from './images/default2.png'
import default3 from './images/default3.svg'
import default4 from './images/default4.svg'
import default5 from './images/default5.svg'
import default6 from './images/default6.svg'

/** 如果头像形如'default1' 就将其转化为默认头像 */
export const getAvatar = (avatar?: string) => {
  if (!avatar) return default1
  const defaultAvatarMap: Record<string, string> = {
    default1,
    default2,
    default3,
    default4,
    default5,
    default6,
  }
  const defaultAvatar = defaultAvatarMap[avatar]
  return defaultAvatar ?? avatar
}
