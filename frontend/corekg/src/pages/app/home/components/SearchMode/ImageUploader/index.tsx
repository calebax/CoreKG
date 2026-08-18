import { FC, useState } from 'react'
import { Button, Divider, Image, Typography, Tooltip, Spin } from 'antd'
import { useRequest } from 'ahooks'
import { useTranslation } from 'react-i18next'
import { Updater } from 'use-immer'
import { cn } from '@/utils'
import { uploadImage } from '@/api/common'
import { loadFile } from '@/utils/loadFile'
import CloseIcon from '../../../images/close.svg?react'
import Pic from '../../../images/pic.svg?react'

export type ImageInfo = {
  file?: File
  url?: string
  name?: string
}

type ImageLoader = {
  className?: string
  onChange: (val: ImageInfo) => void
}
export const ImageLoader: FC<ImageLoader> = (props) => {
  const { onChange, className } = props
  const legalExs = ['.png', '.jpg', '.jpeg']

  return (
    <Pic
      className={cn('text-xl cursor-pointer', className)}
      onClick={() =>
        loadFile(
          (fileList) => {
            const file = fileList[0]
            onChange({ file, name: file.name })
          },
          {
            accept: legalExs.join(','),
          },
        )
      }
    />
  )
}

type ImageUploader = {
  className?: string
  image: ImageInfo
  onChange: Updater<ImageInfo>
}
export const ImageUploader: FC<ImageUploader> = (props) => {
  const { className, image, onChange } = props
  const [isHovered, setIsHovered] = useState(false)
  const { t: tC } = useTranslation('common')
  const { data: url, loading } = useRequest(async () => {
    const response = (await uploadImage(
      { file: image.file!, purpose: 'yg-chat' },
      {
        timeout: 0,
        headers: { 'Content-Type': 'multipart/form-data' },
      },
    )) as unknown as { url: string }
    const imageUrl = response.url
    onChange((draft) => {
      draft.url = imageUrl
    })
    return imageUrl
  })

  return (
    <div
      className='relative inline-block'
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      <Button
        type='text'
        className={cn(
          'h-11.5 bg-[#F0F2F7] rounded-lg p-2 flex items-center gap-1 max-w-60',
          className,
        )}
      >
        <div className='w-8 h-8 rounded flex-shrink-0 overflow-hidden relative'>
          {loading ? (
            <div className='w-full h-full flex items-center justify-center'>
              <Spin size='small' />
            </div>
          ) : (
            <Image
              className={cn('w-full h-full object-cover')}
              src={url || image.url}
              style={{ minWidth: '32px', minHeight: '32px' }}
            />
          )}
        </div>
        <Tooltip title={image.name} placement='top'>
          <Typography.Text className='text-sm text-[#0A1A3A] font-normal truncate fontFamily-PingFangSC'>
            {image.name}
          </Typography.Text>
        </Tooltip>
        {loading ? (
          <>
            <Divider type='vertical' />
            {tC('status.uploading')}
          </>
        ) : null}
      </Button>
      {isHovered && (
        <CloseIcon
          onClick={() => onChange({})}
          className='absolute -top-1 -right-1 cursor-pointer w-3 h-3'
        />
      )}
    </div>
  )
}
