FROM python:3.10

# 配置清华 pip 源

COPY requirements_analyser.txt /app/requirements.txt
RUN pip3 install --no-cache-dir -r /app/requirements.txt

WORKDIR /app

# 安装系统和 Python 编译依赖
RUN apt-get update && apt-get install -y \
    g++ \
    gcc \
    make \
    cmake \
    python3-dev \
    libssl-dev \
    libffi-dev \
    libgl1 \
    libglib2.0-0 \
    poppler-utils \
    && rm -rf /var/lib/apt/lists/*

# 安装 Nuitka
RUN pip install nuitka

# 设置环境变量
ENV NUMPY_EXPERIMENTAL_DTYPE_API=1
ENV NUITKA_OPENCV_SKIP=1

COPY doc_worker_main.py /app/main.py
COPY corekg_analyser/ /app/corekg_analyser/
COPY tools/ /app/tools/
COPY config/analyser_config.yaml /app/config/analyser_config.yaml

# 编译代码
RUN python3.10 -m nuitka --remove-output --mode=package \
  --include-package=pypdf \
  --include-package=pdf2image \
  --include-package=tools \
  corekg_analyser

# 清理源码
RUN rm -rf /app/corekg_analyser /app/tools

RUN mkdir -p /app/static
    
ENV LANG=C.UTF-8
ENV LC_ALL=C.UTF-8

WORKDIR /app

CMD [ "python3.10", "main.py" ]
