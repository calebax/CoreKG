"""
CoreKG 知识切块 Worker 启动入口。
负责注册信号处理、启动 CPU 心跳线程和异步主循环。
"""

import asyncio
import threading
import signal
import time, math
import os, sys, certifi
from loguru import logger
os.environ['SSL_CERT_FILE'] = certifi.where()
os.environ['REQUESTS_CA_BUNDLE'] = certifi.where()


_forced = False

logger.add("./logs/{time:YYYYMMDD}.log", rotation="00:00", retention="7 days", encoding="utf-8", enqueue=True)


def cpu_heartbeat():
    """后台 CPU 心跳线程，通过持续浮点计算防止系统因空闲而回收 CPU 资源"""
    def calculate():
        now = time.time()
        x = math.ceil(now) % 10
        y = math.floor(now) % 10
        for i in range(100):
            x += x
            y += y
    while True:
        try:
            calculate()
        except Exception as e:
            logger.exception(f"cpu keep alive bug: {e}")
        time.sleep(30)

def worker():
    """Worker 主函数：启动心跳线程并运行异步任务消费循环"""
    from corekg_chunk.task.work_chunk import main
    try:
        cpu_th = threading.Thread(target=cpu_heartbeat, daemon=True)
        cpu_th.start()
        asyncio.run(main())
        logger.info("Worker 正常退出")
    except KeyboardInterrupt:
        logger.info("Received Ctrl+C, shutting down...")
    except Exception as e:
        logger.error(f"Chunk Work Error: {e}")


def signal_handler(sig, frame):
    """信号处理器：首次信号优雅退出，二次信号强制终止"""
    global _forced
    from corekg_chunk.task.work_chunk import STOP_EVENT
    if not STOP_EVENT.is_set():
        logger.info("接收到退出信号，等待当前任务完成后退出（再次信号强制终止）...")
        STOP_EVENT.set()
    else:
        logger.info("二次信号，强制终止")
        _forced = True
        os._exit(1)


if __name__ == '__main__':
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    worker()
