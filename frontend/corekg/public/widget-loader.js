;(function () {
  'use strict'

  window.MyAIWidget = function (config) {
    const defaultConfig = {
      containerId: null,
      agentId: '',
      agentType: 'role',
      position: 'bottom-right',
      width: 380,
      height: 600,
      minimizable: false,
      sandbox: 'allow-same-origin allow-scripts allow-forms allow-popups',
    }

    const settings = { ...defaultConfig, ...config }

    // 构建iframe URL
    const baseUrl = settings.baseUrl || window.location.origin
    const iframeSrc = `${baseUrl}/iframe/widget/${settings.agentType}/${settings.agentId}`

    // 创建iframe
    const iframe = document.createElement('iframe')
    iframe.src = iframeSrc
    iframe.sandbox = settings.sandbox
    iframe.allow = 'microphone; camera'

    // 计算iframe位置
    const getIframePosition = (position, isOpen) => {
      const buttonSize = 80 // 60px button + 20px margin
      const panelWidth = settings.width
      const panelHeight = settings.height

      const positions = {
        'bottom-right': {
          bottom: '0px',
          right: '0px',
          top: 'auto',
          left: 'auto',
          width: isOpen ? `${panelWidth + 40}px` : `${buttonSize}px`,
          height: isOpen ? `${panelHeight + 40}px` : `${buttonSize}px`,
        },
        'bottom-left': {
          bottom: '0px',
          left: '0px',
          top: 'auto',
          right: 'auto',
          width: isOpen ? `${panelWidth + 40}px` : `${buttonSize}px`,
          height: isOpen ? `${panelHeight + 40}px` : `${buttonSize}px`,
        },
        'top-right': {
          top: '0px',
          right: '0px',
          bottom: 'auto',
          left: 'auto',
          width: isOpen ? `${panelWidth + 40}px` : `${buttonSize}px`,
          height: isOpen ? `${panelHeight + 40}px` : `${buttonSize}px`,
        },
        'top-left': {
          top: '0px',
          left: '0px',
          bottom: 'auto',
          right: 'auto',
          width: isOpen ? `${panelWidth + 40}px` : `${buttonSize}px`,
          height: isOpen ? `${panelHeight + 40}px` : `${buttonSize}px`,
        },
      }

      return positions[position] || positions['bottom-right']
    }

    // 如果指定了容器，内嵌模式
    if (settings.containerId) {
      const container = document.getElementById(settings.containerId)
      if (container) {
        iframe.style.cssText = 'width: 100%; height: 100%; border: none;'
        container.appendChild(iframe)
      } else {
        console.error('Container not found:', settings.containerId)
        return
      }
    } else {
      // 悬浮模式 - 初始状态只显示按钮大小
      const initialPosition = getIframePosition(settings.position, false)
      iframe.style.cssText = `
        position: fixed;
        border: none;
        overflow: hidden;
        background: transparent;
        pointer-events: none;
        transition: width 0.3s ease, height 0.3s ease;
        ${Object.entries(initialPosition)
          .map(([k, v]) => `${k}: ${v}`)
          .join('; ')}
      `
      document.body.appendChild(iframe)
    }

    // 与iframe通信
    let widgetReady = false
    let isWidgetOpen = false

    window.addEventListener('message', function (event) {
      if (event.source !== iframe.contentWindow) return

      if (event.data.type === 'WIDGET_READY') {
        widgetReady = true
        // 启用iframe的鼠标事件
        iframe.style.pointerEvents = 'auto'

        // 发送配置
        iframe.contentWindow.postMessage(
          {
            type: 'WIDGET_CONFIG',
            config: {
              position: settings.position,
              theme: settings.theme,
              primaryColor: settings.primaryColor,
              type: settings.agentType,
              width: settings.width,
              height: settings.height,
              title: settings.title,
              minimizable: settings.minimizable,
            },
          },
          '*',
        )
      }

      // 监听widget状态变化
      if (event.data.type === 'WIDGET_STATE_CHANGE') {
        isWidgetOpen = event.data.isOpen

        if (!settings.containerId) {
          // 根据widget状态调整iframe大小
          const newPosition = getIframePosition(settings.position, isWidgetOpen)
          Object.entries(newPosition).forEach(([key, value]) => {
            iframe.style[key] = value
          })
        }
      }
    })

    return {
      open: function () {
        if (widgetReady) {
          iframe.contentWindow.postMessage({ type: 'WIDGET_OPEN' }, '*')
        }
      },
      close: function () {
        if (widgetReady) {
          iframe.contentWindow.postMessage({ type: 'WIDGET_CLOSE' }, '*')
        }
      },
      destroy: function () {
        iframe.remove()
      },
      sendMessage: function (message) {
        if (widgetReady) {
          iframe.contentWindow.postMessage(
            {
              type: 'WIDGET_MESSAGE',
              message: message,
            },
            '*',
          )
        }
      },
    }
  }
})()
