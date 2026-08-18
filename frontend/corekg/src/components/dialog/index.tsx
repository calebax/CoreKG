import { AIDialog } from './AIDialog'
import { UserDialog } from './UserDialog'

export { AIDialog } from './AIDialog'
export { UserDialog, type Attachment } from './UserDialog'
export { AttachmentList } from './AttachmentList'
/** 问答数组 */
// eslint-disable-next-line @typescript-eslint/no-empty-object-type
export type DialogList<T extends object = {}> = ((AIDialog | UserDialog) & T)[]
