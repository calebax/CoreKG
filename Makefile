PRJ := corekg
APP ?= roc
ENV ?= test
PKGDIR := github.com/insmtx/corekg

PWD := $(shell pwd)


GO ?= go

# GO='GOOS=windows GOARCH=386 go'
VERSION ?= $(shell git describe --tags | sed 's/\(.*\)-.*/\1/')
GIT_COMMIT = $(shell git rev-parse --short HEAD || echo unsupported)
GO_VERSION = $(shell go version)
APP_VERSION = $(shell git describe --tags --abbrev=0)
BUILD_AT = $(shell date "+%Y-%m-%dT%H:%M:%S")
TIMESTAMP := $(shell date +%s)
DEPLOY_MODE ?=
TARGET_PLATFORM ?= linux/amd64
BUILDER_IMAGE ?= golang:1.24
BUILDER_PLATFORM ?= linux/amd64
TARGETOS ?= linux
TARGETARCH ?= amd64
CONFIG_PATH ?= ./apps/${APP}/conf/${ENV}/config.yaml
# 开源镜像默认推送到 Docker Hub（ghcr.io 等），可按需覆盖：
IMAGE_NAME = docker.io/${PRJ}/${APP}-api
# IMAGE_TAG = ${VERSION}_${GIT_COMMIT}_${TIMESTAMP}
IMAGE_TAG = ${VERSION}_${GIT_COMMIT}
IMAGE ?= ${IMAGE_NAME}:${IMAGE_TAG}
BASE_IMAGE ?= docker.io/${PRJ}/${APP}:base
WEB_IMAGE_TAG = w_${APP}_${VERSION}
WEB_IMAGE ?= ${IMAGE_NAME}:${WEB_IMAGE_TAG}

LD_FLAGS = -X ${PKGDIR}/version.version=$(VERSION) \
 -X ${PKGDIR}/version.gitCommit=$(GIT_COMMIT) \
 -X ${PKGDIR}/version.builtAt=$(BUILD_AT) \
 -X ${PKGDIR}/version.deployMode=$(DEPLOY_MODE)


ver:
	@echo "Version:   " $(VERSION)
	@echo "Major:     " $(APP_VERSION)
	@echo "Git commit:" $(GIT_COMMIT)
	@echo "Go version:" $(GO_VERSION)
	@echo "OS env:    " $(shell go env GOOS)-$(shell go env GOARCH)
	@echo "Build time:" $(BUILD_AT)


image-tag:
	@echo ${IMAGE_TAG}

run: local
	./bundles/$(APP) -c ${CONFIG_PATH}

local:
	$(GO) build -ldflags="$(LD_FLAGS)" -v -o bundles/$(APP) ./apps/${APP}/cmd/

linux:
	GOOS=linux GOARCH=amd64 $(GO) build -ldflags="$(LD_FLAGS)" -v -o bundles/$(APP)-linux ./apps/${APP}/cmd/

puppyui-windows:
	CGO_ENABLED=1 GOOS=windows GOARCH=386 $(GO) build -ldflags="$(LD_FLAGS)" -v -o bundles/puppyui.exe ./clients/puppyui
	# CGO_ENABLED=1 GOOS=windows GOARCH=386 fyne package -os windows -exe ./clients/puppyui/ -icon ./web/admin/public/images/logo.png

build: generate-docs local

test:
	$(GO) test -v ./...

test-indocker:
	docker run --rm -i \
	-v ${PWD}:/go/src/${PKGDIR} \
	-w /go/src/${PKGDIR} \
	golang:1.24 make test

generate-docs:
	-which swag || go install github.com/swaggo/swag/cmd/swag@latest
	-swag i --parseDependency --parseInternal -g app.go --output ./apps/${APP}/internal/docs --dir ./apps/${APP} --outputTypes=go --instanceName ${APP}

YGG_CLI_VERSION := v1.1.23
YGG_CLI_PKG     := git.yygu.cn/pkg/yggocli

codegen:
	$(if $(MODE),, $(error ❌ 请使用 MODE 参数指定生成模式，例如 MODE=api,module,model))

	@set -e; \
	if ! command -v yggocli >/dev/null 2>&1; then \
		echo "⚠️ 未检测到 yggocli，配置私有仓库并安装 $(YGG_CLI_VERSION)..."; \
		go env -w GOPRIVATE=git.yygu.cn; \
		git config --global url."git@git.yygu.cn:".insteadOf "https://git.yygu.cn/"; \
		go install $(YGG_CLI_PKG)@$(YGG_CLI_VERSION); \
	else \
		INSTALLED_VER=$$(go version -m $$(which yggocli) 2>/dev/null | grep -E "^\s+mod\s+$(YGG_CLI_PKG)" | awk '{print $$3}' || echo ""); \
		echo "🔍 已安装的 yggocli 版本: $$INSTALLED_VER"; \
		echo "🎯 目标版本: $(YGG_CLI_VERSION)"; \
		if [ "$$INSTALLED_VER" != "$(YGG_CLI_VERSION)" ]; then \
			echo "⚠️ yggocli 版本不匹配，重新安装 $(YGG_CLI_VERSION)..."; \
			go install $(YGG_CLI_PKG)@$(YGG_CLI_VERSION); \
		else \
			echo "✅ yggocli 版本已是最新"; \
		fi; \
	fi

	@echo "🔧 开始生成代码：APP=$(APP)，MODE=$(MODE)"
	@cd apps/$(APP) && yggocli generate --mode=$(MODE)

# Docker

build-base-image:
	docker build --platform ${TARGET_PLATFORM} -f ./apps/${APP}/script/Dockerfile.base -t ${BASE_IMAGE} .

build-image:
	docker build \
		--platform ${TARGET_PLATFORM} \
		--build-arg DEPLOY_MODE=${DEPLOY_MODE} \
		--build-arg BUILDER_IMAGE=${BUILDER_IMAGE} \
		--build-arg BUILDER_PLATFORM=${BUILDER_PLATFORM} \
		--build-arg TARGETOS=${TARGETOS} \
		--build-arg TARGETARCH=${TARGETARCH} \
		-f ./apps/${APP}/script/Dockerfile \
		-t ${IMAGE} .

push-image: build-image push-image-exist

push-image-exist:
	docker push ${IMAGE}

linux-client:
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LD_FLAGS)" -v -o bundles/${APP}-linux ./clients/${APP}
	
release-windows:
	mkdir -p ./bundles/${APP_VERSION}/
	GOOS=windows GOARCH=386 go build -ldflags="$(LD_FLAGS)" -v -o bundles/${APP_VERSION}/main.exe ./cmds/${APP}
	cd bundles/${APP_VERSION} && zip dist_windows.zip main.exe
	qshell rput -w ckeyer ${APP}/${APP_VERSION}/dist_windows.zip ./bundles/${APP_VERSION}/dist_windows.zip

# 本地开发依赖：优先使用 docker compose（见 docker-compose.yml.example）。
# 以下 target 仅作兼容保留，密码等默认值与 docker-compose.yml.example 保持一致。
dev-mysql:
	docker compose up -d mysql

dev-es:
	docker compose up -d elasticsearch
