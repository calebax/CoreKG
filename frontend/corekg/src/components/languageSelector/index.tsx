import { Select, Space } from 'antd'
import { GlobalOutlined } from '@ant-design/icons'
import { useLocaleStore } from '@/stores/locale'
import { languages } from '@/utils/locale'

export const LanguageSelector: React.FC = () => {
  const { language, setLanguage, isChanging } = useLocaleStore()

  const options = languages.map((lang) => ({
    value: lang.code,
    label: (
      <Space>
        <span>{lang.flag}</span>
        <span>{lang.name}</span>
      </Space>
    ),
  }))

  return (
    <Select
      value={language}
      onChange={setLanguage}
      options={options}
      loading={isChanging}
      style={{ width: 140 }}
      suffixIcon={<GlobalOutlined />}
      placement='bottomRight'
    />
  )
}
