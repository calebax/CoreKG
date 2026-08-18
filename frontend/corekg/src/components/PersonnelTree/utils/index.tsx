/* eslint-disable react-refresh/only-export-components */
import { useState, useEffect, useRef, useMemo } from 'react'
import type {
  Dispatch,
  FC,
  PropsWithChildren,
  ReactNode,
  SetStateAction,
} from 'react'
import { App, Input, InputRef, Popover } from 'antd'
import { LoadingOutlined } from '@ant-design/icons'
import { usePersonnelData } from 'Personnel'
import { useClickAway } from 'ahooks'
import { match, P } from 'ts-pattern'
import { cn } from '@/utils'
import { useDeployConfig } from '@/utils/useDeployConfig'
import type { PersonnelTreeNode, PersonnelTree } from '..'
import Add from './images/add.svg?react'
import Del from './images/del.svg?react'
import DepartmentIcon from './images/department.svg?react'
import EmployeeIcon from './images/employee.svg?react'
import MoveDown from './images/movedown.svg?react'
import MoveUp from './images/moveup.svg?react'
import OperatorIcon from './images/operator.svg?react'
import Rename from './images/rename.svg?react'

/** 筛选匹配的树节点  */
export const filterTreeBySearch = (
  tree: PersonnelTreeNode[],
  config: {
    search?: string
    /** 必须被保留的节点 */
    preservedKeys?: any[]
  } = {},
): PersonnelTreeNode[] => {
  const { search, preservedKeys } = config
  if (!search) {
    return tree
  }
  const _filter = (node: PersonnelTreeNode): PersonnelTreeNode | null => {
    // 本节点匹配
    if (node.name.includes(search) || preservedKeys?.includes(node.key)) {
      return node
    }
    const filteredChildren = node.children?.map(_filter).filter(Boolean)
    if (!filteredChildren?.length) return null
    // 子节点匹配
    return {
      ...node,
      children: filteredChildren as PersonnelTreeNode[],
    }
  }
  return tree
    .map((node) => _filter(node))
    .filter(Boolean) as PersonnelTreeNode[]
}

/** 统一生成树节点的key */
export const getTreeNodeKey = (
  type: 'department' | 'employee',
  info: { departmentId: any; employeeId?: any },
) => {
  const { departmentId, employeeId } = info
  return match(type)
    .with('department', () => `${departmentId}`)
    .with('employee', () => `employee-${departmentId}-${employeeId}`)
    .exhaustive()
}

/** 将人事数据转化为树节点的map */
export const useTreeMap = (
  config: Pick<
    PersonnelTree,
    | 'showDepartmentOperators'
    | 'showEmployees'
    | 'onlyDepartmentsWithEmployees'
    | 'disabledIds'
  > & {
    search?: string
    setExpandedKeys: Dispatch<SetStateAction<string[]>>
  },
) => {
  const { version } = useDeployConfig()
  const { message } = App.useApp()
  const {
    showDepartmentOperators,
    showEmployees,
    onlyDepartmentsWithEmployees,
    disabledIds,
    search,
    setExpandedKeys,
  } = config
  const { data, dispatchDepartmentAction } = usePersonnelData()
  const [editingNode, setEditingNode] = useState<
    | {
        key: string
        defaultNewName?: string
        parentKey?: string
        status: 'editing' | 'loading'
      }
    | undefined
  >()
  const treeMap = useMemo(() => {
    const { department, employee } = data
    if (!(department || (showEmployees && department && employee))) {
      return null
    }

    const treeMap = new Map<PersonnelTreeNode['key'], PersonnelTreeNode>()

    // 辅助函数：递归检查部门及其所有子部门下是否有员工
    const hasEmployeesInDepartment = (departmentId: number): boolean => {
      if (!showEmployees || !employee) return false

      // 检查当前部门下是否有员工（兼容 departmentIds 为空）
      let employeesInDept = employee.filter((emp) => {
        const deptIds = Array.isArray((emp as any).departmentIds)
          ? (emp as any).departmentIds
          : []
        return deptIds.includes(Number(departmentId))
      })

      // 如果有搜索条件，只考虑匹配搜索的员工
      if (search) {
        employeesInDept = employeesInDept.filter((emp) =>
          emp.name.includes(search),
        )
      }

      if (employeesInDept.length > 0) return true

      // 检查子部门下是否有员工
      const childDepartments = department.filter(
        (dept) => Number(dept.parentId) === departmentId,
      )

      for (const childDept of childDepartments) {
        if (hasEmployeesInDepartment(Number(childDept.id))) {
          return true
        }
      }

      return false
    }

    // 添加部门节点
    department.forEach((item) => {
      const key = getTreeNodeKey('department', { departmentId: item.id })
      const parentKey = getTreeNodeKey('department', {
        departmentId: item.parentId,
      })

      // 只有设置了 onlyDepartmentsWithEmployees 时才检查员工
      const hasEmployees = onlyDepartmentsWithEmployees
        ? hasEmployeesInDepartment(Number(item.id))
        : true

      const isDisabled = disabledIds?.some(
        (d) => d.id === item.id && d.type === 'department',
      )

      treeMap.set(key, {
        key,
        parentKey: item.parentId ? parentKey : undefined,
        type: 'department',
        ...item,
        // 所有部门都可选，但没有员工的部门会被禁用
        checkable: !isDisabled,
        selectable: !isDisabled,
        disableCheckbox: onlyDepartmentsWithEmployees
          ? !hasEmployees || isDisabled
          : isDisabled,
      })
    })

    // 建立父子关系
    ;[...treeMap.values()].forEach((n) => {
      const parentNode = treeMap.get(n.parentKey!)
      if (!parentNode) return
      if (!parentNode.children) {
        parentNode.children = []
      }
      parentNode.children.push(n)
    })

    // 排序
    ;[...treeMap.values()].forEach((n) => {
      if (!n.children) return
      n.children = n.children.sort((v1: any, v2: any) => {
        return v1.sort - v2.sort
      })
    })
    // 正在编辑的节点
    if (editingNode) {
      const target = treeMap.get(editingNode.key)
      if (target) {
        target.checkable = false
        target.selectable = false
      } else {
        const editTreeNode: PersonnelTreeNode = {
          key: editingNode.key,
          parentKey: editingNode.parentKey,
          type: 'department',
          name: '',
          id: '',
          checkable: false,
          selectable: false,
        }
        treeMap.set(editingNode.key, editTreeNode)
        const parentNode = treeMap.get(editingNode.parentKey!)!
        if (!parentNode.children) {
          parentNode.children = []
        }
        parentNode.children.push(editTreeNode)
      }
    }

    if (employee && showEmployees) {
      employee.forEach((item) => {
        const { id } = item
        const deptIds = Array.isArray((item as any).departmentIds)
          ? (item as any).departmentIds
          : []
        deptIds.forEach((departmentId: string | number) => {
          const departmentKey = getTreeNodeKey('department', { departmentId })
          const currentKey = getTreeNodeKey('employee', {
            departmentId,
            employeeId: id,
          })
          const isDisabled = disabledIds?.some(
            (d) => d.id === id && d.type === 'employee',
          )
          const currentNode: PersonnelTreeNode = {
            key: currentKey,
            parentKey: departmentKey,
            type: 'employee',
            ...item,
            checkable: !isDisabled,
            selectable: !isDisabled,
            disableCheckbox: isDisabled,
          }
          treeMap.set(currentKey, currentNode)
          const parentNode = treeMap.get(departmentKey)
          if (!parentNode) return
          // 确保parentNode.children是一个数组
          if (!parentNode.children) {
            parentNode.children = []
          }
          parentNode.children.push(currentNode)
        })
      })
    }

    // 为树添加title 操作符等
    ;[...treeMap.values()].forEach((n) => {
      const parentNode = treeMap.get(n.parentKey!)

      // 根据节点类型设置不同的图标
      if (n.type === 'department') {
        n.icon = <DepartmentIcon className='mx-auto mt-1' />
      } else if (n.type === 'employee') {
        n.icon = <EmployeeIcon className='mx-auto mt-1' />
      } else {
        n.icon = <DepartmentIcon className='mx-auto mt-1' />
      }

      if (editingNode?.key === n.key) {
        n.title = (
          <EditingInput
            disabled={editingNode.status === 'loading'}
            defaultName={editingNode.defaultNewName}
            onComplete={async (newName) => {
              if (!newName) {
                setEditingNode(undefined)
                return
              }
              // 全部门名称不能重复
              if (
                Object.values(department).some(
                  (item) => item.id !== n.id && item.name === newName,
                )
              ) {
                message.warning('部门名称不能重复')
                return
              }
              setEditingNode((prev) => {
                return {
                  ...prev!,
                  status: 'loading',
                }
              })
              if (Object.values(department).some((item) => item.id === n.id)) {
                // id存在 说明是重命名
                await dispatchDepartmentAction({
                  type: 'rename',
                  id: n.id,
                  newName,
                  version,
                })
                message.success('编辑成功')
              } else {
                await dispatchDepartmentAction({
                  type: 'add',
                  name: newName,
                  parentId: parentNode!.id,
                  version,
                })
                message.success('添加成功')
              }

              setEditingNode(undefined)
            }}
          />
        )
        if (editingNode.status === 'loading') {
          n.icon = <LoadingOutlined />
        }
      } else {
        const textTitle = match(search)
          .when(
            (v): v is undefined => !v,
            () => n.name,
          )
          .otherwise((search) => {
            const seqs = n.name.split(search)
            return (
              <span>
                {seqs.map((s, i) => {
                  if (i === seqs.length - 1) return s
                  return (
                    <>
                      {s}
                      <span className='bg-[#0C99FF]/20'>{search}</span>
                    </>
                  )
                })}
              </span>
            )
          })
        if (n.type === 'department' && showDepartmentOperators) {
          const operators: {
            icon: ReactNode
            name: string
            onClick: () => void
          }[] = [
            {
              icon: <Add />,
              name: '添加子部门',
              onClick: () => {
                setExpandedKeys((keys) => keys.concat(n.key))
                setEditingNode({
                  key: 'editingNode',
                  parentKey: n.key,
                  status: 'editing',
                })
              },
            },
            {
              icon: <Rename />,
              name: '重命名',
              onClick: () => {
                setEditingNode({
                  key: n.key,
                  defaultNewName: n.name,
                  status: 'editing',
                })
              },
            },
          ]

          if (parentNode) {
            const departmentBrothers = parentNode.children!.filter(
              (item) => item.type === 'department',
            )
            const position = departmentBrothers.findIndex(
              (item) => item.key === n.key,
            )
            if (position > 0) {
              operators.push({
                icon: <MoveUp />,
                name: '上移',
                onClick: async () => {
                  await dispatchDepartmentAction({
                    type: 'moveup',
                    id: n.id,
                    version,
                  })
                  message.success('移动成功')
                },
              })
            }
            if (position < departmentBrothers.length - 1) {
              operators.push({
                icon: <MoveDown />,
                name: '下移',
                onClick: async () => {
                  await dispatchDepartmentAction({
                    type: 'movedown',
                    id: n.id,
                    version,
                  })
                  message.success('移动成功')
                },
              })
            }

            operators.push({
              icon: <Del />,
              name: '删除',
              onClick: async () => {
                if (n.children && n.children.length) {
                  message.warning('该部门下有子部门或成员，不能删除')
                  return
                }
                await dispatchDepartmentAction({
                  type: 'delete',
                  id: n.id,
                  version,
                })
                message.success('删除成功')
              },
            })
          }

          n.title = (
            <>
              {textTitle}
              <SelfClosePopover>
                {operators.map((v) => (
                  <OperatorItem key={v.name} {...v} />
                ))}
              </SelfClosePopover>
            </>
          )
        } else {
          n.title = textTitle
        }
      }
    })

    return treeMap
  }, [
    data,
    dispatchDepartmentAction,
    editingNode,
    message,
    search,
    setExpandedKeys,
    showDepartmentOperators,
    showEmployees,
    version,
  ])
  return { treeMap, editingKey: editingNode?.key }
}

const SelfClosePopover: FC<PropsWithChildren> = (props) => {
  const { children } = props
  const [open, setOpen] = useState<false | undefined>()
  return (
    <Popover
      open={open}
      arrow={false}
      placement='bottomRight'
      trigger={['hover', 'focus']}
      content={
        <div
          className='flex flex-col p-2'
          onClick={(e) => {
            e.stopPropagation()
            e.preventDefault()
            setOpen(false)
            setTimeout(() => {
              setOpen(undefined)
            })
          }}
        >
          {children}
        </div>
      }
    >
      <OperatorIcon
        className='my-1 ml-auto operator'
        onClick={(e) => {
          e.stopPropagation()
          e.preventDefault()
        }}
      />
    </Popover>
  )
}

const EditingInput: FC<{
  defaultName?: string
  onComplete: (val: string) => void
  disabled?: boolean
}> = (props) => {
  const { defaultName = '', onComplete, disabled } = props
  const [value, setValue] = useState(defaultName)
  const inputRef = useRef<InputRef>(null)
  useEffect(() => {
    inputRef.current?.focus()
  }, [])
  useClickAway(
    () => {
      onComplete(value)
    },
    () => inputRef.current?.nativeElement,
  )
  return (
    <Input
      disabled={disabled}
      size='small'
      ref={inputRef}
      value={value}
      onChange={(e) => setValue(e.target.value)}
      onPressEnter={() => {
        if (value) onComplete(value)
      }}
    />
  )
}

const OperatorItem: FC<{
  icon: ReactNode
  name: string
  onClick: () => void
}> = (props) => {
  const { icon, name, onClick } = props
  return (
    <div
      className={cn(
        ' rounded hover:bg-[#F7F7F7] cursor-pointer',
        'w-44 h-7 py-2 pl-1 flex items-center gap-2',
      )}
      onClick={onClick}
    >
      {icon}
      {name}
    </div>
  )
}
