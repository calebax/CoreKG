import { Dispatch, FC, SetStateAction } from 'react'
import { Badge, Button, Image, Input } from 'antd'
import { CloseCircleOutlined, PictureOutlined } from '@ant-design/icons'
import { useRequest } from 'ahooks'
import { cn } from '@/utils'
import { uploadImage } from '@/api/common'
import { loadFile } from '@/utils/loadFile'
import SearchButtonActive from './SearchButtonActive.svg?react'
import SearchButtonDisabled from './SearchButtonDisabled.svg?react'

export type ComplexSearchValue = {
  text?: string
  img?: string
}
export type ComplexSearchProps = {
  value: ComplexSearchValue
  onChange: Dispatch<SetStateAction<ComplexSearchValue>>
  hidden?: boolean
  onSearch?: (value: ComplexSearchValue) => void
}
/** 复杂搜索 */
export const ComplexSearch: FC<ComplexSearchProps> = (props) => {
  const { hidden, onSearch, value, onChange } = props
  const { text, img } = value
  const setText = useCallback(
    (text: string) => {
      onChange((prev) => ({
        ...prev,
        text,
      }))
    },
    [onChange],
  )
  const setImg = useCallback(
    (img: string) => {
      onChange((prev) => ({
        ...prev,
        img,
      }))
    },
    [onChange],
  )
  return (
    <div className={cn('flex flex-col p-4', { hidden })}>
      <div
        className={cn(
          'flex flex-col mb-4 p-1',
          'resize-y min-h-20 overflow-hidden',
          'border border-gray-400 rounded-xl',
        )}
      >
        <Input.TextArea
          value={text}
          onChange={(e) => setText(e.target.value)}
          maxLength={150}
          className={cn(
            'flex-1 ',
            'resize-none overflow-auto',
            'border-none shadow-none',
          )}
        ></Input.TextArea>
        <span className='self-end flex items-center gap-2'>
          <span>{text?.length ?? 0}/150</span>
          <SearchButton text={text} img={img} onSearch={onSearch} />
        </span>
      </div>
      <span className='flex justify-between'>
        <ImageUploader value={img} onChange={setImg} />
        <Button type='text' onClick={() => setText('')}>
          清空
        </Button>
      </span>
    </div>
  )
}

type ImageUploaderProps = {
  value?: string
  onChange?: (val: string) => void
}
const ImageUploader: FC<ImageUploaderProps> = (props) => {
  const { value, onChange } = props
  const legalExs = ['.png', '.jpg', '.jpeg']
  const { loading, run: upload } = useRequest(
    async (file: File) => {
      const { url } = (await uploadImage(
        { file, purpose: 'yg-chat' },
        {
          timeout: 0,
          headers: { 'Content-Type': 'multipart/form-data' },
        },
      )) as any
      onChange?.(url)
    },
    { manual: true },
  )
  if (loading) {
    return (
      <Button
        type='text'
        iconPosition='end'
        loading
        className='pointer-events-none'
      >
        上传中...
      </Button>
    )
  }
  if (!value) {
    return (
      <PictureOutlined
        className='text-xl'
        title='上传图片'
        onClick={() =>
          loadFile((fileList) => upload(fileList[0]), {
            accept: legalExs.join(','),
          })
        }
      />
    )
  }
  return (
    <Badge count={<CloseCircleOutlined onClick={() => onChange?.('')} />}>
      <Image className='w-[2.5em] h-[2.5em]' src={value}></Image>
    </Badge>
  )
}

type SearchButtonProps = ComplexSearchValue & {
  className?: string
  onSearch?: (value: ComplexSearchValue) => void
}
const SearchButton: FC<SearchButtonProps> = (props) => {
  const { className, text, img, onSearch } = props
  const disabled = Boolean(!text && !img)
  if (disabled) {
    return <SearchButtonDisabled className={className} />
  }
  return (
    <SearchButtonActive
      className={cn('cursor-pointer', className)}
      onClick={() => onSearch?.({ text, img })}
    />
  )
}
