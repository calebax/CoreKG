import { Input } from 'antd'
import { SearchOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import SearchIcon from '../../images/search.svg?react'
import styles from '../styles.module.scss'

interface SearchInputProps {
  onSearch: (value: string) => void
}

export default function SearchInput({ onSearch }: SearchInputProps) {
  const { t } = useTranslation('common')
  const [searchValue, setSearchValue] = useState<string>('')
  return (
    <div>
      <Input
        value={searchValue}
        onChange={(e) => setSearchValue(e.target.value)}
        prefix={<SearchIcon />}
        placeholder={t('button.search')}
        onBlur={() => !searchValue?.trim?.() && onSearch(searchValue)}
        onPressEnter={() => onSearch(searchValue)}
        className={`w-[70px] h-[30px] border-[#0C99FF] shadow-none  ${styles.searchInputWrap} ${searchValue?.trim?.() ? styles.searchInputWrapSearching : ''}`}
      />
    </div>
  )
}
