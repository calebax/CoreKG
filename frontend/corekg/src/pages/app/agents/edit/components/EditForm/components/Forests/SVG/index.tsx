import React, { FC, useMemo } from 'react'
import { Agent } from 'Agent'
import CAD from './images/cad.svg?react'
import DB from './images/db.svg?react'
import FILE from './images/file.svg?react'
import QA from './images/qa.svg?react'

export const SVG: FC<
  Style & {
    type: NonNullable<Agent['forests']>[number]['forest_type']
  }
> = (props) => {
  const { type } = props
  const Comp = useMemo(() => {
    switch (type) {
      case 'file':
        return FILE
      case 'qa':
        return QA
      case 'cad':
        return CAD
      case 'data':
        return DB
      default:
        return React.Fragment
    }
  }, [type])
  return <Comp className={props.className} style={props.style} />
}
