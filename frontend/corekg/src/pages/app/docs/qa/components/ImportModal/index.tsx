import { FC } from 'react'
import {
  App,
  Button,
  Modal,
  ModalProps,
  Skeleton,
  Table,
  Typography,
  Upload,
} from 'antd'
import { DownloadOutlined, UploadOutlined } from '@ant-design/icons'
import { useRequest } from 'ahooks'
import { useTranslation } from 'react-i18next'
import { spliteFileName } from '@/utils'
import { commitQAPair, uploadQAPair } from '@/api/knowledge'
import { useDeployConfig } from '@/utils/useDeployConfig'

const legalExtensions = ['xlsx', 'csv']
const QA_IMPORT_MAX_FILE_COUNT = 1
const QA_IMPORT_MAX_SIZE_MB = 100
const QA_IMPORT_MAX_SIZE_BYTES = QA_IMPORT_MAX_SIZE_MB * 1024 * 1024
const QA_IMPORT_MAX_SIZE_LABEL = `${QA_IMPORT_MAX_SIZE_MB}MB`
export type ImportModal = InnerImportModal &
  Pick<ModalProps, 'open' | 'onCancel'>
export const ImportModal: FC<ImportModal> = (props) => {
  const { t: tC } = useTranslation('common')

  const { open, onCancel, forest_id, onSuccess } = props

  return (
    <Modal
      title={tC('button.import', { target: tC('resource.data') })}
      footer={null}
      open={open}
      onCancel={onCancel}
      centered
      width={1080}
    >
      <InnerImportModal
        forest_id={forest_id}
        onSuccess={onSuccess}
        onCancel={onCancel}
      />
    </Modal>
  )
}

type InnerImportModal = {
  forest_id: number
  onSuccess?: () => void
  onCancel?: () => void
}
const InnerImportModal: FC<ImportModal> = (props) => {
  const { t } = useTranslation('pages')
  const { t: tM } = useTranslation('messages')
  const { t: tC } = useTranslation('common')
  const { forest_id, onSuccess, onCancel } = props
  const { message } = App.useApp()
  const { upload, qa_list, uploading } = useUploadFile(forest_id)
  const { commit, commiting } = useCommitQAList(forest_id, onSuccess)

  if (uploading) return <Skeleton active />
  if (!qa_list)
    return (
      <div className='flex gap-4 min-h-[300px]'>
        <ExcelTemplateContent />
        <Upload.Dragger
          accept={['xlsx', 'xls', 'csv'].map((e) => `.${e}`).join(',')}
          className='flex-1 rounded-lg [&_.ant-upload]:!h-full [&_.ant-upload-drag]:!min-h-[300px] [&_.ant-upload-drag]:!border-[#E6E8F0] hover:[&_.ant-upload-drag]:!border-[#2F51FF]'
          beforeUpload={(file) => {
            const { ext } = spliteFileName(file.name)
            if (!legalExtensions.includes(ext)) {
              message.error(tM('uploadSupport', { target: 'xlsx、xls、csv' }))
            } else if (file.size > QA_IMPORT_MAX_SIZE_BYTES) {
              message.error(
                tM('fileSizeLimit', { target: QA_IMPORT_MAX_SIZE_LABEL }),
              )
            } else {
              upload(file)
            }
            return false
          }}
        >
          <div className='w-full min-h-[300px] flex flex-col items-center justify-center gap-2 py-10'>
            <UploadOutlined className='text-[20px] text-[#8B9099]' />
            <span className='text-sm font-medium text-[#0C99FF]'>
              {t('app.docs.dragOrClickUpload')}
            </span>
            <span className='text-sm text-[#ABAFB2] font-normal'>
              {t('app.docs.qaImportUploadHint', {
                count: QA_IMPORT_MAX_FILE_COUNT,
                size: QA_IMPORT_MAX_SIZE_LABEL,
              })}
            </span>
          </div>
        </Upload.Dragger>
      </div>
    )

  return (
    <div className='flex flex-col max-h-[60vh]'>
      <Table
        className='flex-1 overflow-auto'
        pagination={false}
        size='small'
        dataSource={qa_list}
        rowKey='question'
        columns={[
          {
            title: t('app.docs.question'),
            render: (_, record) => {
              return (
                <Typography.Paragraph
                  ellipsis={{ rows: 1, tooltip: record.question }}
                >
                  {record.question}
                </Typography.Paragraph>
              )
            },
          },
          {
            title: t('app.docs.answer'),
            render: (_, record) => {
              return (
                <Typography.Paragraph
                  ellipsis={{ rows: 1, tooltip: record.answer }}
                >
                  {record.answer}
                </Typography.Paragraph>
              )
            },
          },
        ]}
      />
      <span className='flex-none flex justify-end gap-2.5 mt-4'>
        <Button onClick={onCancel}>{tC('button.cancel')}</Button>
        <Button
          loading={commiting}
          type='primary'
          onClick={() => commit(qa_list)}
        >
          {tC('button.create')}
        </Button>
      </span>
    </div>
  )
}

const ExcelTemplateContent: FC = () => {
  const { t } = useTranslation('pages')
  const { t: tC } = useTranslation('common')
  const { t: tM } = useTranslation('messages')
  const { version } = useDeployConfig()
  return (
    <div className='flex-1 min-h-[300px] rounded-lg border border-[#E6E8F0] p-4 flex flex-col'>
      <div className='p-2.5 flex flex-col gap-2.5'>
        <span className='text-base text-title'>
          {t('app.docs.downloadUploadTemplate')}
        </span>
        <span className='text-base text-description whitespace-pre-wrap'>
          {t('app.docs.downloadTemplateFillUpload', {
            target: legalExtensions.join('、'),
          })}
        </span>
        <span className='text-base text-description whitespace-pre-wrap'>
          {t('app.docs.qaImportRowLimitRefTemplate', { target: '100' })}
        </span>
      </div>
      <div className='rounded bg-[#F5F9FF] p-2.5 flex flex-col gap-2.5'>
        <span className='text-base text-title'>
          {version === 'international'
            ? t('app.docs.corekgQAPairUploadTemplate')
            : `模板：${version === 'custom' ? '' : 'corekg'}问答对上传模板.xlsx，所有允许导入的需求字段请参考模板。`}
        </span>
        <span className='text-base text-description whitespace-pre-wrap'>
          {version === 'international'
            ? t('app.docs.exampleColsQa')
            : `${version === 'custom' ? '' : 'corekg'}问答对上传模板.xlsx`}
        </span>
        <Link
          to={'/qa-template.xlsx'}
          target='_blank'
          download={`${version === 'custom' ? '' : 'corekg'}问答对上传模板.xlsx`}
        >
          <Button icon={<DownloadOutlined />} className='self-start'>
            {tC('button.download')}
          </Button>
        </Link>
      </div>
    </div>
  )
}

type QAList = {
  answer: string
  question: string
}[]
const useUploadFile = (forest_id: number) => {
  const { t: tM } = useTranslation('messages')
  const { message } = App.useApp()
  const { run, data, loading } = useRequest(
    async (file: File) => {
      const { qa_list, valid_lines } = (await uploadQAPair({
        forest_id,
        file,
      })) as any
      if (valid_lines === 0) {
        const msg = tM('extractQAPairFailFromFile')
        message.error(msg)
        throw new Error(msg)
      }
      return qa_list as QAList
    },
    { manual: true },
  )
  return {
    upload: run,
    qa_list: data,
    uploading: loading,
  }
}

const useCommitQAList = (forest_id: number, onSuccess?: () => void) => {
  const { loading, run } = useRequest(
    async (qa_list: QAList) => {
      await commitQAPair({
        forest_id,
        qa_list,
      })
      onSuccess?.()
    },
    { manual: true },
  )
  return {
    commit: run,
    commiting: loading,
  }
}
