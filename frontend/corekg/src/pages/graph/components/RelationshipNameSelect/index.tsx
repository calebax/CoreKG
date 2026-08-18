import { FC, useMemo, useState } from 'react'
import { App, Divider, Input, Modal, Select } from 'antd'
import { useBoolean } from 'ahooks'

export type RelationshipNameOption = { label: string; value: string }

type RelationshipNameSelectProps = {
  value?: string
  onChange?: (value: string) => void
  options: RelationshipNameOption[]
  placeholder?: string
  maxLength?: number
  disabled?: boolean
}

export const RelationshipNameSelect: FC<RelationshipNameSelectProps> = (
  props,
) => {
  const {
    value,
    onChange,
    options,
    placeholder,
    maxLength = 20,
    disabled,
  } = props

  const { message } = App.useApp()
  const [dropdownOpen, setDropdownOpen] = useState(false)
  const [modalOpen, modalActions] = useBoolean(false)
  const [customName, setCustomName] = useState('')

  const normalizedOptions = useMemo(() => {
    const map = new Map<string, RelationshipNameOption>()
    ;(options ?? []).forEach((opt) => {
      const nextValue = (opt?.value ?? '').trim()
      if (!nextValue) return
      if (map.has(nextValue)) return
      map.set(nextValue, {
        ...opt,
        value: nextValue,
        label: opt.label ?? nextValue,
      })
    })
    return Array.from(map.values())
  }, [options])

  const optionValueSet = useMemo(() => {
    return new Set(normalizedOptions.map((opt) => opt.value))
  }, [normalizedOptions])

  const handleOpenCustomModal = () => {
    if (disabled) return
    setDropdownOpen(false)
    setCustomName('')
    modalActions.setTrue()
  }

  const handleCustomOk = () => {
    const nextName = (customName ?? '').trim()
    if (!nextName) {
      message.warning('请输入关系名称')
      return
    }
    if (nextName.length > maxLength) {
      message.warning(`关系名称最多 ${maxLength} 个字符`)
      return
    }

    if (optionValueSet.has(nextName)) {
      message.warning('关系名称已存在，请直接选择')
      return
    }

    onChange?.(nextName)
    modalActions.setFalse()
  }

  return (
    <>
      <Select
        value={value}
        disabled={disabled}
        placeholder={placeholder}
        options={normalizedOptions}
        showSearch={false}
        open={dropdownOpen}
        onDropdownVisibleChange={setDropdownOpen}
        dropdownRender={(menu) => {
          return (
            <>
              {menu}
              <Divider className='my-2' />
              <div
                className='px-3 pb-2 text-[#1677ff] cursor-pointer select-none'
                role='button'
                tabIndex={0}
                aria-label='自定义关系'
                onMouseDown={(e) => {
                  e.preventDefault()
                }}
                onClick={handleOpenCustomModal}
                onKeyDown={(e) => {
                  if (e.key !== 'Enter' && e.key !== ' ') return
                  e.preventDefault()
                  handleOpenCustomModal()
                }}
              >
                自定义关系
              </div>
            </>
          )
        }}
        onChange={(next) => {
          if (!next) return
          onChange?.(String(next))
        }}
      />

      <Modal
        title='新建关系'
        open={modalOpen}
        onCancel={modalActions.setFalse}
        onOk={handleCustomOk}
        okText='确定'
        cancelText='取消'
      >
        <Input
          autoFocus
          value={customName}
          placeholder='请输入关系名称'
          maxLength={maxLength}
          showCount
          onChange={(e) => setCustomName(e.target.value)}
          onKeyDown={(e) => {
            if (e.key !== 'Enter') return
            e.preventDefault()
            handleCustomOk()
          }}
        />
      </Modal>
    </>
  )
}
