import json
import threading
import yaml
import socket
import os
import requests
import asyncio
from loguru import logger 
from corekg_chunk.chunk import chunk_process
from typing import Dict, Optional, Any
from tools.es_storage import ElasticsearchSession, ElasticsearchVectorDBStorage


# ===================== 加载配置 =====================
with open("./config/chunk_config.yaml", "r", encoding="utf-8") as file:
    config = yaml.safe_load(file)

# 全局配置
worker_id = f"{socket.gethostbyname(socket.gethostname())}-{os.getpid()}"
STOP_EVENT = threading.Event()

logger.info("日志初始化完成")

# 初始化 ES 连接
es_session = ElasticsearchSession(
    es_host = config["ES"]["HOST"],
    es_account = config["ES"]["ACCOUNT"],
    es_password = config["ES"]["PASS"], 
    es_pool_size = config["ES"]["POOL_SIZE"],
    heartbeat_interval = 10
)


async def process_task(task_data: Dict[str, Any]) -> Dict[str, Any]:
    """Process chunk task"""
    task_id = task_data.get("task_id")
    try:
        # 转换字符串 为 json
        payload = json.loads(task_data["payload"])
        logger.info(f"Request: {payload}")

        sc = payload.get('split_config', {})
        pre = sc.get('preprocessing_rules', {})
        llm_cfg = payload.get('llm', {})

        # chunk拆分
        logger.info(f"[{task_id}] Processing chunkfile...")

        chunks_emb_dic = await chunk_process(
            url=payload['file_url'],
            forest_id=payload['forest_id'],
            company_id=payload['company_id'],
            uin=payload['uin'],
            file_id=payload['file_id'],
            file_name=payload.get("file_name"),
            file_ext=payload.get("file_ext"),
            index_name=payload.get("es_index"),
            # 预处理
            remove_email=pre.get('remove_email', True),
            remove_URL=pre.get('remove_url', True),
            remove_empty_line=pre.get('remove_empty_line', True),
            # 分块策略
            mode=sc.get('split_mode', 'smart'),
            chunk_token_num=sc.get('chunk_token_num') or sc.get('chunk_size') or 1024,
            min_chunk_tokens=sc.get('min_chunk_tokens', 10),
            split_level=sc.get('split_level', 2),
            # overlap_ratio=sc.get('overlap_ratio') or sc.get('split_overlap') or 0.0,
            overlap_ratio=0.0,
            regex_pattern=sc.get('regex_pattern') or sc.get('split_mark'),
            delimiter=sc.get('delimiter', "\n!?。；！？"),
            enable_heading_in_content=sc.get('enable_heading_in_content', False),
            # 表格增强 LLM
            llm_enabled=sc.get('llm_enabled'),
            llm_model=sc.get('llm_model') or llm_cfg.get('model_name'),
            llm_api_key=sc.get('llm_api_key') or llm_cfg.get('api_key'),
            llm_base_url=sc.get('llm_base_url') or llm_cfg.get('base_url'),
            llm_timeout=sc.get('llm_timeout'),
            # 图片增强 VLLM
            vllm_enabled=sc.get('vllm_enabled'),
            vllm_model=sc.get('vllm_model'),
            vllm_api_key=sc.get('vllm_api_key'),
            vllm_base_url=sc.get('vllm_base_url'),
            image_width=sc.get('image_width'),
            image_height=sc.get('image_height'),
            # Embedding
            embedding_model=sc.get('embedding_model'),
            embedding_api_key=sc.get('embedding_api_key'),
            embedding_base_url=sc.get('embedding_base_url'),
            # 并发控制
            eb_max_concurrency=sc.get('eb_max_concurrency'),
            llm_max_concurrency=sc.get('llm_max_concurrency'),
        )
        logger.info("chunk_process 完成，结果返回。")


        # 4、es批量入库
        storage = ElasticsearchVectorDBStorage(
            index_name = payload.get("es_index"),
            es_session = es_session
        )

        # 先清空对应的历史遗留chunk
        await storage.del_es_chunk(payload['forest_id'], payload['file_id'], payload['company_id'], payload.get("es_index"))

        # 新数据上传
        await storage.upsert_chunks(chunks_emb_dic)
        logger.info("**********************ES上传已完成！*************************")

        logger.info(f"[{task_id}] Processing file success...")

        result = {
            "status": "success",
            "worker_id": worker_id,
            "task_id": task_id
        }

        return result
    except Exception as e:
        logger.info(f"[{task_id}] Task failed: {str(e)}")
        logger.exception(f"拆chunk发生错误❌：task_id--{task_id}, file_id--{payload['file_id']}, exception--{e}")
        raise

class TaskQueue:
    def __init__(self, task_type: str, url: str = "https://api.example.com/v3",
                 apikey: str = None):
        self.url = url
        self.apikey = apikey
        self.task_type = task_type
        self.get_task_url = f"{self.url}/knowledge.GetPendingTask"
        self.callback_url = f"{self.url}/knowledge.TaskCallBack"

    def header(self) -> Dict[str, str]:
        headers = {
            "Content-Type": "application/json",
            "Accept": "application/json",
        }
        if self.apikey:
            headers["Authorization"] = f"Bearer {self.apikey}"
        return headers

    def get_task(self, timeout: Optional[int] = None) -> Optional[Dict[str, Any]]:
        """Get pending task"""
        body = {
            "Request": {
                "task_type": self.task_type,
                "worker_id": worker_id,
            }
        }

        try:
            response = requests.post(
                self.get_task_url,
                json=body,
                headers=self.header(),
                timeout=timeout
            )
            response.raise_for_status()
            data = response.json()
            logger.info("Task response: %s" % data)

            task_data = data.get("Response", {})
            if not task_data:
                logger.debug("No tasks available")
                return None

            return task_data

        except requests.RequestException as e:
            logger.warning(f"Failed to get task: {e}")
            return None

    def callback(self, task_id: str, task_type: str, status: str,
                 err: Optional[str] = None,
                 result: Optional[Any] = None,
                 timeout: Optional[int] = None) -> bool:
        """Callback with task results"""
        try:
            body = {
                "Request": {
                    "task_id": task_id,
                    "task_type": task_type,
                    "status": status,
                    "worker_id": worker_id,
                }
            }

            if err:
                body["Request"]["error_message"] = str(err)
            if result:
                body["Request"]["result"] = result

            response = requests.post(
                self.callback_url,
                json=body,
                headers=self.header(),
                timeout=timeout
            )
            response.raise_for_status()

            logger.info(f"[{task_id}] Callback (status: {status})")
            return True

        except requests.RequestException as e:
            logger.exception(f"[{task_id}] 抛出: {e}")
            return False

class TaskConsumer:
    def __init__(self, task_queue_client: TaskQueue):
        self.task_queue = task_queue_client
        self._active_task: Optional[str] = None  # 当前正在执行的任务 ID

    async def run(self):
        global status, result, error_msg
        logger.info(f"Worker started (ID: {worker_id})")

        while not STOP_EVENT.is_set():
            task_data = None
            try:
                # 使用较短超时以便频繁检查 STOP_EVENT；超时属于正常等待，降级为 warning
                task_data = self.task_queue.get_task(timeout=60)
                if not task_data:
                    if STOP_EVENT.is_set():
                        break
                    await asyncio.sleep(5)
                    continue

                task_id = task_data.get("task_id")
                try:
                    payload = json.loads(task_data.get("payload", "{}"))
                except json.JSONDecodeError as e:
                    logger.warning(f"[INVALID_TASK] Bad JSON in task {task_id}: {str(e)}. Payload: {task_data}...")
                    self.task_queue.callback(
                        task_id=task_id,
                        task_type="unknown",
                        status="fail",
                        err="Bad payload format"
                    )
                    continue

                task_type = payload.get("task_type")
                if not task_id or not task_type:
                    logger.error(f"[INVALID_TASK] Missing required fields, task_id: {task_id}, task_type: {task_type}")
                    self.task_queue.callback(
                        task_id=task_id,
                        task_type="unknown",
                        status="fail",
                        err="Missing required fields"
                    )
                    continue

                # 2. 处理任务
                logger.info(f"[{task_id}] Task acquired")
                self._active_task = task_id
                try:
                    result = await process_task(task_data)
                    status = "success"
                    error_msg = None
                except Exception as e:
                    logger.error(f"[{task_id}] Processing failed: {e}", exc_info=True)
                    result = {"status": "fail", "error": str(e), "task_id": task_id}
                    status = "fail"
                    error_msg = str(e)
                finally:
                    self._active_task = None

                # 3. 发送回调
                self.task_queue.callback(
                    task_id=task_id,
                    task_type=task_type,
                    status=status,
                    result=json.dumps(result) if status == "success" else None,
                    err=error_msg
                )
                logger.info(f"[{task_id}] Callback sent with status: {status}")

            except Exception as e:
                # 仅捕获全局崩溃（如网络断开）
                logger.critical(f"SYSTEM FAILURE: {e}", exc_info=True)
                await asyncio.sleep(5)

        # 优雅退出
        if self._active_task:
            logger.info(f"退出前等待当前任务 [{self._active_task}] 完成...")
        logger.info("Worker 已优雅退出")


async def main():
    """Chunk Worker 主入口：初始化任务队列并启动消费循环"""
    task_type = config["Work"]["TASK_TYPE"]
    api_url = config["Work"]["API_URl"]
    apikey = None

    task_queue = TaskQueue(
        task_type=task_type,
        url=api_url,
        apikey=apikey
    )

    consumer = TaskConsumer(task_queue)
    await consumer.run()

    logger.info("✅ All tasks completed, exiting")

if __name__ == "__main__":
    asyncio.run(main())
