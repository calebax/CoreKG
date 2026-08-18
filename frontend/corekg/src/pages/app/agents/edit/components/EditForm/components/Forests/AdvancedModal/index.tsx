import { FC } from 'react'
import { Input, InputNumber, Modal, Radio, Switch } from 'antd'
import { useImmer } from 'use-immer'
import { AdvancedItem } from './AdvancedItem'

type AdvancedConfig = any

export type AdvancedModal = {
  open: boolean
  onClose: () => void
  // 高级设置
  deafultValue: AdvancedConfig
  onChange: (val: AdvancedConfig) => void
}
export const AdvancedModal: FC<AdvancedModal> = (props) => {
  const { open, onClose, deafultValue, onChange } = props
  const [value, setValue] = useImmer(deafultValue)

  return (
    <Modal open={open} onCancel={onClose} onOk={() => onChange(value)}>
      <div className='flex flex-col'>
        <AdvancedItem title='知识范围'>
          <div className='flex'>
            文档
            <Switch size='small' className='ml-2.5 mr-21' />
            问答
            <Switch size='small' className='ml-2.5' />
          </div>
        </AdvancedItem>

        <AdvancedItem title='意图匹配顺序' tooltip='意图匹配顺序'>
          <Radio.Group
            options={[
              { label: '文档优先', value: '1' },
              { label: '回答优先', value: '2' },
            ]}
          />
        </AdvancedItem>

        <AdvancedItem title='召回策略'>
          <div className='flex'>
            文档召回数量<InputNumber></InputNumber>
            文档检索匹配度<InputNumber></InputNumber>
          </div>
          <div className='flex'>
            问答召回数量<InputNumber></InputNumber>
            问答检索匹配度<InputNumber></InputNumber>
          </div>
        </AdvancedItem>

        <AdvancedItem title='问答回复方式'>
          <Radio.Group
            options={[
              { label: '直接回复答案', value: '1' },
              { label: '大模型润色答案回复', value: '2' },
            ]}
          />
        </AdvancedItem>

        <AdvancedItem title='知识外回答策略'>
          <Radio.Group
            options={[
              { label: '大模型智能回复', value: '1' },
              {
                label: (
                  <div className='flex flex-col gap-1.5'>
                    指定回复
                    <Input.TextArea placeholder='针对您这个问题，我暂时无法回答，请换个问题' />
                  </div>
                ),
                value: '2',
              },
            ]}
          />
        </AdvancedItem>
      </div>
    </Modal>
  )
}
