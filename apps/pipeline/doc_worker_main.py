"""
CoreKG 文档解析 Worker 启动入口。
负责注册信号处理并启动异步任务消费循环。
"""

import asyncio
import os
import signal
from loguru import logger
from corekg_analyser.task.work_pdf import main, STOP_EVENT

_stop_requested = False

def interrupt_handler(*args, **kwargs):
    """信号处理器：首次信号优雅退出，二次信号强制终止"""
    global _stop_requested
    if not _stop_requested:
        logger.info("接收到退出信号，等待当前任务完成后退出（再次信号强制终止）...")
        _stop_requested = True
        STOP_EVENT.set()
    else:
        logger.info("二次信号，强制终止")
        os._exit(1)

def main_():
    """启动入口：注册信号处理并运行异步事件循环"""
    signal.signal(signal.SIGINT, interrupt_handler)
    signal.signal(signal.SIGTERM, interrupt_handler)

    asyncio.run(main())
    logger.info("Worker 正常退出")


if __name__ == '__main__':
    main_()
