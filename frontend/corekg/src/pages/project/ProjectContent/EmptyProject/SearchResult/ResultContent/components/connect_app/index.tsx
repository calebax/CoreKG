import { FC } from 'react'
import { useTranslation } from 'react-i18next'
import { spliteFileName } from '@/utils'
import { CommonResultItem } from '../../CommonResultItem'
import Image from '../Image'
import errorPlaceholderImage from './images/errorPlaceholder.svg'

const images = new Map(
  Object.entries(
    import.meta.glob('./images/*.svg', {
      import: 'default',
      eager: true,
    }),
  ).map(([k, v]) => {
    const { name } = spliteFileName(k.split('/').at(-1)!)
    return [name, v as string]
  }),
)
const ConnectApp: FC<{ value: any }> = (props) => {
  const { t } = useTranslation('pages')
  const { external_type, value } = props.value
  switch (external_type) {
    case 'gmail': {
      const { date, from, snippet, subject, webLink } = value
      return (
        <CommonResultItem
          icon={<img src={images.get(external_type)} className='w-6 h-6' />}
          creator={from}
          creatorAvatar=''
          time={date}
          title_html={subject}
          to={webLink}
        >
          <div className='text-[#3C4149]'>{snippet}</div>
        </CommonResultItem>
      )
    }
    case 'google_drive': {
      const {
        createdTime,
        owners,
        name,
        thumbnailLink,
        /* snippet, subject,*/ webLink,
      } = value
      return (
        <CommonResultItem
          icon={<img src={images.get(external_type)} className='w-6 h-6' />}
          creator={owners.join(' ')}
          creatorAvatar=''
          time={createdTime}
          title_html={name}
          to={webLink}
        >
          <div className='h-[200px]'>
            <Image src={thumbnailLink} errorSrc={errorPlaceholderImage} />
          </div>
        </CommonResultItem>
      )
    }
    case 'confluence': {
      const { title, excerpt, lastModified, resultGlobalContainerURL, id } =
        value
      const webLink = `https://calebax.atlassian.net/wiki${resultGlobalContainerURL}/pages/${id}`
      return (
        <CommonResultItem
          icon={<img src={images.get(external_type)} className='w-6 h-6' />}
          creator={''}
          creatorAvatar=''
          time={lastModified}
          title_html={title}
          to={webLink}
        >
          <div className='text-[#3C4149]'>{excerpt}</div>
        </CommonResultItem>
      )
    }
    case 'slack': {
      const { title, date, webLink } = value
      return (
        <CommonResultItem
          icon={<img src={images.get(external_type)} className='w-6 h-6' />}
          creator={''}
          creatorAvatar=''
          time={date}
          title_html={title}
          to={webLink}
        >
          <div className='text-[#0C99FF] underline'>
            {t('project.clickToJumpAndView')}
          </div>
        </CommonResultItem>
      )
    }
    default:
      return null
  }
}
export default ConnectApp
