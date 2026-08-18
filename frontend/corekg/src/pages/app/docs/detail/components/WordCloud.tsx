import {
  useCallback,
  useEffect,
  useRef,
  useState,
  forwardRef,
  useImperativeHandle,
} from 'react'
import { Spin } from 'antd'
import WordCloud from 'react-d3-cloud'
import Moveable from 'react-moveable'
import { getKnowledgeBaseWordCloud } from '@/api/knowledge'
import './WordCloud.scss'

interface WordCloudData {
  text: string
  value: number
}

interface WordCloudProps {
  knowledgeBaseId?: string
  onWordClick?: (word: string) => void
  containerRef?: React.RefObject<HTMLDivElement> // 外部容器的ref，用于监听滚轮事件
}

export interface WordCloudRef {
  resetView: () => void
}

const WordCloudComponent = forwardRef<WordCloudRef, WordCloudProps>(
  ({ knowledgeBaseId, onWordClick: onWordClickProp, containerRef }, ref) => {
    const [data, setData] = useState<WordCloudData[]>([]) // 词云数据
    const [isLoading, setIsLoading] = useState(false) // 加载状态
    const [scale, setScale] = useState(1) // 缩放比例
    const [position, setPosition] = useState({ x: 0, y: 0 }) // transform 的 x 和 y
    const [isDragging, setIsDragging] = useState(false) // 是否正在拖动
    const [lastMousePosition, setLastMousePosition] = useState({ x: 0, y: 0 }) // 上次鼠标位置

    const wordCloudContainer = useRef<HTMLDivElement>(null) // 词云容器

    // 暴露重置视角方法给父组件
    useImperativeHandle(ref, () => ({
      resetView: () => {
        setScale(1)
        setPosition({ x: 0, y: 0 })
      },
    }))

    // 更新 transform 样式
    const updateTransform = useCallback(() => {
      const container = wordCloudContainer.current
      if (container) {
        container.style.transform = `translate(${position.x}px, ${position.y}px) scale(${scale})`
      }
    }, [position, scale])

    // 分段权重分配字体大小算法
    const fontSize = useCallback(
      (word: WordCloudData) => {
        const minFontSize = 16
        const midFontSize = 48
        const maxFontSize = 82

        const minWeight = data[data.length - 1]?.value || 0
        const maxWeight = data[0]?.value || 1
        const midWeight = (minWeight + maxWeight) / 2

        if (word.value < midWeight) {
          return (
            ((word.value - minWeight) / (midWeight - minWeight)) *
              (midFontSize - minFontSize) +
            minFontSize
          )
        } else {
          return (
            ((word.value - midWeight) / (maxWeight - midWeight)) *
              (maxFontSize - midFontSize) +
            midFontSize
          )
        }
      },
      [data],
    )

    // 旋转角度控制
    const rotate = useCallback((word: WordCloudData) => word.value % 1, [])

    // 点击词云词触发事件
    const onWordClick = useCallback(
      (event: any, word: WordCloudData) => {
        console.log(`onWordClick: ${word.text ? word.text : ''}`)
        // 如果有外部传入的点击回调，调用它
        if (onWordClickProp) {
          onWordClickProp(word.text)
        }
      },
      [onWordClickProp],
    )

    useEffect(() => {
      const fetchWordCloudData = async () => {
        if (!knowledgeBaseId) return

        setIsLoading(true)
        try {
          const response = await getKnowledgeBaseWordCloud(
            Number(knowledgeBaseId),
          )
          console.log('词云 API 响应:', response)

          if (
            response &&
            response.word_cloud &&
            Array.isArray(response.word_cloud)
          ) {
            const formattedData = response.word_cloud.map((item: any) => ({
              text: item.word,
              value: item.weight * 5,
            }))
            setData(formattedData)
          } else {
            console.log('接口返回数据为空，使用模拟数据')
            const mockData = [
              { word: '人工智能', weight: 76 },
              { word: '机器学习', weight: 39 },
              { word: '深度学习', weight: 38 },
              { word: '神经网络', weight: 33 },
              { word: '自然语言处理', weight: 28 },
              { word: '计算机视觉', weight: 26 },
              { word: '数据挖掘', weight: 25 },
              { word: '大数据', weight: 23 },
              { word: '云计算', weight: 20 },
              { word: '物联网', weight: 17 },
            ]
            setData(
              mockData.map((item) => ({
                text: item.word,
                value: item.weight * 5,
              })),
            )
          }
        } catch (error) {
          console.error('获取词云数据失败:', error)
          const mockData = [
            { word: '人工智能', weight: 76 },
            { word: '机器学习', weight: 39 },
            { word: '深度学习', weight: 38 },
            { word: '神经网络', weight: 33 },
            { word: '自然语言处理', weight: 28 },
            { word: '计算机视觉', weight: 26 },
            { word: '数据挖掘', weight: 25 },
            { word: '大数据', weight: 23 },
            { word: '云计算', weight: 20 },
            { word: '物联网', weight: 17 },
          ]
          setData(
            mockData.map((item) => ({
              text: item.word,
              value: item.weight * 5,
            })),
          )
        } finally {
          setIsLoading(false)
        }
      }

      fetchWordCloudData()
    }, [knowledgeBaseId])

    // 滚轮事件处理函数
    const handleWheel = useCallback((event: WheelEvent) => {
      event.preventDefault()
      setScale((prevScale) => {
        const newScale = event.deltaY > 0 ? prevScale * 0.9 : prevScale * 1.1
        return Math.min(Math.max(newScale, 0.3), 3)
      })
    }, [])

    // 鼠标按下事件处理函数
    const handleMouseDown = useCallback((event: MouseEvent) => {
      // 如果点击的是词云中的文字，不启动拖动
      const target = event.target as HTMLElement
      if (target.tagName === 'text') {
        return
      }

      setIsDragging(true)
      setLastMousePosition({ x: event.clientX, y: event.clientY })
      event.preventDefault()
    }, [])

    // 鼠标移动事件处理函数
    const handleMouseMove = useCallback(
      (event: MouseEvent) => {
        if (!isDragging) return

        const deltaX = event.clientX - lastMousePosition.x
        const deltaY = event.clientY - lastMousePosition.y

        setPosition((prev) => ({
          x: prev.x + deltaX,
          y: prev.y + deltaY,
        }))

        setLastMousePosition({ x: event.clientX, y: event.clientY })
      },
      [isDragging, lastMousePosition],
    )

    // 鼠标松开事件处理函数
    const handleMouseUp = useCallback(() => {
      setIsDragging(false)
    }, [])

    // 监听 scale 或 position 变化，更新 transform 样式
    useEffect(() => {
      updateTransform()
    }, [updateTransform])

    // 添加滚轮事件监听器 - 优先使用外部容器，否则使用内部容器
    useEffect(() => {
      const targetContainer =
        containerRef?.current || wordCloudContainer.current
      if (targetContainer) {
        targetContainer.addEventListener('wheel', handleWheel, {
          passive: false,
        })
        return () => {
          targetContainer.removeEventListener('wheel', handleWheel)
        }
      }
    }, [handleWheel, containerRef])

    // 添加鼠标拖动事件监听器
    useEffect(() => {
      const targetContainer =
        containerRef?.current || wordCloudContainer.current
      if (targetContainer) {
        targetContainer.addEventListener('mousedown', handleMouseDown)
        document.addEventListener('mousemove', handleMouseMove)
        document.addEventListener('mouseup', handleMouseUp)

        return () => {
          targetContainer.removeEventListener('mousedown', handleMouseDown)
          document.removeEventListener('mousemove', handleMouseMove)
          document.removeEventListener('mouseup', handleMouseUp)
        }
      }
    }, [handleMouseDown, handleMouseMove, handleMouseUp, containerRef])

    if (isLoading) {
      return (
        <div className='flex h-full w-full items-center justify-center'>
          <Spin size='large' />
        </div>
      )
    }

    return (
      <>
        <div
          id='wordCloudContainer'
          className={`transform overflow-hidden ${isDragging ? 'cursor-grabbing' : 'cursor-grab'}`}
          ref={wordCloudContainer}
        >
          <WordCloud
            data={data}
            font='Inter'
            fontWeight='normal'
            spiral='archimedean'
            fontSize={fontSize}
            rotate={rotate}
            padding={1}
            random={Math.random}
            onWordClick={onWordClick}
          />
        </div>

        <Moveable
          target={wordCloudContainer.current}
          draggable
          scalable
          keepRatio
          rotatable
          onDrag={({ beforeTranslate }) => {
            setPosition({ x: beforeTranslate[0], y: beforeTranslate[1] })
          }}
          onScale={({ target, scale: newScale }) => {
            setScale(newScale[0]) // 等比缩放
            updateTransform()
          }}
          onRotate={({ target, beforeRotate }) => {
            const container = wordCloudContainer.current
            if (container) {
              container.style.transform += ` rotate(${beforeRotate}deg)`
            }
          }}
          className='moveable-line'
        />
      </>
    )
  },
)

export default WordCloudComponent
