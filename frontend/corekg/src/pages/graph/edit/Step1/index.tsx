import { Button, Empty, Skeleton } from 'antd'
import { GraphTemplate } from 'Graph'
import { useRequest, useSize } from 'ahooks'
import { cn } from '@/utils'
import { submitTemplate } from '@/api/graph'
import { Templates, EmptyTemplate } from '@/pages/graph/templates'
import type { StepComponent } from '..'
import { useGraphInfo } from '../../GraphProvider'
import { SpinWithMask } from '../../components/SpinWithMask'
import BtnIcon from './images/btnIcon.svg?react'
import BtnIconActive from './images/btnIconActive.svg?react'
import Logo from './images/logo.svg?react'
import styles from './styles.module.scss'

export const Step1: StepComponent = (props) => {
  const { increase, decrease } = props
  const { data, loading, reloadTag } = useGraphInfo()

  const { run: selectTemplate, loading: submitting } = useRequest(
    async (template: GraphTemplate) => {
      await submitTemplate({ graph_id: data!.id, template })
      reloadTag()
      increase()
    },
    { manual: true },
  )
  // 有tags就不选择模板
  useEffect(() => {
    if (data?.tags.length !== 0) {
      increase()
    }
  }, [data?.tags.length, increase])

  const templateDom = useRef<HTMLDivElement>(null)
  const { width } = useSize(templateDom) ?? {}
  if (loading || data?.tags.length !== 0) return <Skeleton active />
  if (!Templates) return <Empty />
  return (
    <div className='h-full overflow-hidden flex flex-col gap-4 relative'>
      <SpinWithMask show={submitting} />
      <span className='flex items-center gap-2'>
        <Logo />
        <span className='text-xl'>请选择模板</span>
        <span className='text-xs text-[#616373]'>
          以下是系统提供的行业预设模板
        </span>
      </span>
      <div className='flex-1 overflow-auto flex flex-col pt-2'>
        <div
          className='grid gap-5 p-2'
          style={{
            gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))',
          }}
        >
          <Template
            className='self-start'
            {...EmptyTemplate}
            style={{ width }}
            next={() => selectTemplate(EmptyTemplate)}
          ></Template>
          {Templates.map((t) => {
            return (
              <Template
                ref={templateDom}
                key={t.name}
                {...t}
                next={() => selectTemplate(t)}
              ></Template>
            )
          })}
        </div>
      </div>
      <span className='flex items-center self-end gap-4'>
        <Button onClick={decrease}>上一步</Button>
      </span>
    </div>
  )
}

const Template = forwardRef<
  HTMLDivElement,
  Style & GraphTemplate & { next: () => void }
>((props, ref) => {
  const { avatar, description, name, next, className, style } = props
  return (
    <div
      ref={ref}
      className={cn(
        'rounded-xl bg-[#F9F8FF]',
        'p-4 h-full flex flex-col gap-2.5',
        className,
        styles.template,
      )}
      style={style}
    >
      {/* <img
        src={avatar}
        className='rounded-xl flex-1 overflow-hidden'
        style={{ height: 156, minHeight: 156 }}
      /> */}
      <span className='font-semibold text-lg'>{name}</span>
      <span className='text-[#616373] text-sm flex-1'>{description}</span>
      <div
        className={cn(
          'cursor-pointer rounded-full border border-[#653ec4]',
          'w-full mt-1 py-1 flex items-center justify-center gap-2.5',
          styles.nextBtn,
        )}
        onClick={next}
      >
        <BtnIcon className={styles.icon} />
        <BtnIconActive className={styles.activeIcon} />
        立刻使用
      </div>
    </div>
  )
})
