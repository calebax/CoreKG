export type PersonnelType = 'department' | 'employee'

export type Department = {
  id: number | string
  parentId?: number | string
  name: string
  sort: number
}

export type Employee = {
  id: number | string
  created_at: string
  name: string
  user_name: string
  phone?: string
  email?: string
  /** 部门id */
  departmentIds: Department['id'][]
  role: 'sys_admin' | 'sys_employee'
  uin: number
  employee_id: number
}

export * from './usePersonnelData'
