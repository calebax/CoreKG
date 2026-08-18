import { FC } from 'react'
import { App, Popover } from 'antd'
import { useTranslation } from 'react-i18next'
import { useImmer } from 'use-immer'
import { cn } from '@/utils'
import PictureIcon2 from '../../images/picture2.svg'
import PictureIcon from '../../images/picture.svg'
import { CustomDialogInput } from '../CustomDialogInput'
import { SendButton } from '../SendButton'
import { ImageInfo, ImageLoader, ImageUploader } from './ImageUploader'
import { LinkRes } from './LinkRes'
import { SearchType } from './searchType'

export const SearchMode: FC<{ hidden?: boolean }> = (props) => {
  const { hidden } = props

  const { t } = useTranslation('pages')
  const { t: tC } = useTranslation('common')
  const [text, setText] = useState('')

  const [img, setImg] = useImmer<ImageInfo>({})

  const { message } = App.useApp()
  const btnActive = text.trim() || img.url
  const navigate = useNavigate()
  const onSearch = (type: SearchType = 'all') => {
    const search = { text, img: img.url, type }
    const searchParams = new URLSearchParams()
    searchParams.append('search', encodeURIComponent(JSON.stringify(search)))
    navigate(`/search?${searchParams.toString()}`)
  }
  const [open, setOpen] = useState<boolean>(true)
  const inputRef = useRef<HTMLDivElement>()
  useEffect(() => {
    const onClickAway = (e: MouseEvent) => {
      const target = e.target as any
      const container = inputRef.current
      if (!container) return
      if (container.contains(target)) return
      setOpen(false)
    }
    document.addEventListener('click', onClickAway, { capture: true })
    return () => {
      document.removeEventListener('click', onClickAway)
    }
  }, [])
  return (
    <Popover
      arrow={false}
      open={open}
      placement='bottom'
      content={<LinkRes img={img?.url} text={text} startSearch={onSearch} />}
    >
      <CustomDialogInput
        ref={inputRef}
        className={cn({ hidden })}
        value={text}
        onChange={setText}
        onSubmit={() => onSearch()}
        onClick={() => setOpen(true)}
        mode='search'
        prefix={
          img.file ? (
            <ImageUploader
              image={img}
              onChange={(val) => {
                setImg(val)
                setOpen(true)
              }}
            />
          ) : null
        }
      >
        {/* 左侧图片上传图标 */}
        <div className='flex items-center gap-2'>
          {img.file ? (
            <div className='w-6 h-6 relative cursor-not-allowed'>
              <img
                src={PictureIcon2}
                alt={tC('button.upload', { target: tC('file.image') })}
                className='w-full h-full'
              />
            </div>
          ) : (
            <div className='w-6 h-6 relative cursor-pointer'>
              <img
                src={PictureIcon}
                alt={tC('button.upload', { target: tC('file.image') })}
                className='w-full h-full'
                onClick={() => {
                  // 创建一个隐藏的input元素来触发文件选择
                  const input = document.createElement('input')
                  input.type = 'file'
                  input.accept = 'image/*'
                  input.onchange = (e) => {
                    const file = (e.target as HTMLInputElement).files?.[0]
                    if (file) {
                      const maxSize = 5 * 1024 * 1024
                      if (file.size > maxSize) {
                        message.error('文件大小不能超过5M')
                        return
                      }
                      setImg({
                        file,
                        url: URL.createObjectURL(file),
                        name: file.name,
                      })
                      setOpen(true)
                    }
                  }
                  input.click()
                }}
              />
            </div>
          )}
        </div>

        {/* 右侧发送按钮 */}
        <div className='flex items-center'>
          <SendButton
            active={!!btnActive}
            onClick={() => {
              if (btnActive) {
                onSearch()
              } else {
                message.warning(t('app.home.inputTextOrUploadImage'))
              }
            }}
          />
        </div>
      </CustomDialogInput>
    </Popover>
  )
}
