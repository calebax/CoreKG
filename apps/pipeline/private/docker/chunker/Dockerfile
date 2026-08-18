FROM python:3.10

COPY chunk_worker_main.py /app/main.py

COPY corekg_chunk/ /app/corekg_chunk/
COPY tools/ /app/tools/
COPY config/chunk_config.yaml /app/config/chunk_config.yaml


RUN apt-get update && apt-get install -y \
    g++ \
    gcc \
    make \
    cmake \
    python3-dev \
    libssl-dev \
    libffi-dev \
    && rm -rf /var/lib/apt/lists/*

RUN pip install nuitka
WORKDIR /app

RUN python3.10 -m nuitka --remove-output --mode=package \
  --include-package=httpx \
  --include-package=httpcore \
  --include-package=anyio \
  --include-package=certifi \
  --include-package=ssl \
  --include-package=idna \
  --include-package=tools \
  corekg_chunk

RUN rm -rf /app/corekg_chunk /app/tools

VOLUME [ "/app/logs/" ]

CMD ["python3.10", "main.py"]
