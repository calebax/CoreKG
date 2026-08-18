import { BasicAgentInfo } from 'Agent'
import { match, P } from 'ts-pattern'
import prompt from './images/prompt.svg'
import role_play from './images/role_play.svg'
import workflow from './images/workflow.svg'

export const getDefaultAvatar = (type: BasicAgentInfo['type']) => {
  return match(type)
    .with('prompt', () => prompt)
    .with(P.union('role_play', 'knowledge'), () => role_play)
    .with('workflow', () => workflow)
    .exhaustive()
}
