const inputInstance = document.createElement('input')
inputInstance.type = 'file'

type Config = {
  accept?: string
  multiple?: boolean
  directory?: boolean
}
/** 函数式文件加载 */
export function loadFile(
  onLoadFile?: (fileList: FileList) => void,
  config: Config = {},
) {
  const { accept, multiple, directory } = config
  inputInstance.value = ''
  inputInstance.files = null
  inputInstance.accept = accept || ''
  inputInstance.multiple = multiple || false
  inputInstance.webkitdirectory = directory || false
  inputInstance.onchange = () => {
    const fileList = inputInstance.files
    if (!fileList || fileList.length === 0) return
    onLoadFile?.(fileList)
  }
  inputInstance.click()
}

/** 描述一个文件夹内所有的文件和子文件夹 */
export type Directory = Map<string, File | Directory>
export const convertFileListToDirectory = (
  fileList: Iterable<File> | ArrayLike<File>,
): Directory => {
  const directory: Directory = new Map()
  Array.from(fileList).forEach((file) => {
    const paths = file.webkitRelativePath.split('/').slice(0, -1)
    const insertFile = (currentDir: Directory, paths: string[]) => {
      if (paths.length === 0) {
        currentDir.set(file.name, file)
        return
      }
      const currentPath = paths[0]
      if (!currentDir.get(currentPath)) currentDir.set(currentPath, new Map())
      // 文件名和文件夹名不能相同 直接指定类型
      insertFile(currentDir.get(currentPath) as Directory, paths.slice(1))
    }
    insertFile(directory, paths)
  })
  return directory
}
