import { FC } from 'react'
import { Button, Form } from 'antd'
import useFormInstance from 'antd/es/form/hooks/useFormInstance'
import { CloseOutlined, PlusOutlined } from '@ant-design/icons'
import { Agent } from 'Agent'
import { useBoolean } from 'ahooks'
import { cn } from '@/utils'
import { FormItem, useEditContext } from '../..'
import { AdvancedModal } from './AdvancedModal'
import { ForestSelectModal } from './ForestSelectModal'
import { SVG } from './SVG'
import Setting from './setting.svg?react'

/** 选择知识库 */
export const Forests: FC<Style> = (props) => {
  const type = Form.useWatch<Agent['type']>('type', { preserve: true })
  if (type !== 'role_play' && type !== 'knowledge') return null
  return (
    <FormItem name={'forests'}>
      <InnerForests {...props} />
    </FormItem>
  )
}

const InnerForests: FC<Style & ValueController<Agent['forests']>> = (props) => {
  const { forestList } = useEditContext()
  const { value, onChange } = props
  const form = useFormInstance()
  const [forestSelectOpen, { toggle: toggleForestSelectOpen }] =
    useBoolean(false)
  const [advancedModalOpen, { toggle: toggleAdvancedModalOpen }] =
    useBoolean(false)

  const setForest = (val?: Agent['forests']) => {
    if (!val || val.length === 0) {
      onChange?.()
      form.setFieldValue('type', 'role_play')
    } else {
      onChange?.(val)
      form.setFieldValue('type', 'knowledge')
    }
  }

  return (
    <>
      <div
        className={cn('-mt-2 flex flex-col gap-1', props.className)}
        style={props.style}
      >
        <div className='flex items-center justify-between'>
          <span className='text-base text-title font-medium'>知识库</span>
          <Button
            type='link'
            icon={<Setting />}
            className='ml-4 hidden'
            onClick={toggleAdvancedModalOpen}
          >
            高级设置
          </Button>
          <Button
            type='text'
            size='small'
            icon={<PlusOutlined />}
            className='flex items-center gap-1 text-[#0C99FF] text-base font-medium px-[2px]'
            onClick={toggleForestSelectOpen}
          >
            关联
          </Button>
        </div>
        {value?.map((forest) => {
          return (
            <div
              className={cn(
                'h-10 pl-1 pr-2 py-2',
                'bg-white border border-[#0000001A] rounded',
                'flex items-center',
              )}
            >
              <SVG type={forest.forest_type} />
              <span className='text-base text-title'>{forest.name}</span>
              <CloseOutlined
                className='ml-auto text-[#C9CDD4]'
                onClick={() => {
                  const newForest = value?.filter(
                    (item) => item.id !== forest.id,
                  )
                  setForest(newForest)
                }}
              />
            </div>
          )
        })}
      </div>
      {forestSelectOpen ? (
        <ForestSelectModal
          onClose={toggleForestSelectOpen}
          forestList={forestList}
          deafultValue={value ?? []}
          onChange={setForest}
        />
      ) : null}
      {/* 高级设置只能在此处修改 不需要销毁重建组件 */}
      <AdvancedModal
        open={advancedModalOpen}
        onClose={toggleAdvancedModalOpen}
        deafultValue={1}
        onChange={console.log}
      />
    </>
  )
}
