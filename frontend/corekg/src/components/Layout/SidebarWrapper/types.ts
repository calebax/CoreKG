import type { PropsWithChildren } from 'react'

export enum EExpandableStatus {
  EXPAND = 'expand',
  FOLD = 'fold',
}

export enum ERole {
  ADMIN = 'sys_admin',
  MEMBER = 'sys_employee',
}

export interface ISidebarWrapperProps extends PropsWithChildren {
  expandableStatus: EExpandableStatus
  updateExpandableStatus: (item: EExpandableStatus) => void
}
