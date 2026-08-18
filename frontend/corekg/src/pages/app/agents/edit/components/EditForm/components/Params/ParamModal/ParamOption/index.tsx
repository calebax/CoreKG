import { FC } from 'react'
import { Button, Form, Input } from 'antd'
import { DeleteOutlined, PlusOutlined } from '@ant-design/icons'
import { cn } from '@/utils'

export const ParamOption: FC<Style> = (props) => {
  return (
    <div className={cn('flex flex-col gap-2', props.style)} style={props.style}>
      <Form.Item label='选项' required>
        <Form.List
          name='input_array'
          rules={[
            {
              validator: async (_, val) => {
                if (!val || val.length === 0) {
                  throw new Error('至少要有一个选项')
                }
              },
            },
          ]}
        >
          {(fields, operation, { errors }) => (
            <>
              {fields.map((field) => {
                return (
                  <Form.Item
                    {...field}
                    rules={[{ required: true, message: '选项值不能为空' }]}
                  >
                    <Input
                      suffix={
                        <DeleteOutlined
                          onClick={() => operation.remove(field.name)}
                        />
                      }
                    />
                  </Form.Item>
                )
              })}
              <Button icon={<PlusOutlined />} onClick={() => operation.add()}>
                添加选项
              </Button>
              <Form.ErrorList errors={errors} />
            </>
          )}
        </Form.List>
      </Form.Item>
    </div>
  )
}
