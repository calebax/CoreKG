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
# 默认单平台构建（向后兼容）；如需双架构自行覆盖，如：BUILD_PLATFORMS='linux/amd64,linux/arm64'
TARGET_PLATFORM ?= linux/amd64
BUILD_PLATFORMS ?= $(TARGET_PLATFORM)
BUILDER_IMAGE ?= golang:1.24
# buildx 多平台构建时，构建器(builder)使用的平台（默认跟随目标，避免跨架构编译 CGO 依赖）
BUILDER_PLATFORM ?= linux/amd64
TARGETOS ?= linux
# 用于单平台 `linux`/`docker build --platform` 的架构；buildx 多平台时由构建器自动注入
TARGETARCH ?= amd64
GOARCH ?= amd64
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
	GOOS=linux GOARCH=$(GOARCH) $(GO) build -ldflags="$(LD_FLAGS)" -v -o bundles/$(APP)-linux ./apps/${APP}/cmd/

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

# Docker
# 双架构判断：BUILD_PLATFORMS 含逗号 => 多平台 buildx（需推送 registry）；否则单平台（保持原 docker build 行为）
# 注意：多平台构建（BUILD_PLATFORMS 含逗号）要求 buildx builder 使用 docker-container driver，
#       否则会报 "Multi-platform build is not supported for the docker driver"。
#       可用 `docker buildx create --driver docker-container --name corekg-multi` 创建并设置 BUILDER=corekg-multi。
COMMA := ,
MULTI_ARCH := $(findstring $(COMMA),$(BUILD_PLATFORMS))
BUILDER ?= default

build-base-image:
ifeq ($(MULTI_ARCH),)
	docker build --platform ${TARGET_PLATFORM} -f ./apps/${APP}/script/Dockerfile.base -t ${BASE_IMAGE} .
else
	docker buildx build \
		--builder ${BUILDER} \
		--platform ${BUILD_PLATFORMS} \
		-f ./apps/${APP}/script/Dockerfile.base \
		-t ${BASE_IMAGE} \
		--push .
endif

build-image:
ifeq ($(MULTI_ARCH),)
	docker build \
		--platform ${TARGET_PLATFORM} \
		--build-arg DEPLOY_MODE=${DEPLOY_MODE} \
		--build-arg BUILDER_IMAGE=${BUILDER_IMAGE} \
		--build-arg TARGETOS=${TARGETOS} \
		--build-arg TARGETARCH=${TARGETARCH} \
		-f ./apps/${APP}/script/Dockerfile \
		-t ${IMAGE} .
else
	docker buildx build \
		--builder ${BUILDER} \
		--platform ${BUILD_PLATFORMS} \
		--build-arg DEPLOY_MODE=${DEPLOY_MODE} \
		--build-arg BUILDER_IMAGE=${BUILDER_IMAGE} \
		-f ./apps/${APP}/script/Dockerfile \
		-t ${IMAGE} \
		--push .
endif

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
