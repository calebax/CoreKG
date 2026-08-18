/** 模型基本信息 */
export type ModelBaseInfo = {
  show_name: string
  model_provider: 'aliyun' | 'deepseek'
  /** 基础模型 */
  model_name: string
  /** 模型地址 */
  model_url: string
}

/** 模型展示信息 */
export type ModelShowInfo = ModelBaseInfo & {
  ID: number
  head_url: string
  /** 是否系统预置 */
  public_type?: 'system'
  /** 创建者 */
  user_name: string
}
