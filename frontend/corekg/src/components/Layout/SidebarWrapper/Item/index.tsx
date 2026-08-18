import type {
  MouseEventHandler,
  CSSProperties,
  KeyboardEvent,
  ChangeEvent,
  ReactNode,
  ReactElement,
} from 'react'
import { useRef, useState, cloneElement, isValidElement } from 'react'
import { Input, Popover, Tooltip } from 'antd'
import { cn } from '@/utils'
import MoreIcon from '../images/more.svg?react'
import { EExpandableStatus } from '../types'
import styles from './index.module.scss'

type MenuItem = {
  text: string
  key: string | number
  icon: ReactNode
  popoverContent?: ReactNode
  [key: string]: any
}

interface IItemProps<T extends MenuItem> {
  style?: CSSProperties
  status?: EExpandableStatus
  text: string
  icon: ReactNode
  path?: string
  active?: boolean
  onClick?: MouseEventHandler<HTMLDivElement>
  menuList?: T[]
  isEdit?: boolean
  onEditClose?: (isChange: boolean, newText: string) => void
  onMenuClick?: (item: T) => void
  [key: string]: any
}

export default function Item<T extends MenuItem = MenuItem>(
  props: IItemProps<T>,
) {
  const {
    style,
    status,
    text,
    icon,
    path,
    active,
    onClick,
    menuList,
    isEdit,
    onEditClose,
    onMenuClick,
    className: incomingClassName,
    ...rest
  } = props

  const inputRef = useRef<any>()
  const [open, setOpen] = useState<boolean>(false)
  const [editText, setEditText] = useState<string>(props.text)

  if (status === EExpandableStatus.FOLD) {
    if (!icon) {
      // 如果没有图标，在收起状态下不显示任何内容
      return null
    }
    return (
      <Tooltip title={text} placement='right' arrow color='#1E1F28' {...rest}>
        <div
          onClick={onClick}
          className={cn(
            styles.item,
            styles.itemFold,
            { [styles.itemActive]: active },
            incomingClassName,
          )}
        >
          <span className={styles.itemFoldIconWrapper}>{icon}</span>
        </div>
      </Tooltip>
    )
  }

  const handleMenuClick = (item: T) => {
    onMenuClick?.(item)
    setOpen(false)
  }

  const renderMenuList = () => {
    return (
      <div className={styles.itemMenuList}>
        {menuList!.map((item: MenuItem) => {
          const menuItemContent = (
            <Item<T>
              text={item.text}
              key={item.key}
              icon={item.icon}
              onClick={(e) => {
                e.stopPropagation()
                if (!item.popoverContent) {
                  handleMenuClick(item as T)
                }
              }}
            />
          )

          if (item.popoverContent) {
            // 对于有 popoverContent 的菜单项，克隆 popoverContent 并替换其 children
            // 这样点击菜单项时就能触发弹窗打开
            const popoverElement = isValidElement(item.popoverContent)
              ? cloneElement(item.popoverContent as ReactElement, {
                  children: menuItemContent,
                })
              : item.popoverContent

            return (
              <div key={item.key} onClick={(e) => e.stopPropagation()}>
                {popoverElement}
              </div>
            )
          }

          return menuItemContent
        })}
      </div>
    )
  }
  const handleOpenChange = (newOpen: boolean) => {
    setOpen(newOpen)
  }

  const renderMore = () => {
    if (!(menuList && menuList.length)) {
      return null
    }
    return (
      <Popover
        trigger='click'
        arrow={false}
        placement='rightTop'
        content={renderMenuList}
        destroyOnHidden
        open={open}
        onOpenChange={handleOpenChange}
      >
        {/* 用原生元素包一层，避免 Popover 对 SVG 子组件使用 findDOMNode（React 严格模式告警） */}
        <span
          className={styles.itemMenu}
          onClick={(e) => e.stopPropagation()}
        >
          <MoreIcon className={styles.itemMenuIcon} />
        </span>
      </Popover>
    )
  }

  const handleBlur = () => {
    const newText = editText.trim()
    const isChange = newText !== text && newText !== ''
    onEditClose?.(isChange, newText)
  }

  const handleInputChange = (e: ChangeEvent<HTMLInputElement>) => {
    setEditText(e.target.value)
  }
  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      inputRef.current.blur()
    }
  }

  return (
    <div
      onClick={onClick}
      style={style}
      {...rest}
      className={cn(
        styles.item,
        { [styles.itemActive]: active },
        incomingClassName,
      )}
    >
      {icon && <div>{icon}</div>}
      {!isEdit ? (
        // 不用 Typography ellipsis+tooltip（antd 内部会走 findDOMNode）；单行省略 + 原生 title 展示全文
        <div className={styles.itemText} title={text}>
          <span className={styles.itemTextEllipsis}>{text}</span>
        </div>
      ) : (
        <Input
          ref={inputRef}
          className={styles.itemInput}
          autoFocus
          onBlur={handleBlur}
          onKeyDown={handleKeyDown}
          onChange={handleInputChange}
          value={editText}
        />
      )}
      {!isEdit && renderMore()}
    </div>
  )
}
