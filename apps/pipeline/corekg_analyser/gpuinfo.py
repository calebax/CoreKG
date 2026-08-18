import os, signal
import time
import torch
import pynvml
import signal
import threading
from loguru import logger

class GpuInfo():
    """GPU 状态监控与心跳维持，用于 GPU Pod 环境防止 GPU 被系统回收"""

    def __init__(self):
        self.gpu_status = False
        self.stop_event = threading.Event()

    def gpu_info(self):
        """通过 pynvml / nvidia-smi / torch 三重检测当前 GPU 是否可用"""
        try:
            # 获取gpu信息，如果没有gpu则报错，采用三种方式获取gpu信息保证一定程度的准确性
            # pynvml获取gpu信息：发现在一个进程中第一次获取gpu后，后续无论有无gpu都还能获取到信息
            pynvml.nvmlInit()
            handle = pynvml.nvmlDeviceGetHandleByIndex(0)
            memory_info = pynvml.nvmlDeviceGetMemoryInfo(handle)
            utilization = pynvml.nvmlDeviceGetUtilizationRates(handle)
            gpu_info = f"pynvml获取的信息：Total memory: {memory_info.total}, Free memory: {memory_info.free}, Used memory: {memory_info.used}, " + \
                f"GPU Utilization: {utilization.gpu}, Memory Utilization: {utilization.memory}, " + \
                f"GPU used rate: {memory_info.used/memory_info.total}"
            # logger.info(gpu_info)

            # 执行nvidia-smi命令获取gpu信息
            cmd = 'nvidia-smi'
            out = os.popen(cmd)
            gpu_str = out.read()
            if 'NVML: Unknown Error' in gpu_str:
                raise Exception('当前GPU似乎有掉卡，请检查GPU状态。')

            # 通过torch获取gpu信息
            gpu_info = f"torch.cuda获取的信息：当前GPU的存在状态：{torch.cuda.is_available()}" + \
                f"当前GPU的device块数: {torch.cuda.device_count()}" + \
                f"当前GPU的名字：{torch.cuda.get_device_name(torch.cuda.current_device())}" + \
                f"使用率：{torch.cuda.memory_allocated() / 1024**2} MB 已使用, {torch.cuda.memory_reserved() / 1024**2} MB 已缓存"
            free, total = torch.cuda.mem_get_info(0)
            # logger.info(f"显存: {free/1024**3:.2f} GB free / {total/1024**3:.2f} GB total")
            result = True
        except:
            gpu_info = "未获取到任何GPU信息"
            result = False
        logger.info(gpu_info)
        return result

    def gpu_calculate(self):
        """执行一次简单的 GPU 张量运算，确保 GPU 计算通道可用"""
        # 进行一次简单的gpu计算，确保gpu可用
        n = 20_000_000  # 数据规模

        device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
        # logger.info("使用设备:", device)

        # 准备数据（直接放到 GPU 上）
        x = torch.arange(n, dtype=torch.int32, device=device)
        y = 2 * x

        # GPU 计算
        torch.cuda.synchronize()
        start = time.time()
        result = x + y
        torch.cuda.synchronize()
        end = time.time()

        # 验证结果
        expected = torch.arange(n, dtype=torch.int32, device=device) * 3

    def gpu_heartbeat(self, stop_handler):
        """后台心跳循环：每 60 秒检测一次 GPU 状态并执行简单计算，掉卡时发送 SIGTERM 重启 Pod"""
        # 每分钟检查一次gpu状态，如果有gpu则进行一次简单的计算，确保gpu不被系统回收
        self.gpu_status = self.gpu_info()
        logger.info(f"初始化时gpu是否存在: {self.gpu_status}")
        try:
            while not self.stop_event.is_set():
                
                if self.gpu_status:
                    result = self.gpu_info()
                    if not result:
                        raise Exception("未发现GPU")
                    self.gpu_calculate()
                    # logger.info("初始化时有gpu，进行gpu心跳维持运算")
                time.sleep(60)
        except:
            self.stop_event.set()
            logger.error("gpu_Loop发现问题, GPU掉卡，该线程结束，stop_event.set()，POD重启")
            os.kill(os.getpid(), signal.SIGTERM)

    def gpu_check_func(self, stop_handler):
        """启动 GPU 心跳守护线程"""
        gpu_check_thread = threading.Thread(target=self.gpu_heartbeat, args=(stop_handler, ), daemon=True)
        gpu_check_thread.start()
        logger.info("GPU检查后台程序已开始运行✅")
