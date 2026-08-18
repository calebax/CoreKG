/* eslint-disable react-refresh/only-export-components */
import { FC, PropsWithChildren } from 'react'
import { ConfigProvider } from 'antd'
import styles from './styles.module.scss'

const CheckboxStyleProvider: FC<PropsWithChildren> = (props) => {
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
        },
      }}
    >
      {props.children}
    </ConfigProvider>
  )
}

export function withCheckboxStyle<T>(
  _Comp: FC<T & { checkboxClassName?: string }>,
): FC<T> {
  const Comp: any = _Comp
  const CompWithCheckboxStyle: any = forwardRef((props: any, ref) => {
    return (
      <CheckboxStyleProvider>
        <Comp ref={ref} {...props} checkboxClassName={styles.checkbox} />
      </CheckboxStyleProvider>
    )
  })
  return CompWithCheckboxStyle as FC<T>
}
