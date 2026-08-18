import { Input } from 'antd'
import { SearchOutlined } from '@ant-design/icons'
import { Node } from '../graphData'

const SearchPanel = ({
  onSearch,
  nodes,
}: {
  onSearch: (nodeId: string | null) => void
  nodes: Node[]
}) => {
  const [searchValue, setSearchValue] = useState<string>('')
  const [showDropdown, setShowDropdown] = useState<boolean>(false)
  const [filteredNodes, setFilteredNodes] = useState<Node[]>([])

  useEffect(() => {
    if (searchValue.trim()) {
      const filtered = nodes
        .filter((node) =>
          (node.label || node.id)
            .toLowerCase()
            .includes(searchValue.toLowerCase()),
        )
        .slice(0, 10) // 限制结果数量
      setFilteredNodes(filtered)
      setShowDropdown(filtered.length > 0)
    } else {
      setFilteredNodes([])
      setShowDropdown(false)
    }
  }, [searchValue, nodes])

  return (
    <div className='absolute top-4 left-4 z-10 w-[250px]'>
      <div className='relative'>
        <Input
          placeholder='搜索节点...'
          value={searchValue}
          onChange={(e) => setSearchValue(e.target.value)}
          prefix={<SearchOutlined />}
          allowClear
          onFocus={() => setShowDropdown(filteredNodes.length > 0)}
          className='shadow-md'
        />
        {showDropdown && (
          <div className='absolute top-full left-0 right-0 mt-1 bg-white border border-gray-200 rounded shadow-lg max-h-[300px] overflow-y-auto z-20'>
            {filteredNodes.map((node) => (
              <div
                key={node.id}
                className='px-3 py-2 hover:bg-gray-100 cursor-pointer text-sm'
                onClick={() => {
                  onSearch(node.id)
                  setSearchValue('')
                  setShowDropdown(false)
                }}
              >
                <div className='font-medium'>{node.label || node.id}</div>
                {node.tag?.[0]?.description && (
                  <div className='text-xs text-gray-500 truncate'>
                    {node.tag[0].description}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

export default SearchPanel
