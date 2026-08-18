import React, { createContext, useContext, useState, ReactNode } from 'react'

interface TabsContextType {
  activeIndex: number
  setActiveIndex: (index: number) => void
}

const TabsContext = createContext<TabsContextType | undefined>(undefined)

interface TabsProps {
  children: ReactNode
  defaultActiveIndex?: number
}

interface TabListProps {
  children: ReactNode
}

interface TabProps {
  children: ReactNode
  index: number
}

interface TabPanelProps {
  children: ReactNode
  index: number
}

const Tabs = ({ children, defaultActiveIndex = 0 }: TabsProps) => {
  const [activeIndex, setActiveIndex] = useState(defaultActiveIndex)

  return (
    <TabsContext.Provider value={{ activeIndex, setActiveIndex }}>
      <div className='w-full h-full flex flex-col'>{children}</div>
    </TabsContext.Provider>
  )
}

const TabList = ({ children }: TabListProps) => {
  return <div className='flex border-b border-gray-200 mb-4'>{children}</div>
}

const Tab = ({ children, index }: TabProps) => {
  const context = useContext(TabsContext)
  if (!context) {
    throw new Error('Tab must be used within Tabs')
  }

  const { activeIndex, setActiveIndex } = context
  const isActive = activeIndex === index

  return (
    <button
      className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${isActive ? 'text-blue-600 border-blue-600' : 'text-gray-500 border-transparent hover:text-gray-700 hover:border-gray-300'}`}
      onClick={() => setActiveIndex(index)}
    >
      {children}
    </button>
  )
}

const TabPanel = ({ children, index }: TabPanelProps) => {
  const context = useContext(TabsContext)
  if (!context) {
    throw new Error('TabPanel must be used within Tabs')
  }

  const { activeIndex } = context

  if (activeIndex !== index) {
    return null
  }

  return <div className='flex-1 h-full'>{children}</div>
}

Tabs.TabList = TabList
Tabs.Tab = Tab
Tabs.TabPanel = TabPanel

export default Tabs
