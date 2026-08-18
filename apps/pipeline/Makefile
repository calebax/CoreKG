PRJ := corekg
MODULE ?= graphrag
APP ?= chunker

USE ?= no

PWD := $(shell pwd)

# 确保 VERSION 不以 - 开头
VERSION ?= $(shell git describe --tags | sed 's/^-//')
GIT_COMMIT = $(shell git rev-parse --short HEAD || echo unsupported)
APP_VERSION = $(shell git describe --tags --abbrev=0)
BUILD_AT = $(shell date "+%Y-%m-%dT%H:%M:%S")
BUILD_TS := $(shell date "+%s" | cut "-c7-")
IMAGE_NAME = docker.io/${PRJ}/${MODULE}-${APP}
BASE_VERSION = v1.6

ifeq ($(APP),base)
  IMAGE_TAG := $(BASE_VERSION)
else ifeq ($(APP),building)
  IMAGE_TAG := $(BASE_VERSION)
else
  # 确保 IMAGE_TAG 不以 - 开头
  IMAGE_TAG := $(shell echo $(VERSION)-${GIT_COMMIT}-$(BUILD_TS) | sed 's/^-//')
endif

ifeq ($(USE), no)
  IMAGE_FILE := ./docker/${APP}/Dockerfile
else
  IMAGE_FILE := ./${USE}/docker/${APP}/Dockerfile
endif

PUSH_IMAGE = build-image push-image-exist

build: $(DEPS)


IMAGE ?= ${IMAGE_NAME}:${IMAGE_TAG}

ver:
	@echo "Version:   " $(VERSION)
	@echo "Major:     " $(APP_VERSION)
	@echo "Git commit:" $(GIT_COMMIT)
	@echo "Build time:" $(BUILD_AT)

image-tag:
	@echo ${IMAGE_TAG}

build-image:
	docker build --network host -f ${IMAGE_FILE} -t ${IMAGE} .

push-image: ${PUSH_IMAGE}

push-image-exist:
	docker push ${IMAGE}

build:
	python3 setup.py build_ext

dev:
	docker run -it --rm --memory=4G -v ${PWD}:/app -w /app python:3.10 bash

