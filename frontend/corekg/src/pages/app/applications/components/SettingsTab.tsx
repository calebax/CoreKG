import { Button, Form, Input, Popconfirm, message } from 'antd'
import { DeleteOutlined } from '@ant-design/icons'
import type { Application } from '../types'
import styles from './SettingsTab.module.scss'

interface SettingsTabProps {
  app: Application
}

export default function SettingsTab({ app }: SettingsTabProps) {
  const [form] = Form.useForm()

  const handleSave = () => {
    message.success('设置已保存（仅演示）')
  }

  const handleDelete = () => {
    message.warning('删除功能仅演示，未实际执行')
  }

  return (
    <div className={styles.container}>
      <Form
        form={form}
        layout='vertical'
        initialValues={{
          name: app.name,
          description: app.description,
        }}
        style={{ maxWidth: 560 }}
      >
        <Form.Item label='应用名称' name='name' rules={[{ required: true }]}>
          <Input />
        </Form.Item>
        <Form.Item label='描述' name='description'>
          <Input.TextArea rows={3} />
        </Form.Item>
        <Form.Item>
          <Button type='primary' onClick={handleSave}>
            保存设置
          </Button>
        </Form.Item>
      </Form>

      <div className={styles.dangerZone}>
        <h4 className={styles.dangerTitle}>危险区域</h4>
        <Popconfirm
          title='确定要删除此应用吗？'
          description='删除后不可恢复，所有数据将被永久删除。'
          okText='确认删除'
          okType='danger'
          cancelText='取消'
          onConfirm={handleDelete}
        >
          <Button danger icon={<DeleteOutlined />}>
            删除应用
          </Button>
        </Popconfirm>
      </div>
    </div>
  )
}
