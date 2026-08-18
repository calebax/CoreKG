/* eslint-disable react-refresh/only-export-components */
import { FC, PropsWithChildren } from 'react'
import { ConfigProvider } from 'antd'
import styles from './styles.module.scss'

const AgentStyleProvider: FC<PropsWithChildren> = (props) => {
  return (
    <ConfigProvider
      theme={{
        components: {
          Checkbox: {
            colorPrimary: '#0C99FF',
            colorPrimaryHover: '#0C99FF',
            colorPrimaryBorder: '#0C99FF',
            controlInteractiveSize: 14,
          },
          Tree: {
            colorPrimary: '#0C99FF',
            colorPrimaryHover: '#0C99FF',
            colorPrimaryBorder: '#0C99FF',
            controlInteractiveSize: 14,
          },
          Button: {
            colorBgContainer: 'rgba(12,31,23,0.050980392156862744)',
            colorPrimary: '#0C99FF',
          },
          Slider: {
            colorPrimaryBorderHover: '#0C99FF',
            trackBg: '#0C99FF',
            handleActiveColor: '#0C99FF',
            dotActiveBorderColor: '#0C99FF',
            handleActiveOutlineColor: '#0C99FF',
            handleColor: '#0C99FF',
          },
          Radio: {
            colorPrimary: '#0C99FF',
            colorPrimaryHover: '#0C99FF',
          },
          Switch: {
            colorPrimary: '#0C99FF',
            colorPrimaryHover: '#0C99FF',
          },
          Input: {
            hoverBorderColor: '#0C99FF',
            activeBorderColor: '#0C99FF',
          },
        },
      }}
    >
      {props.children}
    </ConfigProvider>
  )
}
/**
 * @example
 * ```
 * const Comp: FC = withAgentStyle((props) => {
 *    reutrn <div className={props.extraClassName}></div>
 * })
 * ```
 */
export function withAgentStyle<T>(
  _Comp: FC<T & { extraClassName: string }>,
): FC<T> {
  const Comp: any = _Comp
  const CompWithStyle: any = forwardRef((props, ref) => {
    return (
      <AgentStyleProvider>
        <Comp ref={ref} {...props} extraClassName={styles.agent} />
      </AgentStyleProvider>
    )
  })
  return CompWithStyle
}
