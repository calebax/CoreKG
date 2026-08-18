import React, { useState, useRef, useEffect } from 'react'
import { Image, Spin, Button } from 'antd'
import {
  ZoomInOutlined,
  ZoomOutOutlined,
  RotateLeftOutlined,
  RotateRightOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  FullscreenOutlined,
  SoundOutlined,
  SoundFilled,
  ReloadOutlined,
} from '@ant-design/icons'

interface MediaViewerProps {
  file: string
  isVideo?: boolean
  location?: number[] | null
  locationKey?: string
}

const MediaViewer: React.FC<MediaViewerProps> = ({
  file,
  isVideo = false,
  location,
  locationKey,
}) => {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // 视频相关状态
  const [isPlaying, setIsPlaying] = useState(false)
  const [isMuted, setIsMuted] = useState(false)
  const [duration, setDuration] = useState(0)
  const [currentTime, setCurrentTime] = useState(0)

  // 视频元素引用
  const videoRef = useRef<HTMLVideoElement>(null)
  const prevLocationKeyRef = useRef<string | undefined>(undefined)
  const [retryCount, setRetryCount] = useState(0)

  const safeUrl = file

  const detectIsVideo = () => {
    if (isVideo) return true
    if (!file) return false

    const fileExt = file.split('?')[0].split('.').pop()?.toLowerCase()
    return fileExt === 'mp4' || fileExt === 'webm' || fileExt === 'ogg'
  }

  const fileIsVideo = detectIsVideo()

  // 图片处理函数
  const handleImageLoad = () => {
    setLoading(false)
  }

  const handleImageError = () => {
    setLoading(false)
    setError('无法加载图片，请检查图片格式或网络连接')
    console.error('图片加载失败:', file)
  }

  // 视频处理函数
  const handleVideoLoadStart = () => {
    console.log('视频开始加载:', file)
    setLoading(true)
  }

  const handleVideoLoadedData = () => {
    console.log('视频数据已加载')
    setLoading(false)
    if (videoRef.current) {
      setDuration(videoRef.current.duration)
    }
  }

  // 监听 location 变化并跳转到指定时间位置
  useEffect(() => {
    if (
      !fileIsVideo ||
      !location?.length ||
      !videoRef.current ||
      loading ||
      prevLocationKeyRef.current === locationKey
    ) {
      return
    }

    prevLocationKeyRef.current = locationKey

    // location[1] 表示视频时间（秒）
    const timeInSeconds = location[1]
    if (typeof timeInSeconds !== 'number' || !Number.isFinite(timeInSeconds)) {
      return
    }

    const timer = setTimeout(() => {
      const video = videoRef.current
      if (!video?.duration) return

      const targetTime = Math.max(0, Math.min(timeInSeconds, video.duration))
      video.currentTime = targetTime
      setCurrentTime(targetTime)
    }, 300)

    return () => clearTimeout(timer)
  }, [location, locationKey, fileIsVideo, loading])

  const handleVideoCanPlay = () => {
    setLoading(false)
  }

  const handleVideoError = (
    e: React.SyntheticEvent<HTMLVideoElement, Event>,
  ) => {
    const target = e.target as HTMLVideoElement
    const errorCode = target.error?.code || 'unknown'
    console.error('视频加载失败:', {
      error: errorCode,
      message: target.error?.message || 'Unknown error',
      src: target.src,
    })
    setLoading(false)
    setError(`无法加载视频 (错误代码: ${errorCode})`)
  }

  const handleVideoTimeUpdate = () => {
    if (videoRef.current) {
      setCurrentTime(videoRef.current.currentTime)
    }
  }

  const handleVideoEnded = () => {
    setIsPlaying(false)
    if (videoRef.current) {
      videoRef.current.currentTime = 0
    }
  }

  // 视频控制函数
  const togglePlay = () => {
    if (videoRef.current) {
      if (isPlaying) {
        videoRef.current.pause()
      } else {
        videoRef.current.play().catch((err) => {
          console.error('播放失败:', err)
          // 尝试静音播放 (解决一些浏览器的自动播放限制)
          if (!isMuted) {
            setIsMuted(true)
            videoRef.current!.muted = true
            videoRef
              .current!.play()
              .catch((e) => console.error('静音播放仍然失败:', e))
          }
        })
      }
      setIsPlaying(!isPlaying)
    }
  }

  const toggleMute = () => {
    if (videoRef.current) {
      videoRef.current.muted = !isMuted
      setIsMuted(!isMuted)
    }
  }

  const handleFullScreen = () => {
    if (videoRef.current) {
      if (videoRef.current.requestFullscreen) {
        videoRef.current.requestFullscreen().catch((err) => {
          console.error('全屏模式失败:', err)
        })
      }
    }
  }

  const formatTime = (timeInSeconds: number) => {
    if (isNaN(timeInSeconds)) return '00:00'
    const minutes = Math.floor(timeInSeconds / 60)
    const seconds = Math.floor(timeInSeconds % 60)
    return `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`
  }

  // 跳转到指定时间点
  const seekTo = (e: React.ChangeEvent<HTMLInputElement>) => {
    const seekTime = parseFloat(e.target.value)
    if (videoRef.current) {
      videoRef.current.currentTime = seekTime
      setCurrentTime(seekTime)
    }
  }

  const retryLoading = () => {
    setError(null)
    setLoading(true)
    setRetryCount((prev) => prev + 1)

    if (fileIsVideo && videoRef.current) {
      videoRef.current.load()
    }
  }

  // 渲染视频播放器
  const renderVideoPlayer = () => {
    return (
      <div className='w-full h-full flex flex-col'>
        {/* 视频播放区域 */}
        <div className='relative flex-1 flex items-center justify-center bg-black'>
          <video
            ref={videoRef}
            key={`video-${retryCount}`} // 重试时强制重新创建视频元素
            src={safeUrl}
            className='max-h-full max-w-full'
            onClick={togglePlay}
            onLoadStart={handleVideoLoadStart}
            onLoadedData={handleVideoLoadedData}
            onCanPlay={handleVideoCanPlay}
            onError={handleVideoError}
            onTimeUpdate={handleVideoTimeUpdate}
            onEnded={handleVideoEnded}
            onPlay={() => setIsPlaying(true)}
            onPause={() => setIsPlaying(false)}
            controls={false}
            playsInline
            controlsList='nodownload' // 禁用下载按钮
            crossOrigin='anonymous' // 启用跨域请求
            preload='auto' // 预加载视频
          />

          {!isPlaying && !loading && !error && (
            <div className='absolute inset-0 flex items-center justify-center bg-transparent'>
              {/* <PlayCircleOutlined 
                onClick={togglePlay}
                className="text-white text-5xl opacity-80 hover:opacity-100 cursor-pointer drop-shadow-lg"
              /> */}
            </div>
          )}
        </div>

        {/* 视频控制栏 */}
        {!error && (
          <div className='bg-gray-100 p-2 flex items-center space-x-3'>
            {/* 播放/暂停按钮 */}
            <Button
              type='text'
              icon={
                isPlaying ? <PauseCircleOutlined /> : <PlayCircleOutlined />
              }
              onClick={togglePlay}
            />

            {/* 进度条 */}
            <div className='flex-1 flex items-center space-x-2'>
              <span className='text-xs'>{formatTime(currentTime)}</span>
              <input
                type='range'
                className='flex-1 h-1'
                min={0}
                max={duration || 100}
                step='any'
                value={currentTime}
                onChange={seekTo}
              />
              <span className='text-xs'>{formatTime(duration)}</span>
            </div>

            {/* 音量控制 */}
            <Button
              type='text'
              icon={isMuted ? <SoundOutlined /> : <SoundFilled />}
              onClick={toggleMute}
            />

            {/* 全屏按钮 */}
            <Button
              type='text'
              icon={<FullscreenOutlined />}
              onClick={handleFullScreen}
            />
          </div>
        )}

        {error && (
          <div className='bg-gray-100 p-4 text-center'>
            <p className='text-red-500 mb-2'>{error}</p>
            <Button
              type='primary'
              icon={<ReloadOutlined />}
              onClick={retryLoading}
            >
              重试加载
            </Button>
          </div>
        )}
      </div>
    )
  }

  // 渲染图片查看器
  const renderImageViewer = () => {
    return (
      <Image
        src={safeUrl}
        className='max-h-full max-w-full object-contain'
        onLoad={handleImageLoad}
        onError={handleImageError}
        preview={{
          toolbarRender: (
            _,
            {
              transform: { scale },
              actions: { onZoomIn, onZoomOut, onRotateLeft, onRotateRight },
            },
          ) => (
            <div className='flex space-x-2'>
              <ZoomOutOutlined onClick={onZoomOut} />
              <ZoomInOutlined onClick={onZoomIn} />
              <RotateLeftOutlined onClick={onRotateLeft} />
              <RotateRightOutlined onClick={onRotateRight} />
              <span>缩放: {Math.round(scale * 100)}%</span>
            </div>
          ),
        }}
      />
    )
  }

  return (
    <div className='flex items-center justify-center'>
      {loading && (
        <div className='absolute inset-0 flex items-center justify-center bg-white bg-opacity-75 z-10'>
          <Spin size='large'>
            <div className='mt-3'>
              {fileIsVideo ? '视频加载中...' : '图片加载中...'}
            </div>
          </Spin>
        </div>
      )}

      {fileIsVideo ? renderVideoPlayer() : renderImageViewer()}
    </div>
  )
}

export default MediaViewer
