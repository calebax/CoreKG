import { createContext, FC, PropsWithChildren } from 'react'
import { useRequest } from 'ahooks'
import { listEmployee } from '@/api'

const EmployeeContext = createContext<any[] | undefined>(undefined)
export const EmployeeProvider: FC<PropsWithChildren> = (props) => {
  const { data: managers } = useRequest(async () => {
    const res: any = await listEmployee({ listAll: true })
    const data: any[] = res.Data ?? []
    return data.map((item) => {
      return {
        uin: item.uin as string,
        name: item.user_name as string,
      }
    })
  })
  return (
    <EmployeeContext.Provider value={managers}>
      {props.children}
    </EmployeeContext.Provider>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export const useEmployee = () => useContext(EmployeeContext)
