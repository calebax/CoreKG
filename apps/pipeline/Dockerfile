# CoreKG Pipeline — 开发用普通 Docker 镜像。
# 直接 COPY 源码 + 安装 requirements（不做 Nuitka 加密编译），供 docker-compose 统一拉起。
# 同一份镜像承载 analyser（doc worker）与 chunker（chunk worker），由 compose 的 command 指定入口。
FROM python:3.10

ENV PYTHONUNBUFFERED=1

WORKDIR /app

# 先装依赖以利用镜像层缓存
COPY requirements.txt requirements_analyser.txt ./
RUN pip install --no-cache-dir -r requirements.txt -r requirements_analyser.txt

# 拷贝全部源码（chunk_worker_main.py / doc_worker_main.py / corekg_chunk / corekg_analyser / tools / config）
COPY . .

VOLUME [ "/app/logs" ]
