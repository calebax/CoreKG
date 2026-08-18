import { useState, useRef, useEffect, FC } from 'react'
import { useClickAway, useHover } from 'ahooks'
import { motion, useDragControls, useMotionValue } from 'motion/react'
import { useDeployConfig } from '@/utils/useDeployConfig'
import IframeLogo from './iframe.svg?react'

export const Agent: FC = () => {
  const { version } = useDeployConfig()
  const [isIframeOpen, setIsIframeOpen] = useState(false)

  const containerRef = useRef<HTMLDivElement>(null)
  const y = useMotionValue(0)
  const dragControls = useDragControls()
  const isDragging = useRef(false)

  const buttonRef = useRef(null)
  const iframeContainerRef = useRef(null)
  const isHovering = useHover(buttonRef)

  useClickAway(() => {
    setIsIframeOpen(false)
    console.log('click away')
  }, [iframeContainerRef, buttonRef])

  useEffect(() => {
    if (isIframeOpen) {
      buttonRef.current = null
    }
  }, [isIframeOpen])

  if (window.location.pathname.startsWith('/iframe')) return null
  if (version === 'custom') return null
  return (
    <>
      {!isIframeOpen && (
        <div className='fixed top-25 bottom-25 right-0 z-50' ref={containerRef}>
          <motion.div
            className='absolute bottom-0 right-0 cursor-pointer'
            dragControls={dragControls}
            drag='y'
            dragMomentum={false}
            dragElastic={0.5}
            dragConstraints={containerRef}
            whileDrag={{
              cursor: 'grabbing',
              opacity: 0.5,
            }}
            style={{ y }}
            onDragStart={() => {
              isDragging.current = true
            }}
            onDragEnd={() => {
              isDragging.current = false
            }}
          >
            <button
              ref={buttonRef}
              onClick={() => {
                if (!isDragging.current) {
                  setIsIframeOpen(true)
                }
              }}
              className=' cursor-[inherit] relative group animate-fadeIn translate-x-10 hover:translate-x-0 transition-transform duration-300 ease-in-out'
              aria-label='Open assistant'
            >
              <IframeLogo className='w-24 h-24 text-white' />

              {isHovering && !isIframeOpen && (
                <div className='absolute bottom-16 -left-16 whitespace-nowrap bg-gradient-to-br from-[rgb(222,239,255)] to-[rgb(232,238,255)] text-black px-2 py-1.5 rounded-md text-sm pointer-events-none animate-fadeIn'>
                  点我帮忙哦！
                </div>
              )}
            </button>
          </motion.div>
        </div>
      )}

      {isIframeOpen && (
        <>
          <div
            ref={iframeContainerRef}
            className='fixed bottom-6 right-6 w-[500px] h-[700px] bg-white rounded-lg shadow-2xl z-50 animate-slideUp overflow-hidden'
          >
            <div className='absolute top-0 left-0 right-0 h-12 flex items-center justify-between px-4 z-10'>
              <button
                onClick={() => setIsIframeOpen(false)}
                className='text-black hover:bg-[#00000033] rounded p-1 transition-colors'
                aria-label='Close'
              >
                <svg
                  className='w-5 h-5'
                  fill='none'
                  stroke='currentColor'
                  viewBox='0 0 24 24'
                >
                  <path
                    strokeLinecap='round'
                    strokeLinejoin='round'
                    strokeWidth={2}
                    d='M6 18L18 6M6 6l12 12'
                  />
                </svg>
              </button>
            </div>

            {/* Iframe */}
            <iframe
              src='https://app.corekg.com/iframe/detail/role/HjW3zoe'
              className='w-full h-full'
              title='AI Assistant'
            />
          </div>
        </>
      )}
    </>
  )
}
