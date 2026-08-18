import { useState, useCallback } from 'react'
import type { MouseEvent as ReactMouseEvent } from 'react'
import { useProject } from '..'

// 定义不同部分的枚举类型
export enum ESection {
  DRAWER = 'drawer',
  SESSION_HISTORY = 'sessionHistory',
}

export enum EDrawerType {
  GRAPH = 'graph',
  REFERENCE = 'reference',
  CHART = 'chart',
}
// 定义各部分宽度的限制
const SECTION_WIDTH_LIMITS = {
  [ESection.DRAWER]: {
    MAX: 600,
    MIN: 350,
  },
  [ESection.SESSION_HISTORY]: {
    MAX: 400,
    MIN: 260,
  },
  [EDrawerType.CHART]: {
    MAX: 600,
    MIN: 450,
  },
  [EDrawerType.GRAPH]: {
    MAX: 1200,
    MIN: 350,
  },
  [EDrawerType.REFERENCE]: {
    MAX: 600,
    MIN: 350,
  },
}

// 拖拽
const useDragResizeHandler = (key: ESection | EDrawerType) => {
  const [resizing, setResizing] = useState<boolean>(false)
  const [width, setWidth] = useState<number>(SECTION_WIDTH_LIMITS[key].MIN)
  const { MAX: maxWidth, MIN: minWidth } = SECTION_WIDTH_LIMITS[key]
  // 使用 useCallback 记忆化 onDragging 函数
  const onDragging = useCallback(
    (e: MouseEvent) => {
      setWidth((v) => {
        // 根据不同部分计算新的宽度
        const newWidth =
          key === ESection.SESSION_HISTORY ? v + e.movementX : v - e.movementX
        // 确保新宽度在最小和最大宽度限制范围内
        return Math.min(Math.max(newWidth, minWidth), maxWidth)
      })
    },
    [key, minWidth, maxWidth],
  )

  // 使用 useCallback 记忆化 onDragEnd 函数
  const onDragEnd = useCallback(() => {
    document.body.style.cursor = ''
    document.body.style.userSelect = ''
    setResizing(false)
    document.removeEventListener('mousemove', onDragging)
    document.removeEventListener('mouseup', onDragEnd)
  }, [onDragging])

  const handleFn = useCallback(
    (e: ReactMouseEvent) => {
      e.preventDefault()
      e.stopPropagation()

      document.body.style.cursor = 'col-resize'
      document.body.style.userSelect = 'none'
      setResizing(true)
      document.addEventListener('mousemove', onDragging)
      document.addEventListener('mouseup', onDragEnd)
    },
    [onDragging, onDragEnd],
  )

  return {
    width,
    resizing,
    handleFn,
    setWidth,
  }
}

// 自定义 Hook 用于处理项目内容的宽度管理
export function useProjectSection() {
  const {
    data: { sessions },
  } = useProject()

  const [sessionVisible, setSessionVisible] = useState<boolean>(
    !!sessions.length,
  )
  const [drawerVisible, setDrawerVisible] = useState<boolean>(false)
  const [drawerType, setDrawerType] = useState<EDrawerType>(EDrawerType.CHART)
  const [dialogIndex, setDialogIndex] = useState<number>(-1)
  // 获取处理会话历史部分宽度变化的函数
  const {
    width: sessionWidth,
    resizing: sessionResizing,
    handleFn: handleSessionWidthChange,
  } = useDragResizeHandler(ESection.SESSION_HISTORY)

  const handleOpenSession = () => {
    setSessionVisible(true)
  }

  const handleCloseSession = () => {
    setSessionVisible(false)
  }

  // 获取处理图表部分宽度变化的函数
  const {
    width: drawerWidth,
    resizing: drawerResizing,
    handleFn: handleDrawerWidthChange,
  } = useDragResizeHandler(drawerType)

  const handleOpenDrawer = (type: EDrawerType) => {
    setDrawerType(type)
    setDrawerVisible(true)
  }

  const handleCloseDrawer = () => {
    setDrawerVisible(false)
  }

  return {
    // 历史记录侧边栏
    sessionWidth,
    sessionResizing,
    sessionVisible,
    handleOpenSession,
    handleCloseSession,
    handleSessionWidthChange,

    // 画布及搜索资源
    drawerWidth,
    drawerResizing,
    drawerVisible,
    drawerType,

    handleOpenDrawer,
    handleCloseDrawer,
    handleDrawerWidthChange,

    // 当前活跃的dialog
    dialogIndex,
    setDialogIndex,
  }
}
