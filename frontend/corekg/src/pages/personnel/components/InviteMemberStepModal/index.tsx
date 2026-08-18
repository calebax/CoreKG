import { FC, useState, ReactNode, useEffect } from 'react'
import { Button, Modal, Select, Input, Popover, Tooltip, Form } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { usePersonnelData, Department } from 'Personnel'
import { useTranslation } from 'react-i18next'
import { cn, useOpen } from '@/utils'
import { PersonnelTree } from '@/components/PersonnelTree'
import DelIcon from '../EmployeeDrawer/images/del.svg?react'
import DepartmentIcon from '../EmployeeDrawer/images/department.svg?react'
import OperIcon from '../EmployeeDrawer/images/operator.svg?react'
import PrimaryIcon from '../EmployeeDrawer/images/primary.svg?react'
import styles from './InviteMemberStepModal.module.scss'

interface InviteMemberStepModalProps {
  open: boolean
  onCancel: () => void
  onNext: (data: {
    departmentIds: Department['id'][]
    role: 'sys_admin' | 'sys_employee'
  }) => void
  initialValues?: {
    departmentIds: Department['id'][]
    role: 'sys_admin' | 'sys_employee'
  }
}

export const InviteMemberStepModal: FC<InviteMemberStepModalProps> = ({
  open,
  onCancel,
  onNext,
  initialValues,
}) => {
  const { t } = useTranslation('pages')
  const {
    data: { department },
  } = usePersonnelData()
  const [form] = Form.useForm<{
    departmentIds: Department['id'][]
    role: 'sys_admin' | 'sys_employee'
  }>()
  const [showTree, { toggle }, treeKey] = useOpen()
  // 监听角色字段，控制下一步按钮的禁用状态
  const role = Form.useWatch('role', form)

  // 初始化默认部门（当前选中的部门或顶级部门）
  useEffect(() => {
    if (open && department) {
      if (initialValues) {
        // 如果有初始值，使用初始值
        form.setFieldsValue({
          departmentIds: initialValues.departmentIds,
          role: initialValues.role,
        })
      } else {
        // 否则使用默认值（顶级部门）
        const topDepartment = department.find((item) => !item.parentId)
        if (topDepartment) {
          form.setFieldsValue({
            departmentIds: [topDepartment.id],
            role: undefined,
          })
        }
      }
    }
  }, [open, department, form, initialValues])

  // 确定按钮是否可用
  const handleNext = async () => {
    const values = await form.validateFields()
    onNext({
      departmentIds: values.departmentIds,
      role: values.role,
    })
  }

  // 处理取消
  const handleCancel = () => {
    form.resetFields()
    onCancel()
  }

  return (
    <Modal
      title={t('settings.inviteMember') || '邀请成员'}
      open={open}
      onCancel={handleCancel}
      onOk={handleNext}
      okText='下一步'
      cancelText='取消'
      okButtonProps={{
        className: 'bg-[#0C99FF] text-white',
        disabled: !role,
      }}
      cancelButtonProps={{
        className: 'bg-[#F1F5F9] text-[#0C1F17]',
      }}
      closable
      keyboard
      maskClosable
      className={styles.inviteMemberStepModal}
      width={520}
    >
      <Form form={form} layout='vertical' className={styles.form}>
        {/* 所属部门 */}
        <Form.List name='departmentIds'>
          {(fields, { remove, move }, { errors }) => {
            return (
              <>
                <div className={styles.departmentSection}>
                  <div className={styles.departmentHeader}>
                    <div className={styles.labelWrapper}>
                      <span className={styles.label}>所属部门</span>
                      <span className={styles.required}>*</span>
                    </div>
                    <Button
                      type='text'
                      icon={<PlusOutlined />}
                      className={styles.addButton}
                      onClick={toggle}
                    >
                      添加
                    </Button>
                  </div>
                  <div className={styles.departmentList}>
                    {fields.map((item, i) => (
                      <Form.Item
                        {...item}
                        key={item.key}
                        className={styles.departmentItem}
                      >
                        <DepartmentInput
                          isPrimary={i === 0}
                          setPrimary={() => move(i, 0)}
                          onDel={i !== 0 ? () => remove(i) : undefined}
                        />
                      </Form.Item>
                    ))}
                    {errors.map((s, index) => (
                      <span key={index} className={styles.error}>
                        {s}
                      </span>
                    ))}
                  </div>
                </div>
                <DepartmentSelectModal
                  key={treeKey}
                  open={showTree}
                  onClose={toggle}
                  onOk={(val) => {
                    form.setFieldValue('departmentIds', val)
                    toggle()
                  }}
                  defaultValue={form.getFieldValue('departmentIds')}
                />
              </>
            )
          }}
        </Form.List>

        {/* 所属角色 */}
        <div className={styles.roleSection}>
          <div className={styles.roleWrapper}>
            <div className={styles.roleHeader}>
              <span className={styles.label}>所属角色</span>
              <span className={styles.required}>*</span>
            </div>
            <Form.Item
              name='role'
              rules={[{ required: true, message: '请选择角色' }]}
              className='m-0'
            >
              <Select
                placeholder='请选择'
                className={styles.roleSelect}
                getPopupContainer={(triggerNode) =>
                  triggerNode?.closest('.ant-modal-content') || document.body
                }
                dropdownStyle={{ zIndex: 1050 }}
                placement='bottomLeft'
                options={[
                  { label: '管理员', value: 'sys_admin' },
                  { label: '成员', value: 'sys_employee' },
                ]}
              />
            </Form.Item>
          </div>
        </div>
      </Form>
    </Modal>
  )
}

// 部门输入组件
const DepartmentInput: FC<
  ValueController<Department['id']> & {
    isPrimary?: boolean
    setPrimary?: () => void
    onDel?: () => void
  }
> = (props) => {
  const { value, isPrimary, setPrimary, onDel } = props
  const {
    data: { department },
  } = usePersonnelData()
  if (!value) return null

  const getOperatorItem = (
    icon: ReactNode,
    label: ReactNode,
    onClick?: () => void,
  ) => (
    <div
      className={cn(
        'rounded hover:bg-[#F7F7F7] cursor-pointer',
        'pl-1 px-1 flex items-center gap-2',
      )}
      onClick={onClick}
    >
      {icon}
      {label}
    </div>
  )

  return (
    <div className={styles.departmentInputWrapper}>
      <Input
        value={department?.find((item) => item.id === value)?.name}
        className='pointer-events-none'
        prefix={
          isPrimary ? (
            <Tooltip
              placement='topLeft'
              title={'设置为主部门后，该部门将显示在成员部门的首位'}
            >
              <div className={styles.primaryTag}>主</div>
            </Tooltip>
          ) : null
        }
        suffix={
          !isPrimary ? (
            <Popover
              arrow={false}
              placement='bottomRight'
              content={
                <div className='flex flex-col p-2.5'>
                  {getOperatorItem(<PrimaryIcon />, '设为主部门', setPrimary)}
                  {onDel && getOperatorItem(<DelIcon />, '删除', onDel)}
                </div>
              }
            >
              <div className={styles.operatorButton}>
                <OperIcon />
              </div>
            </Popover>
          ) : null
        }
      />
    </div>
  )
}

// 部门选择弹窗组件
const DepartmentSelectModal: FC<{
  open: boolean
  onClose: () => void
  onOk: (val: Department['id'][]) => void
  defaultValue?: Department['id'][]
}> = (props) => {
  const { open, onClose, onOk, defaultValue } = props
  const {
    data: { department },
  } = usePersonnelData()
  const [checkedDepartment, setDepartment] = useState<
    PersonnelTree['checkedIds']
  >(() =>
    defaultValue?.map(
      (id) =>
        ({
          id,
          type: 'department',
        }) as const,
    ),
  )
  return (
    <Modal
      title='添加部门'
      open={open}
      onCancel={onClose}
      onOk={() => onOk((checkedDepartment ?? []).map((item) => item.id))}
      okButtonProps={{ className: 'bg-[#0C99FF] text-[#ffffff]' }}
      width={700}
      className='top-10'
    >
      <div className='flex h-[450px] gap-4 overflow-hidden'>
        <PersonnelTree
          checkable
          checkedIds={checkedDepartment}
          onCheck={setDepartment}
          checkStrictly
          placeholder='搜索部门'
          className='flex-1'
          selectable={false}
        />
        <div
          className={cn(
            'flex-1 p-2.5  flex flex-col',
            'bg-[#FAFAFA] overflow-hidden rounded',
          )}
        >
          <div className='py-3 flex justify-between font-medium'>
            <span>已选择{checkedDepartment?.length ?? 0}项</span>
            <span className='cursor-pointer' onClick={() => setDepartment([])}>
              清空已选
            </span>
          </div>
          <div className='flex-1 overflow-auto flex flex-col'>
            {checkedDepartment?.map((item, index) => {
              const { id } = item
              const currentDepartment = department?.find((d) => d.id === id)
              if (!currentDepartment) return null
              return (
                <div key={id} className='flex items-center px-0.5 py-1'>
                  <DepartmentIcon className='mr-2' />
                  {currentDepartment.name}
                  <DelIcon
                    className='ml-auto cursor-pointer'
                    onClick={() => {
                      setDepartment((v) => v?.filter((_, i) => i !== index))
                    }}
                  />
                </div>
              )
            })}
          </div>
        </div>
      </div>
    </Modal>
  )
}
