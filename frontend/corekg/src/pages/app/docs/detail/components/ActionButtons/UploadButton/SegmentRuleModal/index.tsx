import { FC, useState, useCallback } from 'react'
import {
  Modal,
  Radio,
  InputNumber,
  Select,
  Slider,
  Checkbox,
  Tooltip,
  Space,
  Row,
  Col,
  Button,
  RadioChangeEvent,
} from 'antd'
import { QuestionCircleOutlined } from '@ant-design/icons'
import WarningIcon from '@/assets/icons/docs/warning.svg?react'
import scrollStyles from '@/styles/scroll/styles.module.scss'
import '@/styles/segmentRules.css'

export interface SegmentRule {
  type: 'auto' | 'rule'
  segmentLength?: number
  segmentSeparator?: string
  segmentOverlap?: number
  textPreprocessing?: {
    removeExtraSpaces: boolean
    removeLineBreaks: boolean
    removeSpecialChars: boolean
  }
}

interface SegmentRuleModalProps {
  open: boolean
  onOk: (rule: SegmentRule) => void
  onCancel: () => void
  uploadType: 'file' | 'folder'
}

const SEPARATOR_OPTIONS = [
  { label: '换行', value: '\n' },
  { label: '2个换行', value: '\n\n' },
  { label: '中文句号', value: '。' },
  { label: '英文句号', value: '.' },
  { label: '中文感叹号', value: '！' },
  { label: '英文感叹号', value: '!' },
  { label: '中文问号', value: '？' },
]

export const SegmentRuleModal: FC<SegmentRuleModalProps> = ({
  open,
  onOk,
  onCancel,
  uploadType,
}) => {
  const [ruleType, setRuleType] = useState<'auto' | 'rule'>('auto')
  const [segmentLength, setSegmentLength] = useState(256)
  const [segmentSeparator, setSegmentSeparator] = useState('\n')
  const [segmentOverlap, setSegmentOverlap] = useState(25)
  const [textPreprocessing, setTextPreprocessing] = useState({
    removeExtraSpaces: false,
    removeLineBreaks: false,
    removeSpecialChars: false,
  })
  const [selectAll, setSelectAll] = useState(false)

  const handleRuleTypeChange = useCallback((e: RadioChangeEvent) => {
    const value = e.target.value as 'auto' | 'rule'
    setRuleType(value)
    if (value === 'rule') {
      // 切换到自定义规则时重置表单为默认值
      setSegmentLength(256)
      setSegmentSeparator('\n')
      setSegmentOverlap(25)
      setTextPreprocessing({
        removeExtraSpaces: false,
        removeLineBreaks: false,
        removeSpecialChars: false,
      })
      setSelectAll(false)
    }
  }, [])

  const handleSegmentLengthChange = useCallback((value: number | null) => {
    if (value === null) return
    if (value < 256) {
      setSegmentLength(256)
    } else if (value > 1024) {
      setSegmentLength(1024)
    } else {
      setSegmentLength(value)
    }
  }, [])

  const handleSelectAllChange = useCallback((checked: boolean) => {
    setSelectAll(checked)
    setTextPreprocessing({
      removeExtraSpaces: checked,
      removeLineBreaks: checked,
      removeSpecialChars: checked,
    })
  }, [])

  const handlePreprocessingChange = useCallback(
    (field: keyof typeof textPreprocessing) => {
      return (checked: boolean) => {
        const newTextPreprocessing = {
          ...textPreprocessing,
          [field]: checked,
        }
        setTextPreprocessing(newTextPreprocessing)

        // 检查是否全部选中
        const allSelected = Object.values(newTextPreprocessing).every(Boolean)
        const noneSelected = Object.values(newTextPreprocessing).every(
          (val) => !val,
        )
        setSelectAll(allSelected || (!noneSelected && selectAll))
      }
    },
    [textPreprocessing, selectAll],
  )

  const handleOk = useCallback(() => {
    const rule: SegmentRule = {
      type: ruleType,
    }

    if (ruleType === 'rule') {
      rule.segmentLength = segmentLength
      rule.segmentSeparator = segmentSeparator
      rule.segmentOverlap = segmentOverlap
      rule.textPreprocessing = textPreprocessing
    }

    onOk(rule)
  }, [
    ruleType,
    segmentLength,
    segmentSeparator,
    segmentOverlap,
    textPreprocessing,
    onOk,
  ])

  const handleCancel = useCallback(() => {
    // 重置状态
    setRuleType('auto')
    setSegmentLength(256)
    setSegmentSeparator('\n')
    setSegmentOverlap(25)
    setTextPreprocessing({
      removeExtraSpaces: false,
      removeLineBreaks: false,
      removeSpecialChars: false,
    })
    setSelectAll(false)
    onCancel()
  }, [onCancel])

  return (
    <Modal
      title={
        <span className='text-[18px] text-[#1E1F28] font-medium'>分段规则</span>
      }
      open={open}
      onOk={handleOk}
      onCancel={handleCancel}
      okText='保存并导入'
      cancelText={null}
      width={750}
      centered
      closable
      className='segment-rules-modal'
      style={{ maxHeight: '90vh' }}
      bodyStyle={{
        maxHeight: 'calc(90vh - 140px)',
        overflow: 'hidden',
        padding: 0,
      }}
      footer={[
        <Button
          key='submit'
          type='primary'
          onClick={handleOk}
          className='!bg-[#3d7fff] !py-[5px] !px-3 !text-[16px] !font-medium !text-[#fcfcfe] !rounded-[4px] !leading-[24px] !min-w-20'
        >
          保存并导入
        </Button>,
      ]}
    >
      <div
        className={`flex flex-col gap-6 h-full overflow-y-auto ${scrollStyles.scroll} pr-2`}
        style={{ maxHeight: 'calc(90vh - 140px)' }}
      >
        {/* 提醒信息 */}
        <div className='flex items-start gap-2 px-3 py-[9px] bg-[#f8f9fd] rounded-[10px]'>
          <div className='w-4 h-4 mt-[3px] rounded-full flex items-center justify-center flex-shrink-0'>
            <WarningIcon className='w-4 h-4' />
          </div>
          <div className='text-[13px] text-[#1E1F28] leading-[22px] font-normal fontFamily-pingFangSC'>
            推荐使用"默认规则"：系统基于自研多模态解析与层级语义自适应分段，自动识别表格/公式/图表及跨页内容并完成预处理，保证上下文连续可用。
          </div>
        </div>

        {/* 分段规则选择 */}
        <div>
          <Radio.Group
            value={ruleType}
            onChange={handleRuleTypeChange}
            className='w-full'
          >
            <div className='flex flex-col gap-4'>
              {/* 默认规则 - 推荐选项 */}
              <div
                className={`relative py-5 px-8 rounded-[10px] border h-[100px] flex flex-col justify-center cursor-pointer ${
                  ruleType === 'auto'
                    ? 'bg-[#eaf2ff] border-[#3d7fff] shadow-[0px_0px_21.8px_0px_rgba(125,168,248,0.29)]'
                    : 'bg-white border-[#d7d9e5]'
                }`}
                onClick={() => setRuleType('auto')}
                role='button'
                tabIndex={0}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault()
                    setRuleType('auto')
                  }
                }}
                aria-label='选择默认规则'
              >
                {
                  <div className='absolute top-0 left-0 bg-[#0C99FF] shadow-[0px_4px_5px_0px_rgba(125,168,248,0.4)] rounded-tl-[10px] rounded-br-[10px] px-1 py-0.5 w-[55px] flex items-center justify-center'>
                    <span className='text-xs text-white leading-4'>推荐</span>
                  </div>
                }
                <div className='flex gap-3 items-center mb-2.5'>
                  <Radio value='auto' />
                  <div className='text-[18px] text-[#1E1F28] font-medium leading-normal'>
                    默认规则
                  </div>
                </div>
                <div className='flex gap-3 items-center'>
                  <div className='w-5 h-5'></div>
                  <div className='text-[16px] text-[#616373] leading-6 font-normal fontFamily-pingFangSC'>
                    智能识别段落和句子边界
                  </div>
                </div>
              </div>

              {/* 自定义规则 */}
              <div
                className={`py-5 px-8 rounded-[10px] border h-[100px] flex flex-col justify-center cursor-pointer ${
                  ruleType === 'rule'
                    ? 'bg-[#eaf2ff] border-[#0C99FF] shadow-[0px_0px_21.8px_0px_rgba(125,168,248,0.29)]'
                    : 'bg-white border-[#d7d9e5]'
                }`}
                onClick={() => setRuleType('rule')}
                role='button'
                tabIndex={0}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault()
                    setRuleType('rule')
                  }
                }}
                aria-label='选择自定义规则'
              >
                <div className='flex gap-3 items-center mb-2.5'>
                  <Radio value='rule' />
                  <div className='text-[18px] text-[#1E1F28] font-medium leading-normal'>
                    自定义规则
                  </div>
                </div>
                <div className='flex gap-3 items-center'>
                  <div className='w-5 h-5'></div>
                  <div className='text-[16px] text-[#616373] leading-6 font-normal fontFamily-pingFangSC'>
                    手动调整段长、重叠与分隔符
                  </div>
                </div>
              </div>
            </div>
          </Radio.Group>
        </div>

        {/* 自定义规则表单 */}
        {ruleType === 'rule' && (
          <div className='flex flex-col gap-6'>
            {/* 自定义规则说明 */}
            <div className='flex flex-col gap-[6px] text-[14px] text-[#616373] leading-5 font-normal fontFamily-pingFangSC '>
              <p>使用自定义拆分时，问答不支持跳转到原文位置。</p>
              <p>
                图片、音频、视频等非文本文件，系统仅使用推荐分段，暂不支持自定义。
              </p>
            </div>
            {/* 分段长度 */}
            <div className='flex flex-col gap-2'>
              <div className='flex items-center pb-2'>
                <span className='text-[#D54941] mr-1'>*</span>
                <span className='text-sm text-[#1E1F28] font-normal'>
                  分段长度
                </span>
                <span className='text-sm text-[#86909C] font-normal'>
                  （请输入256~1024之间的数值）
                </span>
              </div>
              <div className='w-full'>
                <div className='relative w-full'>
                  <InputNumber
                    value={segmentLength}
                    onChange={handleSegmentLengthChange}
                    min={256}
                    max={1024}
                    className='w-full'
                    controls={false}
                    placeholder='256'
                    style={{
                      borderRadius: '4px',
                      paddingRight: '40px',
                      position: 'relative',
                      boxSizing: 'border-box',
                    }}
                  />
                  <span className='absolute right-[1px] top-[1px] transform w-[38px] h-[30px] rounded-r border-l border-0 border-[#D7D9E5] flex items-center justify-center bg-[#F7F8FA] text-sm text-[#1E1F28] pointer-events-none'>
                    字
                  </span>
                </div>
              </div>
            </div>

            {/* 分段标识符 */}
            <div className='flex flex-col gap-2'>
              <div className='flex items-center pb-2'>
                <span className='text-[#D54941] mr-1'>*</span>
                <span className='text-sm text-[#1E1F28] font-normal'>
                  分段标识符
                </span>
                <span className='text-sm text-[#86909C] font-normal'>
                  （系统未识别到标识符时，将按照设置的分段长度来分割文本）
                </span>
              </div>
              <Select
                value={segmentSeparator}
                onChange={setSegmentSeparator}
                className='w-full'
                options={SEPARATOR_OPTIONS}
                placeholder='请选择'
              />
            </div>

            {/* 分段重叠度 */}
            <div className='flex flex-col gap-2'>
              <div className='flex items-center pb-2'>
                <span className='text-[#D54941] mr-1'>*</span>
                <span className='text-sm text-[#1E1F28] font-normal'>
                  分段重叠度%
                </span>
                <span className='text-sm text-[#86909C] font-normal'>
                  （设置相邻两个chunk之间重叠部分所占比例，提升召回效果）
                </span>
                <Tooltip
                  title='分段重叠度=10%对应2个字，即前一段的末尾2个字会重复出现在下一段。'
                  placement='topRight'
                >
                  <QuestionCircleOutlined className='text-[#999999] text-base cursor-help' />
                </Tooltip>
              </div>
              <div className='w-full'>
                {/* <div className='flex justify-between items-center mb-2'>
                  <span className='text-sm text-[#1E1F28] font-normal'>
                    分段重叠度
                  </span>
                </div> */}
                <Slider
                  value={segmentOverlap}
                  onChange={setSegmentOverlap}
                  min={1}
                  max={90}
                  tooltip={{
                    placement: 'top',
                    formatter: (value) => `${value}%`,
                  }}
                  className='w-full'
                />
              </div>
            </div>
            {/* 文本预处理规则 */}
            <div>
              <div className='flex items-center mb-3'>
                <Checkbox
                  checked={selectAll}
                  indeterminate={
                    Object.values(textPreprocessing).some(Boolean) &&
                    !Object.values(textPreprocessing).every(Boolean)
                  }
                  onChange={(e) => handleSelectAllChange(e.target.checked)}
                  className='text-sm font-medium text-gray-900'
                >
                  文本预处理规则
                </Checkbox>
              </div>
              <div className='pl-0 flex flex-col gap-3'>
                <Checkbox
                  checked={textPreprocessing.removeExtraSpaces}
                  onChange={(e) =>
                    handlePreprocessingChange('removeExtraSpaces')(
                      e.target.checked,
                    )
                  }
                  className='text-sm text-[#1E1F28]'
                >
                  替换掉连续的空格、换行符和制表符
                </Checkbox>
                <Checkbox
                  checked={textPreprocessing.removeLineBreaks}
                  onChange={(e) =>
                    handlePreprocessingChange('removeLineBreaks')(
                      e.target.checked,
                    )
                  }
                  className='text-sm text-[#1E1F28]'
                >
                  删除所有URL
                </Checkbox>
                <Checkbox
                  checked={textPreprocessing.removeSpecialChars}
                  onChange={(e) =>
                    handlePreprocessingChange('removeSpecialChars')(
                      e.target.checked,
                    )
                  }
                  className='text-sm text-[#1E1F28]'
                >
                  删除所有电子邮箱地址
                </Checkbox>
              </div>
            </div>
          </div>
        )}
      </div>
    </Modal>
  )
}
