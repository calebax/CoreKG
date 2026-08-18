#!/usr/bin/env bash
#
# CoreKG 基础设施初始化统一入口镜像。
#
# 设计说明：
#   - 对外暴露单一入口脚本，按第一个参数（子命令）分发到对应的 init 任务；
#     后续新增其它 init 需求时，只需新增一个 case 分支即可扩成同一个镜像。
#   - 默认通过 docker-compose 挂载的命名卷持久化，实现"只初始化一次"的幂等目标。
#
# 用法：
#   docker run --rm \
#     -v es_plugins:/usr/share/elasticsearch/plugins \
#     corekg-init es-init [ES_PLUGINS_DIR]
#
# 可用子命令：
#   es-init   幂等安装 IK 分词插件（ik_max_word / ik_smart）
#   空参数     打印用法

set -Eeuo pipefail

# 插件目录，可通过第二个位置参数覆盖（便于测试）；默认与 ES 镜像的插件目录一致。
ES_PLUGINS_DIR="${2:-/usr/share/elasticsearch/plugins}"
# IK 插件(analysis-ik)可执行 zip 的下载地址。需与 ES 版本精确匹配。
IK_ZIP_URL="${IK_ZIP_URL:-https://get.infini.cloud/elasticsearch/analysis-ik/8.18.1}"
# IK 插件解压后必须存在的标志文件，用于幂等判断（目录存在且该文件存在即视为已安装）。
# 不同版本该文件可能变化；判定时以目录/文件两者同时满足为准。
IK_MARKER="${IK_PLUGIN_DIR:-analysis-ik}"

log()  { printf '[init] %s\n' "$*" >&2; }
fail() { log "ERROR: $*"; exit 1; }

# 检查是否已安装 IK 插件（幂等依据：插件目录存在）
ik_installed() {
  [ -d "${ES_PLUGINS_DIR}/${IK_MARKER}" ]
}

es_init() {
  if ik_installed; then
    log "IK 插件已存在于 ${ES_PLUGINS_DIR}/${IK_MARKER}，跳过安装（幂等）"
    return 0
  fi

  log "IK 插件未安装，开始从 ${IK_ZIP_URL} 下载安装..."
  if ! elasticsearch-plugin install --batch "${IK_ZIP_URL}"; then
    log "IK 插件安装失败"
    return 1
  fi

  if ik_installed; then
    log "IK 插件安装完成: ${ES_PLUGINS_DIR}/${IK_MARKER}"
  else
    fail "IK 插件声称安装成功，但未检测到 ${ES_PLUGINS_DIR}/${IK_MARKER}"
  fi
}

usage() {
  sed -n '2,14p' "$0" | sed 's/^  # /  /'
}

CMD="${1:-}"
case "${CMD}" in
  es-init)
    es_init
    ;;
  ""|-h|--help|help)
    usage
    ;;
  *)
    fail "未知子命令: ${CMD}; 可用命令见上方用法"
    ;;
esac
