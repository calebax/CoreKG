export enum EProvider {
  GMAIL = 'gmail',
  SLACK = 'slack',
  DRIVE = 'googleDrive',
  CONFLUENCE = 'confluence',
}

export type BaseBindingItem = {
  provider: EProvider
  logo: string
}
export type AccountBindItem = {
  account: string
  id: number
  provider: EProvider
}

export type BindableItem = BaseBindingItem & { name: string }
