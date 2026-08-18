import { FC, ReactNode, memo, useState, useEffect, useMemo } from 'react'
import { ConfigProvider, Empty, Input, Skeleton, Tree } from 'antd'
import { PersonnelType, usePersonnelData } from 'Personnel'
import { useControllableValue } from 'ahooks'
import { useTranslation } from 'react-i18next'
import { match, P } from 'ts-pattern'
import { cn } from '@/utils'
import SearchIcon from './images/search.svg?react'
import styles from './styles.module.css'
import { filterTreeBySearch, getTreeNodeKey, useTreeMap } from './utils'

export type PersonnelTreeNode = {
  key: string
  parentKey?: string
  children?: PersonnelTreeNode[]

  id: string | number
  name: string

  title?: ReactNode
  loading?: boolean
  icon?: ReactNode
  type?: 'department' | 'employee'
  checkable?: boolean
  selectable?: boolean
  sort?: number

  [key: string]: any
}
export type PersonnelTree = {
  /** 参考antd文档 */
  checkStrictly?: boolean
  /** 允许选中节点 需要传入id并指明选中节点的类型  */
  selectable?: boolean
  selectedIds?: { id: any; type: PersonnelType }[]
  onSelect?: (
    selectedIds: PersonnelTree['selectedIds'],
    node: PersonnelTreeNode,
    type: 'select' | 'unselected',
  ) => void

  /** 允许勾选节点 需要传入id并指明选中节点的类型 */
  checkable?: boolean
  checkedIds?: { id: any; type: PersonnelType }[]
  onCheck?: (
    checkedIds: PersonnelTree['checkedIds'],
    node: PersonnelTreeNode,
    type: 'check' | 'uncheck',
  ) => void

  /** 禁止选中/勾选的节点 */
  disabledIds?: { id: any; type: PersonnelType }[]

  /** 展示编辑、重命名、上下移等部门节点操作 */
  showDepartmentOperators?: boolean

  /** 是否展示员工节点 */
  showEmployees?: boolean

  /** 隐藏搜索框 */
  hideSearch?: boolean
  placeholder?: string

  /** 搜索框样式类名 */
  searchClassName?: string

  /** 搜索框样式对象 */
  searchStyle?: React.CSSProperties

  /** 外部传入的搜索值 */
  externalSearch?: string

  /** 只有员工的部门才能被选中（用于权限管理等场景） */
  onlyDepartmentsWithEmployees?: boolean

  /** 样式类名 */
  className?: string

  /** 样式对象 */
  style?: React.CSSProperties
}
/**
 * 以树形式展示当前公司的部门信息\
 * 数据源来自`import {usePersonnelData} from 'Personnel'`\
 * 这是全局的人事数据信息\
 * 节点的关联通过key进行 和id有所不同.
 * */
export const PersonnelTree: FC<PersonnelTree> = memo((props) => {
  const { t } = useTranslation('common')
  const {
    checkStrictly,
    selectable,
    checkable,
    showDepartmentOperators,
    showEmployees,
    hideSearch,
    placeholder,
    searchClassName,
    searchStyle,
    externalSearch,
    onlyDepartmentsWithEmployees,
    disabledIds,
    className,
    style,
  } = props
  const [selectedIds, onSelect] = useControllableValue<
    PersonnelTree['selectedIds']
  >(props, {
    valuePropName: 'selectedIds',
    trigger: 'onSelect',
  })
  const [checkedIds, onCheck] = useControllableValue<
    PersonnelTree['checkedIds']
  >(props, {
    valuePropName: 'checkedIds',
    trigger: 'onCheck',
  })
  const { data, loadData } = usePersonnelData()
  const [internalSearch, setInternalSearch] = useState('')
  const [expandedKeys, setExpandedKeys] = useState<string[]>([])
  // 使用外部搜索值或内部搜索值
  const search = externalSearch !== undefined ? externalSearch : internalSearch
  const { treeMap, editingKey } = useTreeMap({
    showDepartmentOperators,
    showEmployees,
    search,
    setExpandedKeys,
    onlyDepartmentsWithEmployees,
    disabledIds,
  })
  const [selectedKeys, checkedKeys] = useMemo(() => {
    if (!treeMap) return []
    const getKeysByIds = (ids?: PersonnelTree['selectedIds']) => {
      const keys: string[] = []
      ids?.forEach(({ type, id }) => {
        switch (type) {
          case 'department':
            keys.push(getTreeNodeKey('department', { departmentId: id }))
            break
          case 'employee': {
            const employee = data.employee?.find((item) => item.id === id)
            employee?.departmentIds?.forEach((departmentId) => {
              keys.push(
                getTreeNodeKey('employee', {
                  departmentId,
                  employeeId: employee.id!,
                }),
              )
            })
            break
          }
        }
      })
      return keys
    }

    return [getKeysByIds(selectedIds), getKeysByIds(checkedIds)]
  }, [checkedIds, data.employee, selectedIds, treeMap])

  const treeData = useMemo(() => {
    if (!treeMap) return null
    return filterTreeBySearch(
      [...treeMap.values()].filter((item) => !item.parentKey),
      { search, preservedKeys: [editingKey] },
    )
  }, [editingKey, search, treeMap])

  useEffect(() => {
    if (treeData) return
    loadData(showEmployees ? 'employee' : 'department')
  }, [loadData, showEmployees, treeData])

  const first = useRef(true)
  useEffect(() => {
    if (treeData && treeData.length && first.current) {
      first.current = false
      setExpandedKeys(treeData.map((item) => item.key))
    }
  }, [treeData])

  const getIdsFromNodes = (nodes: PersonnelTreeNode[]) => {
    const departmentIds = new Set<any>()
    const employeeIds = new Set<any>()
    nodes.forEach((n) => {
      switch (n.type) {
        case 'department':
          departmentIds.add(n.id)
          break
        case 'employee':
          employeeIds.add(n.id)
          break
      }
    })
    const ids: NonNullable<PersonnelTree['selectedIds']> = []
    departmentIds.forEach((id) => {
      ids.push({
        id,
        type: 'department',
      })
    })
    employeeIds.forEach((id) => {
      ids.push({
        id,
        type: 'employee',
      })
    })
    return ids
  }
  return (
    <ConfigProvider
      theme={{
        components: {
          Tree: {
            colorPrimary: '#0C99FF',
            colorPrimaryHover: '#0C99FF',
            colorPrimaryBorder: '#0C99FF',
            controlInteractiveSize: 14,
            nodeSelectedBg: 'rgba(0,0,0,0.04)',
          },
        },
      }}
    >
      <div
        className={cn('flex flex-col gap-4 overflow-hidden', className)}
        style={style}
      >
        {!hideSearch && (
          <Input
            value={search}
            onChange={(e) => setInternalSearch(e.target.value)}
            className={cn(searchClassName, styles.input)}
            style={searchStyle}
            prefix={<SearchIcon />}
            placeholder={placeholder}
          />
        )}
        {match(treeData)
          .with(P.not(P.nonNullable), () => (
            <Skeleton active className={cn('p-4', className)} style={style} />
          ))
          .when(
            (v) => v.length === 0,
            () => (
              <Empty
                description={t('empty.noData')}
                className={cn('m-4', className)}
                style={style}
              />
            ),
          )
          .otherwise((treeData) => (
            <Tree
              className={cn(styles.tree, 'overflow-auto')}
              showIcon
              defaultExpandedKeys={[treeData[0].key]}
              checkStrictly={checkStrictly}
              treeData={treeData}
              expandedKeys={expandedKeys}
              onExpand={(v) => setExpandedKeys(v as any)}
              selectable={selectable}
              selectedKeys={selectedKeys}
              onSelect={(_, e) => {
                const { node, selected, selectedNodes } = e
                onSelect?.(
                  getIdsFromNodes(selectedNodes),
                  node,
                  selected ? 'select' : 'unselected',
                )
              }}
              checkable={checkable}
              checkedKeys={checkedKeys}
              onCheck={(_, info) => {
                const { checked, checkedNodes, node } = info
                onCheck?.(
                  getIdsFromNodes(checkedNodes),
                  node,
                  checked ? 'check' : 'uncheck',
                )
              }}
            ></Tree>
          ))}
      </div>
    </ConfigProvider>
  )
})
