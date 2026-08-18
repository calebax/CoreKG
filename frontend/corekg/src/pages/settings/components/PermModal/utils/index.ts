import { produce } from 'immer'

export type Perm = {
  manage_perm: boolean
  use_perm: boolean
}
/**
 * 将指定权限取反后得到更新后的对象\
 * 有管理权限必定有使用权限\
 * 勾选管理权限时也勾选使用权限
 */
export function getUpdatedPerm<T extends Perm>(
  originValue: T,
  /** 更新管理权限或使用权限 */
  key: 'manage_perm' | 'use_perm',
): T {
  return produce(originValue, (draft) => {
    if (key === 'use_perm') {
      if (!draft.manage_perm) {
        draft.use_perm = !draft.use_perm
      } else {
        // 有管理权限时必定有使用权限
        draft.use_perm = true
      }
      return
    }
    if (draft.manage_perm) {
      draft.manage_perm = false
    } else {
      // 勾选管理权限时也勾选使用权限
      draft.manage_perm = draft.use_perm = true
    }
  })
}
