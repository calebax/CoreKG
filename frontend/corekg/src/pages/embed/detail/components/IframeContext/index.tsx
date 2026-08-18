import { createContext, FC, PropsWithChildren, useContext } from 'react'

const IframeContext = createContext<{
  agentDetail?: any
} | null>(null)

export const IframeProvider: FC<PropsWithChildren & { agentDetail?: any }> = (
  props,
) => {
  const { agentDetail } = props

  return (
    <IframeContext.Provider value={{ agentDetail }}>
      {props.children}
    </IframeContext.Provider>
  )
}

export const useIframeContext = () => {
  const context = useContext(IframeContext)
  if (!context) {
    throw new Error('应当处于iframe context下')
  }
  return context
}
