import { useDeployConfig } from './useDeployConfig'

const DEFAULT_QA_INPUT_MAX_LENGTH = 500
const EXTENDED_QA_INPUT_MAX_LENGTH = 10000

export const useQaInputMaxLength = () => {
  const { version, qaInputMaxLength } = useDeployConfig()

  const isExtendedEnv =
    import.meta.env.DEV ||
    import.meta.env.MODE === 'test' ||
    version === 'custom'

  // 本地、测试、custom 环境至少放宽到 10000，与项目其它环境开关保持一致。
  if (isExtendedEnv) {
    return Math.max(qaInputMaxLength ?? 0, EXTENDED_QA_INPUT_MAX_LENGTH)
  }

  // 其他环境仍按部署配置生效，未配置时保持 500 的默认上限。
  return qaInputMaxLength ?? DEFAULT_QA_INPUT_MAX_LENGTH
}
