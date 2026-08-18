import { useState, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import config from '@/config'
import {
  loadAccountBindingList,
  unBindAccount,
  getBindAccountState,
} from '@/api/accountBindings'
import {
  AccountBindItem,
  BaseBindingItem,
  BindableItem,
  EProvider,
} from './types'

export function useAccountBindings() {
  const loadedRef = useRef<boolean>(false)
  const navigate = useNavigate()

  const [loadingList, setLoadingList] = useState<string[]>([])

  const [bindableList, setBindableList] = useState<BindableItem[]>([])

  const [bindInfos, setBindInfos] = useState<Record<string, AccountBindItem>>(
    {},
  )

  const loadAccountList = async () => {
    try {
      const { bindings, supported } = await loadAccountBindingList()

      const bindingsInfos: Record<string, AccountBindItem> = {}

      bindings?.forEach((item: AccountBindItem) => {
        bindingsInfos[item.provider] = item
      })

      setBindInfos(bindingsInfos)

      setBindableList(
        (supported as BaseBindingItem[]).map((item) => ({
          ...item,
          name: item.provider.charAt(0).toUpperCase() + item.provider.slice(1),
        })),
      )
    } catch (error) {
      console.log(error)
    }
    loadedRef.current = true
  }
  const handleBind = async (provider: EProvider) => {
    try {
      setLoadingList((list) => {
        return [...list, provider]
      })
      const { state } = await getBindAccountState({
        provider,
        redirectUrl: location.href,
      })

      location.href = `${config.withPrefix('account.Connect')}/${provider}?state=${state}`
    } catch (error) {
      console.log(error)
      setLoadingList((list) => {
        return list.filter((item) => item !== provider)
      })
    }
  }

  const handleUnBind = async (id: number, provider: EProvider) => {
    try {
      setLoadingList((list) => {
        return [...list, provider]
      })
      await unBindAccount(id)

      setBindInfos((state) => {
        // eslint-disable-next-line @typescript-eslint/no-unused-vars
        const { [provider]: _, ...args } = state
        return args
      })
    } catch (error) {
      console.log(error)
    } finally {
      setLoadingList((list) => {
        return list.filter((item) => item !== provider)
      })
    }
  }

  const goBack = () => {
    navigate('/')
  }

  useEffect(() => {
    loadAccountList()
  }, [])

  return {
    bindableList,
    bindInfos,
    handleBind,
    handleUnBind,
    goBack,
    loadedRef,
    loadingList,
  }
}
