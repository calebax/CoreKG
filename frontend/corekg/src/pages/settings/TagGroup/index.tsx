import { FC, useEffect, useState } from 'react'
import {
  Table,
  Button,
  Input,
  Space,
  Modal,
  Form,
  message,
  Pagination,
} from 'antd'
import { PlusOutlined, ExclamationCircleOutlined } from '@ant-design/icons'
import { useRequest } from 'ahooks'
import dayjs from 'dayjs'
import { cn } from '@/utils'
import {
  listTagGroup,
  createTagGroup,
  modifyTagGroup,
  deleteTagGroup,
} from '@/api/knowledge'
import DeleteConfirmModal from '@/pages/app/docs/detail/components/DeleteConfirmModal'

const TagGroup: FC = () => {
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
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
      listTagGroup({
        limit: pageSize,
        offset: (page - 1) * pageSize,
      }),
    {
      refreshDeps: [page, pageSize],
    },
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
      await deleteTagGroup({ tag_group_id: editingItem.tag_group_id })
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
        await modifyTagGroup({
          tag_group_id: editingItem.tag_group_id,
          name: values.name,
        })
        message.success('修改成功')
      } else {
        await createTagGroup({ name: values.name })
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
      title: '标签分类',
      dataIndex: 'name',
      key: 'name',
      width: 200,
    },
    {
      title: '创建时间',
      dataIndex: 'create_at',
      key: 'create_at',
      width: 200,
      render: (text: number) =>
        text ? dayjs(text * 1000).format('YYYY-MM-DD HH:mm:ss') : '-',
    },
    {
      title: '操作',
      key: 'action',
      width: 80,
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
        <Button
          type='primary'
          icon={<PlusOutlined />}
          onClick={handleAdd}
          className='bg-[#0C99FF]'
        >
          创建标签分类
        </Button>
      </div>

      <div className='flex-1 bg-white rounded-lg shadow-sm overflow-hidden flex flex-col'>
        <Table
          columns={columns}
          dataSource={data?.list || []}
          loading={loading}
          pagination={false}
          rowKey='tag_group_id'
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
        title={editingItem ? '编辑标签分类' : '创建标签分类'}
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
            label='标签分类名称'
            rules={[
              { required: true, message: '请输入标签分类名称' },
              { max: 20, message: '标签分类名称最多20个字符' },
            ]}
          >
            <Input placeholder='请输入标签分类名称' maxLength={20} />
          </Form.Item>
        </Form>
      </Modal>

      <DeleteConfirmModal
        visible={isDeleteModalOpen}
        customText={`确定要删除标签分类 "${editingItem?.name}" 吗？删除后不可恢复。`}
        onCancel={() => setIsDeleteModalOpen(false)}
        onConfirm={handleConfirmDelete}
      />
    </div>
  )
}

export default TagGroup
