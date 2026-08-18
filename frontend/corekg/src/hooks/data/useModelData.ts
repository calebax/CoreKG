import { useMount } from 'ahooks'
import { listModel } from '@/api'
import useDataStore from '@/stores/data'

export default function useModelData() {
  const { modelList, setModelList } = useDataStore()

  const loadModelList = async () => {
    try {
      const res = await listModel({ listAll: true })
      const list = res.Data || []
      setModelList(
        list.map((x) => ({
          value: x.ID,
          label: x.show_name,
          highAvailable: x.high_available,
        })),
      )
      return list
    } catch (error) {
      console.log('error', error)
      return []
    }
  }

  const refreshModelList = async () => {
    await loadModelList()
  }

  useMount(() => {
    if (modelList.length === 0) {
      loadModelList()
    }
  })

  return {
    modelList,
    loadModelList,
    refreshModelList,
  }
}
