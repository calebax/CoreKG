type ImageProps = {
  src: string
  errorSrc: string
}

export default function Image(props: ImageProps) {
  const [url, setUrl] = useState<string>(props.src)

  const handleError = () => setUrl(props.errorSrc)

  useEffect(() => {
    setUrl(props.src)
  }, [props.src, props.errorSrc])

  return (
    <img
      src={url}
      className='h-[100%] w-[auto] border-[1px] border-[#F7F7F7] rounded-[16px] object-cover'
      onError={handleError}
    />
  )
}
