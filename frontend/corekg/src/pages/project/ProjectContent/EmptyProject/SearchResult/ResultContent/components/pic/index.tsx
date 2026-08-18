import { FC } from 'react'

const PicItem: FC<{ value?: any }> = (props) => {
  const { value } = props
  const { id, forest_id, highlighted_file_name, highlights } = value
  const { image_url, location } = highlights[0]
  // console.log(location)

  const url = useMemo(() => {
    const searchParams = new URLSearchParams()

    searchParams.append(
      'location',
      encodeURIComponent(JSON.stringify(location)),
    )
    return `/docs/detail/${forest_id}/file/${id}?${searchParams.toString()}`
  }, [forest_id, id, location])

  return (
    <Link className='flex flex-col gap-4' to={url} target='_blank'>
      <img
        src={image_url}
        className='rounded w-53 h-16 object-contain object-center'
      />
      <div
        className='text-[#3C4149] line-clamp-3 text-ellipsis h-16'
        dangerouslySetInnerHTML={{ __html: highlighted_file_name }}
      ></div>
    </Link>
  )
}
export default PicItem
