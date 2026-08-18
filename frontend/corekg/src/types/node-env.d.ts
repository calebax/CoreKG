// 声明node:url模块
declare module 'node:url' {
  export function fileURLToPath(url: URL | string): string
  export class URL {
    constructor(input: string, base?: string | URL)
    toString(): string
    static createObjectURL(object: Blob | MediaSource): string
    static revokeObjectURL(url: string): void
  }
}

// 扩展ImportMeta接口
interface ImportMeta {
  url: string
  env: Record<string, string>
}
