import { FC, useState, useCallback } from 'react'
import { Select, InputNumber, Checkbox, Button } from 'antd'
import { CloseOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import scrollStyles from '@/styles/scroll/styles.module.scss'
import styles from './styles.module.scss'

export interface ModifySegmentRule {
  type: 'default' | 'custom'
  segmentLength?: number
  segmentSeparator?: string
  segmentOverlap?: number
  textPreprocessing?: {
    removeExtraSpaces: boolean
    removeLineBreaks: boolean
    removeSpecialChars: boolean
  }
}

interface ModifySegmentRuleModalProps {
  open: boolean
  onOk: (rule: ModifySegmentRule) => void
  onCancel: () => void
  initialRule?: ModifySegmentRule
}

export const ModifySegmentRuleModal: FC<ModifySegmentRuleModalProps> = ({
  open,
  onOk,
  onCancel,
  initialRule,
}) => {
  const { t } = useTranslation('pages')

  const getRuleTypeOptions = () => [
    { label: t('app.docs.fileDetail.ruleTypeDefault'), value: 'default' },
    { label: t('app.docs.fileDetail.ruleTypeCustom'), value: 'custom' },
  ]

  const getSeparatorOptions = () => [
    { label: t('app.docs.fileDetail.separatorLineBreak'), value: '\n' },
    { label: t('app.docs.fileDetail.separatorTwoLineBreaks'), value: '\n\n' },
    { label: t('app.docs.fileDetail.separatorChinesePeriod'), value: '。' },
    { label: t('app.docs.fileDetail.separatorEnglishPeriod'), value: '.' },
    {
      label: t('app.docs.fileDetail.separatorChineseExclamation'),
      value: '！',
    },
    { label: t('app.docs.fileDetail.separatorEnglishExclamation'), value: '!' },
    { label: t('app.docs.fileDetail.separatorChineseQuestion'), value: '？' },
  ]

  const [ruleType, setRuleType] = useState<'default' | 'custom'>(
    initialRule?.type || 'default',
  )
  const [segmentLength, setSegmentLength] = useState(
    initialRule?.segmentLength || 256,
  )
  const [segmentSeparator, setSegmentSeparator] = useState(
    initialRule?.segmentSeparator || '\n',
  )
  const [segmentOverlap, setSegmentOverlap] = useState(
    initialRule?.segmentOverlap ?? 30,
  )
  const [textPreprocessing, setTextPreprocessing] = useState({
    removeExtraSpaces:
      initialRule?.textPreprocessing?.removeExtraSpaces ?? false,
    removeLineBreaks: initialRule?.textPreprocessing?.removeLineBreaks ?? false,
    removeSpecialChars:
      initialRule?.textPreprocessing?.removeSpecialChars ?? true,
  })

  const handleRuleTypeChange = useCallback(
    (value: 'default' | 'custom') => {
      setRuleType(value)
      if (value === 'custom') {
        // 切换到自定义规则时设置默认值
        setSegmentLength(initialRule?.segmentLength || 256)
        setSegmentSeparator(initialRule?.segmentSeparator || '\n')
        setSegmentOverlap(initialRule?.segmentOverlap || 30)
      }
    },
    [initialRule],
  )

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

  const handleSegmentOverlapChange = useCallback((value: number | null) => {
    if (value === null) return
    if (value < 0) {
      setSegmentOverlap(0)
    } else if (value > 100) {
      setSegmentOverlap(100)
    } else {
      setSegmentOverlap(value)
    }
  }, [])

  const handlePreprocessingChange = useCallback(
    (field: keyof typeof textPreprocessing) => {
      return (checked: boolean) => {
        setTextPreprocessing((prev) => ({
          ...prev,
          [field]: checked,
        }))
      }
    },
    [],
  )

  const handleOk = useCallback(() => {
    // 构建当前规则配置
    const currentRule: ModifySegmentRule = {
      type: ruleType,
    }

    if (ruleType === 'custom') {
      currentRule.segmentLength = segmentLength
      currentRule.segmentSeparator = segmentSeparator
      currentRule.segmentOverlap = segmentOverlap
      currentRule.textPreprocessing = textPreprocessing
    }

    // 检查是否有变更
    const hasChanges = () => {
      // 比较规则类型
      if (currentRule.type !== initialRule?.type) {
        return true
      }

      // 如果是自定义规则，比较所有自定义字段
      if (currentRule.type === 'custom') {
        if (
          currentRule.segmentLength !== initialRule?.segmentLength ||
          currentRule.segmentSeparator !== initialRule?.segmentSeparator ||
          currentRule.segmentOverlap !== initialRule?.segmentOverlap
        ) {
          return true
        }

        // 比较文本预处理规则
        const currentPreprocessing = currentRule.textPreprocessing
        const initialPreprocessing = initialRule?.textPreprocessing
        if (
          currentPreprocessing?.removeExtraSpaces !==
            initialPreprocessing?.removeExtraSpaces ||
          currentPreprocessing?.removeLineBreaks !==
            initialPreprocessing?.removeLineBreaks ||
          currentPreprocessing?.removeSpecialChars !==
            initialPreprocessing?.removeSpecialChars
        ) {
          return true
        }
      }

      return false
    }

    // 如果没有变更，直接关闭弹窗
    if (!hasChanges()) {
      onCancel()
      return
    }

    // 有变更，调用接口
    onOk(currentRule)
  }, [
    ruleType,
    segmentLength,
    segmentSeparator,
    segmentOverlap,
    textPreprocessing,
    initialRule,
    onOk,
    onCancel,
  ])

  const handleCancel = useCallback(() => {
    // 重置状态为初始值
    setRuleType(initialRule?.type || 'default')
    setSegmentLength(initialRule?.segmentLength || 256)
    setSegmentSeparator(initialRule?.segmentSeparator || '\n')
    setSegmentOverlap(initialRule?.segmentOverlap ?? 30)
    setTextPreprocessing({
      removeExtraSpaces:
        initialRule?.textPreprocessing?.removeExtraSpaces ?? false,
      removeLineBreaks:
        initialRule?.textPreprocessing?.removeLineBreaks ?? false,
      removeSpecialChars:
        initialRule?.textPreprocessing?.removeSpecialChars ?? true,
    })
    onCancel()
  }, [initialRule, onCancel])

  if (!open) return null

  return (
    <div className='fixed inset-0 z-50 flex items-center justify-center'>
      {/* 背景遮罩 */}
      <div
        className='absolute inset-0 bg-[rgba(0,0,0,0.45)] transition-opacity duration-300'
        onClick={handleCancel}
      />

      {/* 弹窗内容 */}
      <div className='modify-segment-rule-modal relative bg-white rounded-lg shadow-xl w-[520px] max-w-[90vw] max-h-[90vh] overflow-hidden animate-in fade-in-0 zoom-in-95 duration-300'>
        {/* 标题栏 */}
        <div className='flex items-center justify-between mx-4 py-4 border-b border-gray-200'>
          <h2 className="font-['Inter'] text-[#1e1f28] text-[16px] leading-[24px] font-medium">
            {t('app.docs.fileDetail.editRules')}
          </h2>
          <button
            onClick={handleCancel}
            className='cursor-pointer p-1 hover:bg-gray-100 rounded transition-colors'
          >
            <CloseOutlined className='text-[#616373] text-base' />
          </button>
        </div>

        {/* 内容区域 */}
        <div
          className={`${scrollStyles.scroll} px-4 py-4`}
          style={{ minHeight: '320px', maxHeight: '60vh' }}
        >
          <style>
            {`
              .modify-segment-rule-modal .ant-select-selector {
                background-color: #F5F5F5 !important;
              }
              .modify-segment-rule-modal .ant-checkbox-checked .ant-checkbox-inner {
                background-color: #099CFF !important;
                border-color: #099CFF !important;
              }
              .modify-segment-rule-modal .ant-checkbox:hover .ant-checkbox-inner {
                border-color: #099CFF !important;
              }
              .modify-segment-rule-modal .ant-checkbox-checked::after {
                border-color: #099CFF !important;
              }
              .modify-segment-rule-modal .ant-checkbox .ant-checkbox-inner:after {
                border-color: #fff !important;
              }
            `}
          </style>
          <div className='flex flex-col gap-6'>
            {/* Rule types */}
            <div>
              <label className='block text-sm font-medium text-[#919497] mb-2.5'>
                {t('app.docs.fileDetail.ruleTypes')}
              </label>
              <Select
                value={ruleType}
                onChange={handleRuleTypeChange}
                options={getRuleTypeOptions()}
                className={`w-full ${styles.ruleSelect}`}
                size='large'
                style={{
                  backgroundColor: '#F5F5F5 !important',
                }}
              />
            </div>

            {/* Custom rule settings */}
            {ruleType === 'custom' && (
              <div className='flex flex-col gap-5'>
                {/* Fractional length */}
                <div>
                  <label className='block text-sm font-medium text-[#919497] mb-2.5'>
                    {t('app.docs.fileDetail.fractionalLength')} (256-1024)
                  </label>
                  <InputNumber
                    value={segmentLength}
                    onChange={handleSegmentLengthChange}
                    min={256}
                    max={1024}
                    className='w-full'
                    size='large'
                    controls={false}
                  />
                </div>

                {/* Fractional identifier */}
                <div>
                  <label className='block text-sm font-medium text-[#919497] mb-2.5'>
                    {t('app.docs.fileDetail.fractionalIdentifier')}
                  </label>
                  <Select
                    value={segmentSeparator}
                    onChange={setSegmentSeparator}
                    options={getSeparatorOptions()}
                    className={`w-full ${styles.ruleSelect}`}
                    size='large'
                    style={{
                      backgroundColor: '#F5F5F5 !important',
                    }}
                  />
                </div>

                {/* Fractional length (%) */}
                <div>
                  <label className='block text-sm font-medium text-[#919497] mb-2.5'>
                    {t('app.docs.fileDetail.fractionalLengthPercent')} (%)
                  </label>
                  <div className='relative'>
                    <InputNumber
                      value={segmentOverlap}
                      onChange={handleSegmentOverlapChange}
                      min={0}
                      max={100}
                      className='w-full'
                      size='large'
                      controls={false}
                    />
                    {/* <span className='absolute right-3 top-1/2 transform -translate-y-1/2 text-[#6b7280] text-sm pointer-events-none'>
                      %
                    </span> */}
                  </div>
                </div>

                {/* Text processing rules */}
                <div>
                  <label className='block text-sm font-medium text-[#919497] mb-2.5'>
                    {t('app.docs.fileDetail.textProcessingRules')}
                  </label>
                  <div className='flex flex-col gap-3'>
                    <Checkbox
                      checked={textPreprocessing.removeExtraSpaces}
                      onChange={(e) =>
                        handlePreprocessingChange('removeExtraSpaces')(
                          e.target.checked,
                        )
                      }
                      className='text-sm text-[#0C1F17]'
                    >
                      {t('app.docs.fileDetail.replaceConsecutiveSpaces')}
                    </Checkbox>
                    <Checkbox
                      checked={textPreprocessing.removeLineBreaks}
                      onChange={(e) =>
                        handlePreprocessingChange('removeLineBreaks')(
                          e.target.checked,
                        )
                      }
                      className='text-sm text-[#0C1F17]'
                    >
                      {t('app.docs.fileDetail.deleteAllURLs')}
                    </Checkbox>
                    <Checkbox
                      checked={textPreprocessing.removeSpecialChars}
                      onChange={(e) =>
                        handlePreprocessingChange('removeSpecialChars')(
                          e.target.checked,
                        )
                      }
                      className='text-sm text-[#0C1F17]'
                    >
                      {t('app.docs.fileDetail.deleteAllEmailAddresses')}
                    </Checkbox>
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* 底部按钮区域 */}
        <div className='flex justify-end gap-1.5 py-4 border-t border-[#EFF1F4] mx-4'>
          <Button
            onClick={handleCancel}
            className='bg-[#F5F5F5] !text-[#0C1F17] px-4 py-2.5 text-sm font-medium border-none'
          >
            {t('app.docs.fileDetail.cancel')}
          </Button>
          <Button
            type='primary'
            onClick={handleOk}
            className='!bg-[#0C99FF] hover:!bg-[#0C99FF] text-[#fff] text-sm font-medium border-none'
          >
            {t('app.docs.fileDetail.done')}
          </Button>
        </div>
      </div>
    </div>
  )
}
