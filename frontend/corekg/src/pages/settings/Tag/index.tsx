import { FC, useState } from 'react'
import {
  Table,
  Button,
  Input,
  Space,
  Modal,
  Form,
  Select,
  message,
  Pagination,
} from 'antd'
import {
  PlusOutlined,
  ExclamationCircleOutlined,
  SearchOutlined,
} from '@ant-design/icons'
import { useRequest } from 'ahooks'
import dayjs from 'dayjs'
import {
  listResourceTag,
  createResourceTag,
  modifyResourceTag,
  deleteResourceTag,
  listTagGroup,
} from '@/api/knowledge'
import DeleteConfirmModal from '@/pages/app/docs/detail/components/DeleteConfirmModal'

const Tag: FC = () => {
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [searchText, setSearchText] = useState('')
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<any>(null)
  const [form] = Form.useForm()

  const {
    data,
    loading,
    run: refresh,
  } = useRequest(
    () =>
      listResourceTag({
        limit: pageSize,
        offset: (page - 1) * pageSize,
        name: searchText || undefined,
      }),
    {
      refreshDeps: [page, pageSize, searchText],
    },
  )

  const { data: tagGroupsData } = useRequest(() =>
    listTagGroup({ limit: 1000, offset: 0 }),
  )

  const handleAdd = () => {
    setEditingItem(null)
    form.resetFields()
    setIsModalOpen(true)
  }

  const handleEdit = (record: any) => {
    setEditingItem(record)
    setIsModalOpen(true)
  }

  // 监听弹窗打开状态，重置并填充表单
  useEffect(() => {
    if (isModalOpen) {
      if (editingItem) {
        form.setFieldsValue({
          name: editingItem.name,
          tag_group_id: editingItem.tag_group_id,
        })
      } else {
        form.resetFields()
      }
    }
  }, [isModalOpen, editingItem, form])

  const handleDelete = (record: any) => {
    setEditingItem(record)
    setIsDeleteModalOpen(true)
  }

  const handleConfirmDelete = async () => {
    if (!editingItem) return
    try {
      await deleteResourceTag({ tag_id: editingItem.tag_id })
      message.success('删除成功')
      setIsDeleteModalOpen(false)
      refresh()
    } catch (error: any) {
      console.log(error)
    }
  }

  const handleModalOk = async () => {
    try {
      const values = await form.validateFields()
      if (editingItem) {
        await modifyResourceTag({
          tag_id: editingItem.tag_id,
          name: values.name,
          tag_group_id: values.tag_group_id,
        })
        message.success('修改成功')
      } else {
        await createResourceTag({
          name: values.name,
          tag_group_id: values.tag_group_id,
        })
        message.success('创建成功')
      }
      setIsModalOpen(false)
      refresh()
    } catch (error: any) {
      console.log(error)
    }
  }

  const columns = [
    {
      title: '序号',
      dataIndex: 'index',
      key: 'index',
      width: 80,
      render: (_: any, __: any, index: number) =>
        (page - 1) * pageSize + index + 1,
    },
    {
      title: '标签名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '标签分类',
      dataIndex: 'tag_group_name',
      key: 'tag_group_name',
    },
    {
      title: '创建时间',
      dataIndex: 'create_at',
      key: 'create_at',
      render: (text: number) =>
        text ? dayjs(text * 1000).format('YYYY-MM-DD HH:mm:ss') : '-',
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_: any, record: any) => (
        <Space size='middle'>
          <Button
            type='link'
            className='p-0 h-auto text-[#0C99FF] hover:text-[#38b2ff]'
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Button
            type='link'
            danger
            className='p-0 h-auto'
            onClick={() => handleDelete(record)}
          >
            删除
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <div className='flex flex-col h-full bg-[#FAFAFA] p-4 gap-4 overflow-hidden'>
      <div className='flex justify-between items-center'>
        <Space size='middle'>
          <Button
            type='primary'
            icon={<PlusOutlined />}
            onClick={handleAdd}
            className='bg-[#0C99FF]'
          >
            创建标签
          </Button>
          <Input
            placeholder='搜索标签名称'
            prefix={<SearchOutlined className='text-gray-400' />}
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            style={{ width: 250 }}
            allowClear
          />
        </Space>
      </div>

      <div className='flex-1 bg-white rounded-lg shadow-sm overflow-hidden flex flex-col'>
        <Table
          columns={columns}
          dataSource={data?.list || []}
          loading={loading}
          pagination={false}
          rowKey='tag_id'
          className='flex-1 overflow-auto'
        />
        <div className='p-4 border-t border-[#EFF1F4] flex justify-end'>
          <Pagination
            current={page}
            pageSize={pageSize}
            total={data?.total || 0}
            onChange={(p, ps) => {
              setPage(p)
              setPageSize(ps)
            }}
            showSizeChanger
            showQuickJumper
            showTotal={(total) => `共计 ${total} 条数据`}
          />
        </div>
      </div>

      <Modal
        title={editingItem ? '编辑标签' : '创建标签'}
        open={isModalOpen}
        onOk={handleModalOk}
        onCancel={() => setIsModalOpen(false)}
        okText='确认'
        cancelText='取消'
        destroyOnClose
      >
        <Form form={form} layout='vertical' className='mt-4'>
          <Form.Item
            name='name'
            label='标签名称'
            rules={[
              { required: true, message: '请输入标签名称' },
              { max: 20, message: '标签名称最多20个字符' },
            ]}
          >
            <Input placeholder='请输入标签名称' maxLength={20} />
          </Form.Item>
          <Form.Item
            name='tag_group_id'
            label='标签分类'
            rules={[{ required: true, message: '请选择标签分类' }]}
          >
            <Select
              placeholder='请选择标签分类'
              options={tagGroupsData?.list?.map((group: any) => ({
                label: group.name,
                value: group.tag_group_id,
              }))}
            />
          </Form.Item>
        </Form>
      </Modal>

      <DeleteConfirmModal
        visible={isDeleteModalOpen}
        customText={`确定要删除标签 "${editingItem?.name}" 吗？删除后不可恢复。`}
        onCancel={() => setIsDeleteModalOpen(false)}
        onConfirm={handleConfirmDelete}
      />
    </div>
  )
}

export default Tag
