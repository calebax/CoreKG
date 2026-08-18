import { useState, useEffect } from 'react'
import { Modal, Input, Button, Tag, message, Spin } from 'antd'
import { CloseOutlined } from '@ant-design/icons'
import search from './images/search.svg'
import personIcon from '@/assets/icons/docs/person.svg'
import deletePersonIcon from '@/assets/icons/docs/deletePerson.svg'
import { useTranslation } from 'react-i18next'
import useUserData from '@/hooks/data/useUserData'
import scrollStyles from '@/styles/scroll/styles.module.scss'
import { PersonnelTree, PersonnelTreeNode } from '@/components/PersonnelTree'
import { usePersonnelData, PersonnelType, Department, Employee } from 'Personnel'

// 定义用户列表项类型
interface UserListItem {
  value: number
  label: string
}

interface AddMembersModalProps {
  open: boolean
  onClose: () => void
  onConfirm: (selectedIds: number[]) => void
  initialSelectedIds?: number[]
  lockedIds?: number[]
  minSelected?: number
}

export default function AddMembersModal({
  open,
  onClose,
  onConfirm,
  initialSelectedIds = [],
  lockedIds = [],
  minSelected = 0,
}: AddMembersModalProps) {
  const { t } = useTranslation(['pages', 'messages'])
  const { userList } = useUserData()
  const { loadData } = usePersonnelData()
  const [searchValue, setSearchValue] = useState('')
  const [selectedIds, setSelectedIds] = useState<number[]>(initialSelectedIds)
  const [loading, setLoading] = useState(false)
  const [treeKey, setTreeKey] = useState(0) // 用于强制重新渲染 PersonnelTree

  // 将权限系统的 uin 转换为 PersonnelTree 的 employee_id（基于员工数据的 uin->employee_id 一对一映射）
  const convertUinToEmployeeId = (uinIds: number[]): number[] => {
    const { data } = usePersonnelData.getState()
    const employees: Employee[] = (data as any)?.employee || []
    const uinToEmpId = new Map<number, number>()
    employees.forEach((emp) => {
      uinToEmpId.set(Number((emp as any).uin), Number((emp as any).id))
    })
    return uinIds.map((uin) => uinToEmpId.get(Number(uin)) ?? Number(uin))
  }

  // 初始化选中状态
  useEffect(() => {
    if (open) {
      setSearchValue('')
      // 加载人事数据（包括部门和成员）
      loadData('employee').then(() => {
        // 将权限系统的 uin 转换为 PersonnelTree 的 employee_id
        const convertedIds = convertUinToEmployeeId(initialSelectedIds)
        setSelectedIds(convertedIds)
      })
    }
  }, [open, initialSelectedIds, loadData])

  // 获取所有选中的成员信息
  const getSelectedMembers = () => {
    const { data } = usePersonnelData.getState()
    return selectedIds.map((employeeId) => {
      // 从人事数据中查找员工信息
      const employee = data.employee?.find((emp) => Number(emp.id) === employeeId)
      if (employee) {
        return {
          id: employeeId,
          name: employee.name,
          avatar: '',
        }
      }

      // 如果人事数据中找不到，再从用户列表中查找
      const user = userList.find((userItem: UserListItem) => userItem.value === employeeId)
      return {
        id: employeeId,
        name: user ? user.label : `用户${employeeId}`,
        avatar: '',
      }
    })
  }

  // 递归获取部门下的所有员工ID
  const getAllEmployeesInDepartment = (departmentId: number, includeSubDepartments = true): number[] => {
    const { data } = usePersonnelData.getState()
    // 添加类型断言，确保TypeScript能正确识别属性类型
    const { department, employee } = data as {
      department: Department[],
      employee: Employee[]
    }

    if (!department || !employee) return []

    // 1. 找到当前部门及其所有子部门
    const departmentIds: number[] = [departmentId]
    if (includeSubDepartments) {
      const findSubDepartments = (parentId: number) => {
        department
          .filter((dept) => Number(dept.parentId) === parentId)
          .forEach((subDept) => {
            departmentIds.push(Number(subDept.id))
            findSubDepartments(Number(subDept.id)) // 递归查找子部门
          })
      }
      findSubDepartments(departmentId)
    }

    // 2. 找到属于这些部门的所有员工
    const employees = employee.filter((emp) => {
      const deptIds = Array.isArray((emp as any).departmentIds) ? (emp as any).departmentIds : []
      return deptIds.some((deptId: string | number) =>
        departmentIds.includes(Number(deptId))
      )
    })

    return employees.map((emp) => Number(emp.id))
  }

  // 检查部门是否应该被选中（该部门下的所有员工都已被选中）
  const isDepartmentFullySelected = (departmentId: number, selectedEmployeeIds: number[]): boolean => {
    const employeesInDept = getAllEmployeesInDepartment(departmentId)
    return employeesInDept.length > 0 &&
      employeesInDept.every((empId) => selectedEmployeeIds.includes(empId))
  }

  // 获取所有应该被选中的部门ID
  const getSelectedDepartmentIds = (selectedEmployeeIds: number[]): number[] => {
    const { data } = usePersonnelData.getState()
    // 添加类型断言
    const { department } = data as {
      department: Department[]
    }

    if (!department) return []

    return department
      .filter((dept) => isDepartmentFullySelected(Number(dept.id), selectedEmployeeIds))
      .map((dept) => Number(dept.id))
  }

  // 处理树选择回调
  const handleTreeCheck = (
    checkedIds: { id: string | number; type: PersonnelType }[] | undefined,
    node: PersonnelTreeNode,
    type: 'check' | 'uncheck'
  ) => {
    // 当搜索时，checkedIds 只包含当前可见的选中节点
    // 我们需要根据操作类型来正确更新选中状态
    if (type === 'check') {
      // 选中操作：只添加当前点击的节点对应的员工
      let newEmployeeIds: number[] = []

      if (node.type === 'employee') {
        // 选中员工节点：只添加这个员工
        newEmployeeIds = [Number(node.id)]
      } else if (node.type === 'department') {
        // 选中部门节点：根据是否在搜索状态决定添加范围
        if (searchValue) {
          // 搜索状态下：只添加搜索结果中该部门下的员工
          const { data } = usePersonnelData.getState()
          const employeesInSearchResult = data.employee?.filter((emp) => {
            // 检查员工是否属于该部门且姓名匹配搜索
            return emp.departmentIds.includes(Number(node.id)) &&
              emp.name.includes(searchValue)
          }) || []
          newEmployeeIds = employeesInSearchResult.map(emp => Number(emp.id))
        } else {
          // 非搜索状态：添加该部门下的所有员工
          newEmployeeIds = getAllEmployeesInDepartment(Number(node.id))
        }
      }

      // 只添加尚未选中的员工（避免重复添加）
      const uniqueNewIds = newEmployeeIds.filter(id => !selectedIds.includes(id))
      if (uniqueNewIds.length > 0) {
        setSelectedIds([...selectedIds, ...uniqueNewIds])
      }
    } else {
      // 取消选中操作：需要区分是否在搜索状态
      let employeeIdsToRemove: number[] = []

      if (searchValue) {
        // 搜索状态下：只移除当前搜索结果中可见的员工
        if (node.type === 'employee') {
          // 取消选中员工节点
          employeeIdsToRemove = [Number(node.id)]
        } else if (node.type === 'department') {
          // 在搜索状态下，只移除当前搜索结果中该部门下的员工
          // 获取搜索结果中该部门下的所有员工节点
          const { data } = usePersonnelData.getState()
          const employeesInSearchResult = data.employee?.filter((emp) => {
            // 检查员工是否属于该部门且姓名匹配搜索
            return emp.departmentIds.includes(Number(node.id)) &&
              emp.name.includes(searchValue)
          }) || []

          employeeIdsToRemove = employeesInSearchResult.map(emp => Number(emp.id))
        }
      } else {
        // 非搜索状态：移除该部门下的所有员工
        if (node.type === 'employee') {
          employeeIdsToRemove = [Number(node.id)]
        } else if (node.type === 'department') {
          employeeIdsToRemove = getAllEmployeesInDepartment(Number(node.id))
        }
      }

      const candidates = selectedIds.filter(id => !employeeIdsToRemove.includes(id) || lockedIds.includes(id))
      const minAllowed = Math.max(minSelected, lockedIds.length)
      if (candidates.length < minAllowed) {
        return
      }
      setSelectedIds(candidates)
    }
  }

  // 将员工ID映射到树节点key格式
  const getTreeCheckedKeys = (): { id: string | number; type: PersonnelType }[] => {
    // 1. 添加所有选中的员工节点
    const employeeItems = selectedIds.map((id) => ({
      id,
      type: 'employee' as PersonnelType,
    }))

    // 2. 添加所有应该被选中的部门节点
    const selectedDepartmentIds = getSelectedDepartmentIds(selectedIds)
    const departmentItems = selectedDepartmentIds.map((id) => ({
      id,
      type: 'department' as PersonnelType,
    }))

    const result = [...employeeItems, ...departmentItems]

    return result
  }

  // 清空已选
  const handleClearSelected = () => {
    const minAllowed = Math.max(minSelected, lockedIds.length)
    if (minAllowed === 0) {
      setSelectedIds([])
    } else {
      setSelectedIds(Array.from(new Set(lockedIds)))
    }
    // 强制重新渲染 PersonnelTree 以同步状态
    setTreeKey(prev => prev + 1)
  }

  // 移除单个已选成员
  const handleRemoveSelected = (memberId: number) => {
    if (lockedIds.includes(memberId)) return
    const minAllowed = Math.max(minSelected, lockedIds.length)
    const newSelectedIds = selectedIds.filter(id => id !== memberId)
    if (newSelectedIds.length < minAllowed) return
    setSelectedIds(newSelectedIds)

    // 强制重新渲染 PersonnelTree 以同步状态
    setTreeKey(prev => prev + 1)
  }

  // 将 PersonnelTree 的 employee_id 严格转换为权限系统的 uin（无法匹配则判为失败）
  const convertEmployeeIdToUin = (employeeIds: number[]): { converted: number[]; failed: number[] } => {
    const { data } = usePersonnelData.getState()
    const employees: Employee[] = (data as any)?.employee || []

    const converted: number[] = []
    const failed: number[] = []

    employeeIds.forEach((employeeId) => {
      const emp = employees.find((e) => Number((e as any).id) === Number(employeeId))
      if (emp && (emp as any).uin) {
        converted.push(Number((emp as any).uin))
      } else {
        failed.push(employeeId)
      }
    })

    return { converted, failed }
  }

  // 确认添加
  const handleConfirm = () => {
    if (selectedIds.length === 0) {
      message.warning(t('app.docs.detail.selectAtLeastOneMember', { defaultValue: '请至少选择一个成员' }))
      return
    }

    // 将 PersonnelTree 的 employee_id 转换回权限系统的 uin（严格校验）
    const { converted, failed } = convertEmployeeIdToUin(selectedIds)
    if (failed.length > 0) {
      message.warning(t('messages:invalidMembersFound', { defaultValue: '存在无法匹配到当前公司账号的成员，请检查选择' }))
      return
    }
    setLoading(true)
    setTimeout(() => {
      onConfirm(converted)
      onClose()
      setLoading(false)
    }, 500)
  }

  const selectedMembers = getSelectedMembers()
  const minAllowed = Math.max(minSelected, lockedIds.length)

  return (
    <Modal
      open={open}
      onCancel={onClose}
      footer={null}
      width={600}
      centered
      destroyOnClose
      closable={false}
      styles={{
        body: { padding: 0 },
        content: { borderRadius: '10px', boxShadow: '0px 2px 30px 0px rgba(0,0,0,0.1)' }
      }}
    >
      <div className="bg-white rounded-lg flex flex-col gap-2.5">
        {/* 头部 */}
        <div className="flex py-3 items-center justify-between border-b border-[#EFF1F4]">
          <span className="text-sm font-medium text-[#000000]">
            {t('app.docs.detail.selectMembers', { defaultValue: '选择成员' })}
          </span>
          <button
            onClick={onClose}
            className="w-[23px] h-6 rounded flex items-center justify-center cursor-pointer hover:bg-gray-100 transition-colors"
          >
            <CloseOutlined className="text-gray-500" />
          </button>
        </div>

        {/* 内容区域 */}
        <div className="flex flex-1 gap-4">
          {/* 左侧：搜索和树结构 */}
          <div className="flex flex-col gap-5 w-[263px] bg-white">
            <div>
              <Input
                placeholder={t('app.docs.detail.searchDepartmentOrMember', { defaultValue: '搜索部门或成员' })}
                prefix={<img src={search} alt="search" className="w-4 h-4 flex-shrink-0" />}
                value={searchValue}
                onChange={(e) => setSearchValue(e.target.value)}
                className="w-full h-[30px] bg-[#F5F5F5] border-none px-2"
                style={{
                  borderRadius: '6px',
                  backgroundColor: '#F5F5F5'
                }}
              />
            </div>

            <div className={`overflow-y-auto rounded-lg pr-2 ${scrollStyles.scroll}`} style={{ height: '430px' }}>
              <PersonnelTree
                key={treeKey}
                checkable={true}
                showDepartmentOperators={false}
                showEmployees={true}
                hideSearch={true}
                externalSearch={searchValue}
                checkedIds={getTreeCheckedKeys()}
                onCheck={handleTreeCheck}
                onlyDepartmentsWithEmployees={true}
                className="h-full"
                style={{ height: '100%' }}
              />
            </div>
          </div>

          {/* 右侧：已选择成员 */}
          <div className="w-[287px] p-[10px] rounded-md bg-[#FAFAFA] relative h-[480px] flex flex-col">
            <div className="flex items-center justify-between h-[30px] mb-5 flex-shrink-0">
              <div className="flex items-center gap-1">
                <span className="text-sm font-medium text-[#0C1F17] leading-[22px]">
                  {t('app.docs.detail.selected', { defaultValue: '已选择' })}
                </span>
                <span className="text-sm font-medium text-[#919497] leading-[22px]">
                  {selectedMembers.length}{t('app.docs.detail.items', { defaultValue: '项' })}
                </span>
              </div>
              <button
                onClick={handleClearSelected}
                disabled={selectedMembers.length <= minAllowed}
                className={`text-sm font-medium  transition-colors ${selectedMembers.length > 0
                  ? 'text-[#919497] cursor-pointer hover:text-[#0C1F17]'
                  : 'text-[#919497] opacity-50 cursor-not-allowed'
                  }`}
              >
                {t('app.docs.detail.clearSelected', { defaultValue: '清空已选' })}
              </button>
            </div>

            <div className={`overflow-y-auto rounded-lg ${scrollStyles.scroll} flex-1`}>
              {selectedMembers.length > 0 ? (
                <div className="space-y-1">
                  {selectedMembers.map(member => {
                    const isLocked = lockedIds.includes(member.id)
                    const disableDelete = isLocked || selectedMembers.length <= minAllowed
                    return (
                    <div
                      key={member.id}
                      className="flex items-center justify-between h-[30px] pr-2 hover:bg-gray-50 rounded"
                    >
                      <div className="flex items-center gap-2 flex-1">
                        <img
                          src={personIcon}
                          alt="person"
                          className="w-4 h-4 flex-shrink-0"
                        />
                        <span className="text-sm font-medium text-[#0C1F17]">{member.name}</span>
                      </div>
                      <button
                        onClick={() => handleRemoveSelected(member.id)}
                        disabled={disableDelete}
                        className={`text-gray-400 transition-colors ${disableDelete ? 'opacity-40 cursor-not-allowed' : 'cursor-pointer hover:text-gray-600'}`}
                      >
                        <img
                          src={deletePersonIcon}
                          alt="delete"
                          className="w-4 h-4"
                        />
                      </button>
                    </div>
                  )})}
                </div>
              ) : (
                <div className="text-center text-gray-500 py-8">
                  {t('app.docs.detail.noSelectedMembers', { defaultValue: '暂无选择成员' })}
                </div>
              )}
            </div>

          </div>
        </div>

        {/* 底部分割线 */}
        <div className="h-px bg-[#EFF1F4]" />

        {/* 底部按钮 */}
        <div className="flex justify-end gap-2">
          <button
            className="px-6 py-2 bg-[#F5F5F5] text-[#0C1F17] rounded-md text-sm font-medium cursor-pointer hover:bg-[#F5F5F5] transition-colors"
            onClick={onClose}
          >
            {t('app.docs.detail.cancel', { defaultValue: '取消' })}
          </button>
          <button
            className={`px-6 py-2 rounded-md text-sm font-medium flex items-center gap-2 cursor-pointer ${!loading && selectedIds.length > 0
              ? 'bg-[#0C99FF] text-[#ffffff] hover:bg-[#0C99FF]'
              : 'bg-[#0C99FF] text-[#ffffff] opacity-50 cursor-not-allowed'
              }`}
            onClick={handleConfirm}
            disabled={selectedIds.length === 0}
          >
            {loading ? <Spin size='small' /> : null}
            {t('app.docs.detail.confirm', { defaultValue: '确定' })}
          </button>
        </div>
      </div>
    </Modal>
  )
}