import React from 'react'
import { App } from 'antd'
import DeleteWarningIcon from '@/assets/icons/delete-warning.svg?react'

interface ConfirmDeleteOptions {
  title?: string
  content?: string
  onOk: () => void | Promise<void>
  onCancel?: () => void
}

export default function useConfirm() {
  const { modal } = App.useApp()
  const confirmDelete = ({
    title = '确认删除？',
    content = '您确定要删除吗？',
    onOk,
    onCancel,
  }: ConfirmDeleteOptions) => {
    modal.confirm({
      icon: null,
      centered: true,
      className: 'delete-confirm-modal',
      content: React.createElement('div', {}, [
        React.createElement(
          'div',
          {
            key: 'header',
            className: 'flex justify-between',
          },
          [
            React.createElement(
              'div',
              {
                key: 'title',
                className: 'text-xl font-semibold',
              },
              title,
            ),
            React.createElement(DeleteWarningIcon, {
              key: 'icon',
              className: 'w-6 h-6',
            }),
          ],
        ),
        React.createElement(
          'div',
          {
            key: 'content',
            className: 'mt-10 text-base text-gray-500 mb-10',
          },
          content,
        ),
      ]),
      okText: '确认删除',
      okButtonProps: {
        danger: false,
      },
      cancelText: '取消',
      onOk: async () => {
        try {
          await onOk()
        } catch (error) {
          console.error('删除操作失败:', error)
        }
      },
      onCancel,
    })
  }

  return {
    confirmDelete,
  }
}
