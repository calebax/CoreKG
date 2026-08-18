import { FC, ReactNode } from 'react'
import {
  App,
  Button,
  Drawer,
  Form,
  Input,
  Modal,
  Popover,
  Select,
  Tooltip,
} from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { Department, Employee, usePersonnelData } from 'Personnel'
import { useRequest } from 'ahooks'
import { cn, useOpen } from '@/utils'
import { PersonnelTree } from '@/components/PersonnelTree'
import { useDeployConfig } from '@/utils/useDeployConfig'
import DelIcon from './images/del.svg?react'
import DepartmentIcon from './images/department.svg?react'
import OperIcon from './images/operator.svg?react'
import PrimaryIcon from './images/primary.svg?react'

type EmployeeInfo = Omit<Employee, 'id' | 'created_at' | 'uin' | 'employee_id'>

const FormItem = Form.Item<EmployeeInfo>

export const EmployeeDrawer: FC<{
  title: string
  open: boolean
  onClose: () => void
  onSubmit?: (val: EmployeeInfo) => Promise<any>
  defaultValue?: Partial<Employee>
}> = (props) => {
  const { message } = App.useApp()
  const { version } = useDeployConfig()
  const {
    data: { employee },
  } = usePersonnelData()
  const { title, open, onClose, onSubmit, defaultValue } = props

  const { id: currentEmployeeId } = defaultValue ?? {}
  const [form] = Form.useForm<EmployeeInfo>()
  const [showTree, { toggle }, treeKey] = useOpen()
  const { run: submit, loading: submitting } = useRequest(
    async () => {
      const formValue = await form.validateFields()
      await onSubmit?.(formValue)
      message.success('操作成功')
      onClose()
    },
    {
      manual: true,
    },
  )
  return (
    <Drawer
      open={open}
      onClose={onClose}
      maskClosable={!submitting}
      keyboard={!submitting}
      closable={!submitting}
      title={title}
    >
      <div className='flex flex-col min-w-75 min-h-full'>
        <Form
          className='flex-1'
          layout='vertical'
          form={form}
          initialValues={defaultValue}
        >
          <FormItem
            name='name'
            label='组织昵称'
            rules={[{ required: true, message: '组织昵称' }]}
          >
            <Input maxLength={50} showCount />
          </FormItem>
          <FormItem
            name='phone'
            label='手机号码'
            rules={[
              { required: version === 'saas', message: '请填写手机号码' },
              {
                pattern: /^1[3-9]\d{9}$/,
                message: '请填写正确的手机号码',
              },
              {
                validator: async (_, phone) => {
                  // 如果手机号为空，则不进行重复校验
                  if (!phone || phone.trim() === '') return

                  const target = employee?.find(
                    (item) =>
                      item.id !== currentEmployeeId && phone === item.phone,
                  )
                  if (target) throw new Error('手机号不能重复')
                },
              },
            ]}
          >
            <Input />
          </FormItem>
          <FormItem
            name='email'
            className='mb-4'
            label='邮箱'
            rules={[
              { required: version !== 'saas', message: '请填写邮箱' },
              {
                type: 'email',
                message: '请填写正确的邮箱',
              },
              {
                validator: async (_, email) => {
                  // 如果邮箱为空，则不进行重复校验
                  if (!email || email.trim() === '') return

                  const target = employee?.find(
                    (item) =>
                      item.id !== currentEmployeeId && email === item.email,
                  )
                  if (target) throw new Error('邮箱不能重复')
                },
              },
            ]}
          >
            <Input />
          </FormItem>
          <div
            className={cn(
              'bg-[#9194971A] text-[#ABAFB2] rounded',
              'py-0.5 px-9 whitespace-nowrap',
              {
                hidden: defaultValue?.email || version !== 'custom',
              },
            )}
          >
            首次登录时账号的默认密码为「pwd+邮箱」
            <br />
            若注册邮箱为：a@b.com
            <br />
            则默认密码为：pwda@b.com
          </div>
          <Form.List
            name='departmentIds'
            rules={[
              {
                validator: async (_, v) => {
                  if (!v || v.length === 0) throw new Error('必须选择一个部门')
                },
              },
            ]}
          >
            {(fields, { remove, move }, { errors }) => {
              return (
                <>
                  <div className='flex flex-col gap-2'>
                    <Form.Item
                      label='所属部门'
                      layout='horizontal'
                      required
                      className='m-0'
                    >
                      <Button
                        size='small'
                        icon={<PlusOutlined />}
                        type='text'
                        className='text-[#0C99FF]'
                        onClick={toggle}
                      >
                        添加
                      </Button>
                    </Form.Item>
                    {fields.map((item, i) => (
                      <Form.Item {...item} className='m-0'>
                        <DepartmentInput
                          isPrimary={i === 0}
                          setPrimary={() => move(i, 0)}
                          onDel={() => remove(i)}
                        />
                      </Form.Item>
                    ))}

                    {errors.map((s) => (
                      <span className='text-[#ff4d4f]'>{s}</span>
                    ))}
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
          <FormItem
            name='role'
            label='角色'
            rules={[{ required: true, message: '请选择角色' }]}
          >
            <Select
              options={[
                { label: '管理员', value: 'sys_admin' },
                { label: '成员', value: 'sys_employee' },
              ]}
            />
          </FormItem>
        </Form>
        <span className='flex items-center gap-2'>
          <Button onClick={onClose} disabled={submitting}>
            取消
          </Button>
          <Button
            onClick={submit}
            loading={submitting}
            className='bg-[#0C99FF] text-[#ffffff]'
          >
            确定
          </Button>
        </span>
      </div>
    </Drawer>
  )
}

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
  ) => {
    return (
      <div
        className={cn(
          'rounded hover:bg-[#F7F7F7] cursor-pointer',
          ' pl-1 px-1 flex items-center gap-2',
        )}
        onClick={onClick}
      >
        {icon}
        {label}
      </div>
    )
  }
  return (
    <Input
      value={department?.find((item) => item.id === value)?.name}
      className='pointer-events-none'
      prefix={
        isPrimary ? (
          <Tooltip
            placement='topLeft'
            title={'设置为主部门后，该部门将显示在成员部门的首位'}
          >
            <div
              className={cn(
                'bg-[#0C99FF] text-xs pointer-events-auto select-none',
                'w-5 h-5 flex items-center justify-center text-[#ffffff]',
              )}
            >
              主
            </div>
          </Tooltip>
        ) : null
      }
      suffix={
        <Popover
          arrow={false}
          placement='bottomRight'
          content={
            <div className='flex flex-col p-2.5'>
              {isPrimary
                ? null
                : getOperatorItem(<PrimaryIcon />, '设为主部门', setPrimary)}
              {getOperatorItem(<DelIcon />, '删除', onDel)}
            </div>
          }
        >
          <div
            className={cn(
              ' hover:bg-[#9194971A] cursor-pointer pointer-events-auto',
              'w-6 h-6 flex items-center justify-center',
            )}
          >
            <OperIcon />
          </div>
        </Popover>
      }
    />
  )
}

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
              const currentDepartment = department!.find(
                (item) => item.id === id,
              )!
              return (
                <div className='flex items-center px-0.5 py-1'>
                  <DepartmentIcon className='mr-2' />
                  {currentDepartment?.name}
                  <DelIcon
                    className='ml-auto cursor-pointer'
                    onClick={() => {
                      setDepartment((v) => {
                        if (!v) return v
                        return v.filter((_, i) => i !== index)
                      })
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
