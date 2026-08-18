import { FC, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import {
  Table,
  Button,
  Input,
  Space,
  message,
  Pagination,
  Breadcrumb,
  Tooltip,
} from 'antd'
import {
  PlusOutlined,
  SearchOutlined,
  EditOutlined,
  DeleteOutlined,
} from '@ant-design/icons'
import { useRequest } from 'ahooks'
import dayjs from 'dayjs'
import {
  listSynonymKeywords,
  deleteSynonymKeyword,
  getSynonymKeyword,
} from '@/api/knowledge'
import SeparatorIcon from '@/assets/separator.svg?react'
import DeleteConfirmModal from '@/pages/app/docs/detail/components/DeleteConfirmModal'
import SynonymModal from './components/SynonymModal'
import styles from './styles.module.scss'

const Synonym: FC = () => {
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [searchText, setSearchText] = useState('')
  const [searchWord, setSearchWord] = useState('')
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<any>(null)
  const [detailData, setDetailData] = useState<any>(null)

  // 获取列表数据
  const {
    data,
    loading,
    run: refresh,
  } = useRequest(
    () =>
      listSynonymKeywords({
        limit: pageSize,
        offset: (page - 1) * pageSize,
        word: searchWord.trim() || undefined,
      }),
    {
      refreshDeps: [page, pageSize, searchWord],
    },
  )

  const handleSearch = () => {
    setSearchWord(searchText)
    setPage(1)
  }

  const handleReset = () => {
    setSearchText('')
    setSearchWord('')
    setPage(1)
  }

  const handleAdd = () => {
    setEditingItem(null)
    setDetailData(null)
    setIsModalOpen(true)
  }

  const handleEdit = async (record: any) => {
    try {
      const res = await getSynonymKeyword({ id: record.ID })
      setEditingItem(record)
      setDetailData(res.data)
      setIsModalOpen(true)
    } catch (error) {
      console.log('获取详情失败', error)
    }
  }

  const handleDelete = (record: any) => {
    setEditingItem(record)
    setIsDeleteModalOpen(true)
  }

  const handleConfirmDelete = async () => {
    if (!editingItem) return
    try {
      await deleteSynonymKeyword({ id: editingItem.ID })
      message.success('删除成功')
      setIsDeleteModalOpen(false)
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
      width: 100,
      render: (_: any, __: any, index: number) =>
        (page - 1) * pageSize + index + 1,
    },
    {
      title: '主词',
      dataIndex: 'word',
      key: 'word',
      width: 200,
    },
    {
      title: '同义词',
      dataIndex: 'synonym_keywords',
      key: 'synonym_keywords',
      render: (synonyms: any[]) => {
        if (!synonyms || synonyms.length === 0) return '-'
        const text = synonyms.map((s) => s.word).join(', ')
        return (
          <Tooltip title={text}>
            <div className='truncate max-w-[300px]'>{text}</div>
          </Tooltip>
        )
      },
    },
    {
      title: '创建时间',
      dataIndex: 'CreatedAt',
      key: 'CreatedAt',
      width: 250,
      render: (text: string) =>
        text ? dayjs(text).format('YYYY-MM-DD HH:mm:ss') : '-',
    },
    {
      title: '创建人',
      dataIndex: 'user_name',
      key: 'user_name',
      width: 200,
      render: (user_name: string) => (
        <span className='text-gray-500'>{user_name || '-'}</span>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      render: (_: any, record: any) => (
        <Space size='middle'>
          <Button
            type='link'
            className='p-0 h-auto text-[#0C99FF] hover:text-[#38b2ff] flex items-center gap-1'
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Button
            type='link'
            danger
            className='p-0 h-auto flex items-center gap-1'
            onClick={() => handleDelete(record)}
          >
            删除
          </Button>
        </Space>
      ),
    },
  ]

  const navigate = useNavigate()

  return (
    <div className='w-full h-full flex flex-col bg-[#FAFAFA]'>
      <div className='w-full h-12 bg-white flex items-center px-5 pt-[2px] border-b border-[#EFF1F4]'>
        <Breadcrumb
          className={styles.layoutHeader}
          separator={<SeparatorIcon />}
          items={[
            {
              title: (
                <span
                  className='text-sm cursor-pointer hover:text-[#0C99FF] font-medium text-[#3C4149]'
                  onClick={() => {
                    navigate('/')
                  }}
                >
                  问答
                </span>
              ),
            },
            {
              title: (
                <span className='cursor-pointer text-sm font-medium text-[#3C4149]'>
                  同义词管理
                </span>
              ),
            },
          ]}
        />
      </div>
      <div className='flex-1 overflow-auto p-6 bg-[#FAFAFA]'>
        <div className='flex flex-col h-full gap-4'>
          <div className='flex justify-between items-center'>
            <Space size='middle'>
              <Input
                placeholder='请输入主词'
                prefix={<SearchOutlined className='text-gray-400' />}
                value={searchText}
                onChange={(e) => {
                  const value = e.target.value
                  if (value.length <= 20) {
                    setSearchText(value)
                  }
                }}
                onPressEnter={handleSearch}
                onBlur={handleSearch}
                onClear={() => {
                  setSearchText('')
                  setSearchWord('')
                  setPage(1)
                }}
                maxLength={20}
                style={{ width: 250 }}
                allowClear
              />
              <Button onClick={handleReset}>重置</Button>
              <Button
                type='primary'
                icon={<PlusOutlined />}
                onClick={handleAdd}
                className='bg-[#0C99FF]'
              >
                创建同义词组
              </Button>
            </Space>
          </div>

          <div className='flex-1 bg-white rounded-lg shadow-sm overflow-hidden flex flex-col'>
            <Table
              columns={columns}
              dataSource={data?.data || []}
              loading={loading}
              pagination={false}
              rowKey='ID'
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
        </div>
      </div>

      <SynonymModal
        open={isModalOpen}
        onCancel={() => setIsModalOpen(false)}
        onSuccess={() => {
          setIsModalOpen(false)
          refresh()
        }}
        editingItem={editingItem}
        detailData={detailData}
      />

      <DeleteConfirmModal
        visible={isDeleteModalOpen}
        customText='确定要删除该条记录吗？删除后，问答环节将无法基于该主词及其同义词的关联关系进行问题改写。'
        onCancel={() => setIsDeleteModalOpen(false)}
        onConfirm={handleConfirmDelete}
      />
    </div>
  )
}

export default Synonym
