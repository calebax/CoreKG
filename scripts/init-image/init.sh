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
#   es-init   幂等安装 IK（ik_max_word / ik_smart）与 Smart Chinese 分词插件
#   空参数     打印用法

set -Eeuo pipefail

# 插件目录，可通过第二个位置参数覆盖（便于测试）；默认与 ES 镜像的插件目录一致。
ES_PLUGINS_DIR="${2:-/usr/share/elasticsearch/plugins}"
# IK 插件(analysis-ik)可执行 zip 的下载地址。需与 ES 版本精确匹配。
IK_ZIP_URL="${IK_ZIP_URL:-https://get.infini.cloud/elasticsearch/analysis-ik/8.18.1}"
IK_CONFIG_BASE_URL="${IK_CONFIG_BASE_URL:-https://cdn.jsdelivr.net/gh/infinilabs/analysis-ik@856ceb7/config}"
# IK 插件解压后必须存在的标志文件，用于幂等判断（目录存在且该文件存在即视为已安装）。
# 不同版本该文件可能变化；判定时以目录/文件两者同时满足为准。
IK_MARKER="${IK_PLUGIN_DIR:-analysis-ik}"
SMARTCN_MARKER="${SMARTCN_PLUGIN_DIR:-analysis-smartcn}"

log()  { printf '[init] %s\n' "$*" >&2; }
fail() { log "ERROR: $*"; exit 1; }

# 检查是否已安装 IK 插件（幂等依据：插件目录存在）
ik_installed() {
  [ -d "${ES_PLUGINS_DIR}/${IK_MARKER}" ]
}

smartcn_installed() {
  [ -d "${ES_PLUGINS_DIR}/${SMARTCN_MARKER}" ]
}

ensure_ik_config() {
  local config_dir="${ES_PLUGINS_DIR}/${IK_MARKER}/config"
  local config_file="${config_dir}/IKAnalyzer.cfg.xml"
  local dictionary
  local dictionaries=(
    extra_main.dic
    extra_single_word.dic
    extra_single_word_full.dic
    extra_single_word_low_freq.dic
    extra_stopword.dic
    main.dic
    preposition.dic
    quantifier.dic
    stopword.dic
    suffix.dic
    surname.dic
  )

  mkdir -p "${config_dir}"
  if [ ! -f "${config_file}" ]; then
    printf '%s\n' \
      '<?xml version="1.0" encoding="UTF-8"?>' \
      '<!DOCTYPE properties SYSTEM "http://java.sun.com/dtd/properties.dtd">' \
      '<properties>' \
      '  <comment>IK Analyzer local defaults</comment>' \
      '  <entry key="ext_dict"></entry>' \
      '  <entry key="ext_stopwords">stopword.dic</entry>' \
      '  <entry key="remote_ext_dict"></entry>' \
      '  <entry key="remote_ext_stopwords"></entry>' \
      '</properties>' > "${config_file}"
  fi
  for dictionary in "${dictionaries[@]}"; do
    if [ ! -s "${config_dir}/${dictionary}" ]; then
      log "下载 IK 词典: ${dictionary}"
      curl --fail --location --silent --show-error \
        "${IK_CONFIG_BASE_URL}/${dictionary}" \
        --output "${config_dir}/${dictionary}"
    fi
  done
}

es_init() {
  if smartcn_installed; then
    log "Smart Chinese 插件已存在于 ${ES_PLUGINS_DIR}/${SMARTCN_MARKER}，跳过安装（幂等）"
  else
    log "Smart Chinese 插件未安装，开始安装..."
    if ! elasticsearch-plugin install --batch analysis-smartcn; then
      log "Smart Chinese 插件安装失败"
      return 1
    fi
  fi

  if ik_installed; then
    log "IK 插件已存在于 ${ES_PLUGINS_DIR}/${IK_MARKER}，跳过安装（幂等）"
    ensure_ik_config
    return 0
  fi

  log "IK 插件未安装，开始从 ${IK_ZIP_URL} 下载安装..."
  if ! elasticsearch-plugin install --batch "${IK_ZIP_URL}"; then
    log "IK 插件安装失败"
    return 1
  fi

  if ik_installed; then
    ensure_ik_config
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
