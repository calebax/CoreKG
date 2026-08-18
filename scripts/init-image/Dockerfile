# CoreKG 基础设施统一初始化镜像。
#
# 定位：
#   - 聚合各中间件的初始化任务（当前为 es-init：幂等安装 IK 分词插件），
#     后续新增 init 需求时在此镜像内扩展 init.sh 的分发分支即可合并到同一镜像。
#   - 基于官方 elasticsearch 镜像：复用其自带的 elasticsearch-plugin 工具，
#     并保证 IK 版本与 ES 版本（8.18.1）精确匹配的语义不变。
#
# 构建（可选）：
#   docker build -t docker.io/corekg/init:latest -f scripts/init-image/Dockerfile .
#   docker build -t docker.io/corekg/init:latest \
#     --build-arg ES_IMAGE=elasticsearch:8.18.1 \
#     -f scripts/init-image/Dockerfile .

# ES 基础镜像版本可由构建参数覆盖；默认与 apps 使用的 ES 版本保持一致。
ARG ES_IMAGE=elasticsearch:8.18.1

FROM ${ES_IMAGE}

# 注意：elasticsearch 官方镜像默认以 elasticsearch 用户运行，
# 且 /usr/local/bin 对运行用户不可写，故用 COPY --chmod 直接落地可执行位，
# 避免借助 RUN chmod（该 RUN 会以非 root 用户执行而失败）。
COPY --chmod=755 init.sh /usr/local/bin/init.sh

# 与 docker-compose 中挂载到 es-init 的插件命名卷一致，便于幂等判断与持久化。
VOLUME ["/usr/share/elasticsearch/plugins"]

ENTRYPOINT ["/usr/local/bin/init.sh"]
