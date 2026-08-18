import { useState } from 'react'
import { useMount } from 'ahooks'
import { listEmployee } from '@/api'
import useDataStore from '@/stores/data'

// 定义API返回的用户数据类型
interface EmployeeResponse {
  uin: number
  user_name: string
}

export default function useUserData() {
  const { userList, setUserList } = useDataStore()
  const [loading, setLoading] = useState(false)

  const loadUserList = async () => {
    try {
      setLoading(true)
      const res = await listEmployee({ listAll: true })
      const list: EmployeeResponse[] = res.Data || []
      setUserList(
        list.map((x: EmployeeResponse) => ({
          value: x.uin,
          label: x.user_name,
        })),
      )
      return list
    } catch {
      return []
    } finally {
      setLoading(false)
    }
  }

  const refreshUserList = async () => {
    await loadUserList()
  }

  useMount(() => {
    if (userList.length === 0) {
      loadUserList()
    }
  })

  return {
    userList,
    loading,
    loadUserList,
    refreshUserList,
  }
}
