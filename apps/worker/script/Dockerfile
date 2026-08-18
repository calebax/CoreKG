FROM node:22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache bash git

RUN corepack enable && corepack prepare pnpm@latest --activate

COPY apps/worker/package.json apps/worker/pnpm-lock.yaml apps/worker/pnpm-workspace.yaml apps/worker/tsconfig.base.json ./
COPY apps/worker/packages/ ./packages/
COPY apps/worker/apps/ ./apps/

RUN pnpm install --frozen-lockfile
RUN pnpm build

FROM node:22-alpine

WORKDIR /app


COPY --from=builder /app/apps/worker/dist ./dist
COPY --from=builder /app/apps/worker/node_modules ./node_modules
COPY --from=builder /app/apps/worker/package.json ./

ENTRYPOINT ["node", "/app/dist/index.js"]
