import React, { FC, Fragment, useState, useMemo } from 'react'
import { Empty } from 'antd'
import { useTranslation } from 'react-i18next'
import { ResultType } from '..'
import { TypeBtn } from '../TypeBtn'

export const ResultContent: FC<{
  result: { type: ResultType; values: any[] }[]
}> = (props) => {
  const { result } = props
  const { t } = useTranslation('common')
  const [type, setType] = useState<ResultType | 'all'>('all')
  const resultTotal = useMemo(() => {
    let count = 0
    result.forEach((item) => {
      count += item.values.length
    })
    return count
  }, [result])

  const currentResult = useMemo(() => {
    if (type === 'all') return result
    return [result.find((r) => r.type === type)!]
  }, [result, type])
  if (resultTotal === 0) {
    return <Empty description={t('empty.noData')} className='p-4 mx-auto' />
  }

  return (
    <div className='w-full h-full overflow-auto flex flex-col gap-4 rounded-xl'>
      <div className='flex-none flex gap-2.5 overflow-hidden overflow-x-auto p-2'>
        <TypeBtn
          type='all'
          onClick={() => setType('all')}
          active={type === 'all'}
          total={resultTotal}
        />
        {result.map((r) => {
          return (
            <TypeBtn
              type={r.type}
              onClick={() => setType(r.type)}
              active={type === r.type}
              total={r.values.length}
            />
          )
        })}
      </div>
      {currentResult.map((item) => {
        const { type, values } = item
        switch (type) {
          case 'forest':
          case 'doc':
          case 'agent':
          case 'connect_app': {
            return (
              <Fragment key={type}>
                {values.map((v) => (
                  <ResultItem type={type} value={v} key={v.id} />
                ))}
              </Fragment>
            )
          }
          case 'pic':
          case 'video': {
            return (
              <div className='ml-13 grid grid-cols-3 gap-2.5' key={type}>
                {values.map((v) => (
                  <ResultItem type={type} value={v} key={v.id} />
                ))}
              </div>
            )
          }
          case 'table':
            return null
        }
      })}
    </div>
  )
}
const Comps = new Map(
  Object.entries(
    import.meta.glob('./components/*/index.tsx', {
      import: 'default',
      eager: true,
    }),
  ).map(([k, v]) => {
    const type = k.split('/')[2]
    return [type, v]
  }),
) as Map<ResultType, FC<{ value?: any }>>

const ResultItem: FC<{ type: ResultType; value: any }> = (props) => {
  const { type, value } = props
  const Comp = Comps.get(type)
  return Comp ? <Comp value={value} /> : null
}
