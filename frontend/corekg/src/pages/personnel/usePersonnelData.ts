import type { Department, Employee, PersonnelType } from 'Personnel'
import { produce } from 'immer'
import { match } from 'ts-pattern'
import { create } from 'zustand'
import {
  createDepartment,
  createDepartmentEmployee,
  createDepartmentEmployeePrivate,
  deleteDepartment,
  editDepartmentEmployee,
  editDepartmentEmployeePrivate,
  getPersonnelInfo,
  moveDepartment,
  renameDepartment,
} from '@/api/account'

type DepartmentActionMap = {
  add: Omit<Department, 'id' | 'sort'>
  rename: Pick<Department, 'id'> & { newName: string }
  moveup: Pick<Department, 'id'>
  movedown: Pick<Department, 'id'>
  delete: Pick<Department, 'id'>
}

type EmployeeActionMap = {
  add: Omit<Employee, 'id' | 'employee_id' | 'uin' | 'created_at'>
  edit: Partial<Employee> & Pick<Employee, 'id'>
}

export type PersonnelData = {
  /** 来自后端的原始数据 */
  data: {
    department?: Department[]
    employee?: Employee[]
  }
  /** 传入'employee'时 会加载部门和人员数据 */
  loadData: (type: PersonnelType) => Promise<void>
  dispatchDepartmentAction: (
    action: {
      [K in keyof DepartmentActionMap]: { type: K } & DepartmentActionMap[K] & {
          version?: DeployConfig['version']
        }
    }[keyof DepartmentActionMap],
  ) => Promise<void>
  dispatchEmployeeAction: (
    action: {
      [K in keyof EmployeeActionMap]: { type: K } & EmployeeActionMap[K] & {
          version?: DeployConfig['version']
        }
    }[keyof EmployeeActionMap],
  ) => Promise<void>
}
/** 全局的人事数据 */
export const usePersonnelData = create<PersonnelData>((set, get) => {
  /** immer版本的set */
  const immerSet = (cb: (draft: PersonnelData) => void) => {
    set((val) => {
      return produce(val, cb)
    })
  }

  return {
    data: {},
    loadData: async (type) => {
      try {
        const response = await getPersonnelInfo({
          include_employee: type === 'employee',
        })

        immerSet((draft) => {
          // 映射后端部门数据到前端数据结构
          draft.data.department =
            response?.departments?.map((dept: any) => ({
              id: dept.ID,
              parentId: dept.ParentID,
              name: dept.Name,
              sort: dept.Sort,
            })) || []

          // 映射后端员工数据到前端数据结构
          draft.data.employee =
            response?.employees?.map((emp: any) => ({
              id: emp.employee_id,
              name: emp.name,
              phone: emp.phone,
              email: emp.email,
              departmentIds: emp.department_ids,
              role: emp.sys_role,
              ...emp,
            })) || []
        })
      } catch (error) {
        console.error('Failed to load personnel data:', error)
        // 如果接口失败，就设置为空数组
        immerSet((draft) => {
          draft.data.department = []
        })
        if (type === 'employee') {
          immerSet((draft) => {
            draft.data.employee = []
          })
        }
      }
    },
    dispatchDepartmentAction: async (action) => {
      const { department } = get().data
      if (!department) return
      switch (action.type) {
        case 'add': {
          const { parentId, name } = action
          const {
            department: { ID: id, Sort: sort },
          } = await createDepartment({
            name,
            parent_id: parentId as any,
          })
          immerSet((draft) => {
            draft.data.department!.push({
              id,
              parentId,
              name,
              sort,
            })
          })
          break
        }
        case 'rename': {
          const { id, newName } = action
          await renameDepartment({ id: id as any, name: newName })
          immerSet((draft) => {
            const target = draft.data.department?.find((item) => item.id === id)
            if (!target) return
            target.name = newName
          })
          break
        }
        case 'moveup':
        case 'movedown': {
          const { id } = action
          const target = department.find((item) => item.id === id)
          if (!target) return
          const brotherDepartments = department
            .filter((item) => item.parentId === target?.parentId)
            .sort((v1, v2) => v1.sort - v2.sort)
          const position = brotherDepartments.findIndex(
            (item) => item.id === id,
          )
          // 目标位置前后的元素
          const [preIndex, postIndex] =
            action.type === 'movedown'
              ? [position + 1, position + 2]
              : [position - 2, position - 1]
          const pre = brotherDepartments[preIndex] ?? { id: 0, sort: 0 }
          const post = brotherDepartments[postIndex] ?? {
            id: 0,
            sort: 2 * pre.sort,
          }
          await moveDepartment({
            department_id: id as any,
            pre_id: pre.id as any,
            post_id: post.id as any,
          })
          immerSet((draft) => {
            const { department } = draft.data
            const target = department!.find((item) => item.id === id)
            if (!target) return
            target.sort = (pre.sort + post.sort) / 2
          })
          break
        }
        case 'delete': {
          const { id } = action
          await deleteDepartment({ id: id as any })
          immerSet((draft) => {
            const { department } = draft.data
            const index = department!.findIndex((item) => item.id === id)
            if (index === -1) return
            draft.data.department!.splice(index, 1)
          })
          break
        }
      }
    },
    dispatchEmployeeAction: async (action) => {
      if (!get().data.employee) return
      switch (action.type) {
        case 'add': {
          const { ...info } = action
          const body = {
            employee: {
              ...info,
              sys_role: info.role,
              department_ids: info.departmentIds as number[],
            },
          }
          const {
            employee: { uin, employee_id, created_at },
          } = await match(action.version)
            .with('custom', () => createDepartmentEmployeePrivate(body))
            .otherwise(() => createDepartmentEmployee(body))
          immerSet((draft) => {
            draft.data.employee?.push({
              id: employee_id,
              uin,
              employee_id,
              created_at,
              ...info,
            })
          })
          break
        }
        case 'edit': {
          const { id, ...info } = action
          const { uin, employee_id } = get().data.employee!.find(
            (item) => item.id === id,
          )!

          const { departmentIds, ...rest } = info
          const body = {
            employee: {
              ...rest,
              ...(departmentIds
                ? { department_ids: departmentIds as number[] }
                : {}),
              uin,
              employee_id,
            },
          }
          if (action.version === 'custom') {
            await editDepartmentEmployeePrivate(body)
          } else {
            await editDepartmentEmployee(body)
          }
          immerSet((draft) => {
            const { employee } = draft.data
            const index = employee!.findIndex((item) => item.id === id)
            if (index === -1) return
            employee![index] = {
              ...employee![index],
              ...info,
            }
          })
          break
        }
      }
    },
  }
})
