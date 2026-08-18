# AGENTS.md — frontend/workflow

Coze Studio 前端项目，Rush + PNPM monorepo，React 18 + TypeScript + Rsbuild。

## 工具链与版本

- Node.js >= 21，PNPM 8.15.8，Rush 5.147.1
- 依赖管理用 `rush install` / `rush update`，**不要用 `pnpm install`**
- `rush.json` 不在本目录中（由 CI/Dockerfile 从上层仓库根目录提供）；本地开发需确保其存在
- 包内执行脚本用 `rushx`（如 `rushx dev`、`rushx test`、`rushx build`）

## 常用命令

| 操作 | 命令 |
|------|------|
| 启动开发服务器 | `cd apps/coze-studio && rushx dev` |
| 构建应用 | `rush build --to @coze-studio/app`（根目录）或 `rushx build`（应用目录） |
| 运行测试 | `rushx test`（vitest --run --passWithNoTests） |
| 测试覆盖率 | `rushx test:cov` |
| Lint | `rushx lint` |

构建工具为 **Rsbuild**（非 webpack/vite），配置文件 `apps/coze-studio/rsbuild.config.ts`。
开发代理目标：`https://example.com`。

## 测试

- 测试框架：**Vitest**（Jest 已禁用）
- 测试环境：**happy-dom**（jsdom 已禁用）
- 配置通过共享包 `@coze-arch/vitest-config`，preset 可选 `web`、`node`、`default`

## Lint 与格式化

- ESLint 9 flat config（`eslint.config.js`），使用 `@coze-arch/eslint-config`（preset: `web` 或 `node`）
- 每个包**必须**有 `eslint.config.js`（`rushx-config.json` audit 规则强制）
- Prettier 集成在 eslint-config 内；Stylelint 用于 CSS/Less

## 禁用库（disallowed_3rd_libraries.json）

| 禁用 | 替代 |
|------|------|
| jest | vitest |
| jsdom | happy-dom |
| husky / lint-staged | — |
| pdfjs-dist | @coze-arch/pdfjs-shadow |
| @flow-web/md-box | @coze-arch/bot-md-box-adapter |
| inquirer | @inquirer/prompts |

## 架构与包约定

- 唯一应用：`apps/coze-studio`（`@coze-studio/app`）
- 包作用域：`@coze-workflow/*`、`@coze-arch/*`、`@coze-agent-ide/*`、`@coze-studio/*`、`@coze-foundation/*`、`@coze-project-ide/*`
- 工作区依赖使用 `workspace:*` 协议
- 多数包直接导出 `src/index.ts`，无预构建步骤（`"build": "exit 0"`）
- 状态管理：**Zustand**（非 Valtio/Jotai）
- DI：inversify + legacy decorators（rsbuild 中配置）
- 路由：React Router v6
- UI 组件：`@coze-arch/coze-design`（基于 Semi Design），Tailwind CSS 3.3

## 关键目录

- `config/` — 共享配置（eslint、rsbuild、vitest、ts、tailwind、postcss、stylelint）
- `infra/` — 构建插件、eslint plugin、IDL 工具、工具库
- `packages/workflow/` — 工作流引擎（fabric-canvas、nodes、sdk、playground）
- `packages/arch/` — 架构层（bot-api、bot-hooks、i18n、bot-flags、web-context）
- `packages/agent-ide/` — Agent IDE 组件
- `packages/foundation/` — 账户、全局、空间、布局基础设施

## TypeScript

- TypeScript ~5.8.2，项目引用模式（`tsconfig.json` → `tsconfig.build.json` + `tsconfig.misc.json`）
- 共享基础配置：`@coze-arch/ts-config`
