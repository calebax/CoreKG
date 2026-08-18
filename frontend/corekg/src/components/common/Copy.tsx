import { message } from 'antd'
import { CopyOutlined, CheckOutlined } from '@ant-design/icons'
import { CopyToClipboard } from 'react-copy-to-clipboard'

export interface CopyProps {
  text: string | number
  className?: string
}

export default function Copy({ text, className }: CopyProps) {
  const [isCopied, setIsCopied] = useState(false)

  const handleCopy = () => {
    message.success('Copied')
    setIsCopied(true)
    setTimeout(() => {
      setIsCopied(false)
    }, 1000)
  }

  return (
    <CopyToClipboard text={String(text)} onCopy={handleCopy}>
      <button className={`flex items-center gap-1 cursor-pointer ${className}`}>
        {isCopied ? <CheckOutlined /> : <CopyOutlined />}
        <span>Copy</span>
      </button>
    </CopyToClipboard>
  )
}
