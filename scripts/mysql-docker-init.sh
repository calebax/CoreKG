#!/bin/sh
# MySQL docker-entrypoint-initdb.d 初始化脚本（首次创建数据卷时执行一次）。
# 创建 opencoze（Coze）数据库，并授权 corekg 用户，供 kechat/keinit/workflow 等应用连接使用。
set -e

# 主库 corekg 由 MYSQL_DATABASE 自动创建；这里补充创建 opencoze。
mysql -uroot -p"${MYSQL_ROOT_PASSWORD}" <<-EOSQL
  CREATE DATABASE IF NOT EXISTS ${MYSQL_EXTRA_DATABASES} DEFAULT CHARACTER SET utf8mb4;
  GRANT ALL PRIVILEGES ON ${MYSQL_EXTRA_DATABASES}.* TO '${MYSQL_USER}'@'%';
  FLUSH PRIVILEGES;
EOSQL
