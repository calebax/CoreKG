import { convertToInteger } from '@/utils'

export type IDInfo = {
  forest_id: number
  forest_db_name: string
  forest_table_name: string
  type: 'error' | 'db' | 'table' | 'header'
}
/**
 * 从路径参数中获取知识库、db等的id\
 * 根据id判断当前所处的层级
 */
export const useIds = () => {
  const { id: forest_id_string, '*': rest = '' } = useParams()

  const { forest_id, forest_db_name, forest_table_name } = useMemo(() => {
    const [forest_db_name, forest_table_name] = rest.split('/')
    const forest_id = convertToInteger(forest_id_string) ?? 0
    return { forest_id, forest_db_name, forest_table_name }
  }, [forest_id_string, rest])

  const id_info = useMemo<IDInfo>(() => {
    const type: 'error' | 'db' | 'table' | 'header' = (() => {
      if (!forest_id) return 'error'
      if (!forest_db_name) return 'db'
      if (!forest_table_name) return 'table'
      return 'header'
    })()
    return { forest_id, forest_db_name, forest_table_name, type }
  }, [forest_id, forest_db_name, forest_table_name])

  return id_info
}
