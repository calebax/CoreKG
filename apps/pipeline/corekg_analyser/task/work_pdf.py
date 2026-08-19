import json
import threading
import asyncio

import shutil
import socket
import os
import requests

from typing import Dict, Optional, Any
import yaml
from loguru import logger
from tools.s3_common import S3

from corekg_analyser.json_to_md import convert_content_list_to_markdown
from corekg_analyser.analyser_process import process_pdf_task_api, find_output_files, add_prefix_and_save
from corekg_analyser.others.content_process import others_process
# ===================== 加载配置 =====================
with open("./config/analyser_config.yaml", "r", encoding="utf-8") as file:
    config = yaml.safe_load(file)

# 全局配置
worker_id = f"{socket.gethostbyname(socket.gethostname())}-{os.getpid()}"
STOP_EVENT = threading.Event()
LOG_FILE = "../task_logs/worker.log"
webhook = config.get("WEBHOOK")
os.makedirs(os.path.dirname(LOG_FILE), exist_ok=True)

def delete_dir(*args):
    for arg in args:
        if os.path.exists(arg):
            shutil.rmtree(arg)
    logger.info(f"临时目录{args}删除完成")


def send_to_wechat(content: str, webhook: str):
    """发送文本消息到企微群机器人"""
    url = webhook
    headers = {"Content-Type": "application/json"}
    data = {
        "msgtype": "markdown",
        "markdown": {
            "content": content
        }
    }
    response = requests.post(url, json=data, headers=headers)
    if response.status_code != 200:
        logger.critical(f"企微【算子报警】消息发送失败：{response.text}")

def process_pdf_task_mineru(local_file_path, output_target_path, pub_path, task_id=''):
    """PDF 文件解析主流程：MinerU 解析 → bbox 转换 → Markdown 生成 → 前缀注入"""

    output_history_path = os.path.join(output_target_path, "history")

    # [1/5] MinerU API 解析
    logger.info(f"[{task_id}] [1/5] 正在调用 MinerU API 解析文档...")
    process_status = process_pdf_task_api(local_file_path=local_file_path, output_target_path=output_history_path)
    logger.info(f"[{task_id}] [1/5] MinerU 解析完成")

    # [2/5] 查找输出文件 + bbox 坐标转换
    logger.info(f"[{task_id}] [2/5] 正在提取解析产物并转换坐标...")
    process_result = find_output_files(output_history_path)

    if process_result.get('json_path') and process_result.get('origin_path'):
        try:
            import json as _json
            from pypdf import PdfReader

            _export_dpi = 300
            _reader = PdfReader(process_result['origin_path'])
            _page_dims = [
                (float(p.mediabox.width), float(p.mediabox.height))
                for p in _reader.pages
            ]

            with open(process_result['json_path'], 'r', encoding='utf-8') as _f:
                _cl = _json.load(_f)

            _px_per_pt = _export_dpi / 72.0
            _bbox_count = 0
            for _item in _cl:
                _bbox = _item.get('bbox')
                _page = _item.get('page_idx', 0)
                if not _bbox or len(_bbox) != 4 or _page >= len(_page_dims):
                    continue
                _pw, _ph = _page_dims[_page]
                _sx, _sy = _pw / 1000, _ph / 1000
                _item['bbox'] = [
                    int(_bbox[0] * _sx * _px_per_pt),
                    int(_bbox[1] * _sy * _px_per_pt),
                    int(_bbox[2] * _sx * _px_per_pt),
                    int(_bbox[3] * _sy * _px_per_pt),
                ]
                _bbox_count += 1

            with open(process_result['json_path'], 'w', encoding='utf-8') as _f:
                _json.dump(_cl, _f, ensure_ascii=False, indent=2)
            logger.info(f"[{task_id}] [2/5] bbox 坐标已转换 {_bbox_count} 个 (0-1000 → {_export_dpi} DPI)")
        except Exception as e:
            logger.warning(f"[{task_id}] [2/5] bbox 坐标转换跳过: {e}")
    else:
        logger.info(f"[{task_id}] [2/5] 解析产物提取完成")

    # [3/5] JSON → Markdown（bbox 已在步骤 2 转为 300 DPI）
    logger.info(f"[{task_id}] [3/5] 正在将 content_list.json 转换为 Markdown...")
    content = convert_content_list_to_markdown(
        process_result.get('json_path', ''),
        f'{output_target_path}/content.md',
        process_result.get('images_path', ''),
        f'{output_target_path}/images',
    )
    logger.info(f"[{task_id}] [3/5] Markdown 生成完成: {output_target_path}/content.md")

    # [4/5] 清理中间文件
    shutil.rmtree(output_history_path)
    logger.info(f"[{task_id}] [4/5] 中间文件已清理")

    # [5/5] 注入图片公网前缀
    add_prefix_and_save(md_file_path=f'{output_target_path}/content.md', pub_path=f'{pub_path}')
    logger.info(f"[{task_id}] [5/5] 图片前缀注入完成")
    
async def process_pdf_task_local(task_id, payload, local_save_path, output_target_path):
    """文档解析本地处理：S3 下载 → 格式路由 → 解析/转换"""

    file_url = payload.get('file_url')
    file_name = str(file_url).split('/')[-1].split('?')[0]
    file_type = payload.get('file_ext')

    # [1/2] S3 下载
    s3_client = S3()
    logger.info(f"[{task_id}] [1/2] 正在从 S3 下载文件: {file_name}")
    pdf_path, msg = s3_client.download_file(
        download_url=file_url,
        save_dir=local_save_path
    )
    logger.info(f"[{task_id}] [1/2] 下载完成: {msg}")

    public_path = f"{s3_client.get_endpoint()}/{payload.get('bucket')}/{payload.get('upload_path')}"
    os.makedirs(os.path.dirname(f"{local_save_path}"), exist_ok=True)

    # [2/2] 格式路由与解析
    logger.info(f"[{task_id}] [2/2] 正在处理文件 (类型: {file_type or 'PDF'})...")
    if file_type in ['.txt', '.md', '.log', '.csv', '.json', '.mp4', '.avi', '.mkv']:
        others_process(pdf_file_name=pdf_path, final_md_file=f"{output_target_path}/content.md", public_path=public_path, file_ext=file_type)
        logger.info(f"[{task_id}] [2/2] 文件处理完成")
    else:
        process_pdf_task_mineru(local_file_path=pdf_path, output_target_path=output_target_path, pub_path=public_path, task_id=task_id)

    s3_client.s3_client.close()
    return True

async def process_pdf_task(task_data: Dict[str, Any]) -> Dict[str, Any]:
    """文档解析任务入口：本地处理 → S3 上传 → 清理临时文件"""
    task_id = task_data.get("task_id")
    payload = json.loads(task_data["payload"])
    file_id = payload.get('file_id')

    logger.info(f"[{task_id}] ========== 文档解析任务开始 ==========")
    logger.info(f"[{task_id}] 文件: {payload.get('file_name', 'unknown')}, 类型: {payload.get('file_ext', 'unknown')}")

    try:
        import uuid
        uuid4_str = uuid.uuid4().hex
        local_save_path = f'./results/{uuid4_str}_tmp/'
        output_target_path = f"./results/{uuid4_str}_output"
        os.makedirs(output_target_path, exist_ok=True)

        # [1/3] 下载 + 解析
        await process_pdf_task_local(task_id, payload, local_save_path, output_target_path)

        # [2/3] S3 上传
        logger.info(f"[{task_id}] [2/3] 正在上传解析结果到 S3...")
        s3_client = S3()
        s3_file_urls, log_msg = s3_client.upload_directory(
            s3_base_path=payload.get('upload_path').strip(),
            local_dir=output_target_path,
            bucket=payload.get('bucket')
        )
        s3_file_path = s3_file_urls[0]
        logger.info(f"[{task_id}] [2/3] S3 上传完成: {s3_file_path}")

        # [3/3] 清理临时文件
        logger.info(f"[{task_id}] [3/3] 正在清理临时文件...")
        delete_dir(local_save_path, output_target_path)
        logger.info(f"[{task_id}] [3/3] 临时文件清理完成")

        result = {
            "status": "success",
            "s3_file_path": s3_file_path,
            "worker_id": worker_id,
            "task_id": task_id
        }
        logger.info(f"[{task_id}] ========== 文档解析任务完成 ==========")

    except Exception as e:
        logger.exception(f"文档解析发生错误❌：task_id--{task_id}, file_id--{file_id}, exception--{e}")
        
        try:
            if webhook:
                message = f'<font color="warning">** 文档解析发生错误❌：task_id--{task_id}, file_id--{file_id}, exception--{e} **</font>\n'
                send_to_wechat(f'项目名:<font color="info">**【 corekg_doc_analyser 】**</font>\n {message}', webhook)
        except Exception as e:
            logger.critical(f"报错告警无法推送：{e}")

        result = {
            "status": "fail",
            "s3_file_path": None,        # 默认只去第一项
            "worker_id": worker_id,
            "task_id": task_id
        }
    delete_dir(local_save_path, output_target_path)
    return result

class TaskQueue:
    def __init__(self, task_type: str, url: str = "https://api.example.com/v3",
                 apikey: str = None):
        self.url = url
        self.apikey = apikey
        self.task_type = task_type
        self.get_task_url = f"{self.url}/knowledge.GetPendingTask"
        self.callback_url = f"{self.url}/knowledge.TaskCallBack"
        logger.info(f"Worker get task from {self.url}")

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
            logger.debug(f"Task response: {data}")

            task_data = data.get("Response", {})
            actual_id = task_data.get("task_id", 0)
            # corekg 在无任务时返回 HTTP 200 + code=404 + Response:{task_id:0,payload:""}，
            # 需把空任务识别为“无任务”，否则会误当成真实任务触发失败回调风暴 + 端口耗尽。
            if not task_data or not actual_id or not task_data.get("payload"):
                logger.debug("No tasks available")
                return None

            return task_data

        except requests.RequestException as e:
            logger.error(f"Failed to get task: {e}")
            return None

    def callback(self, task_id: int, task_type: str, status: str,
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
            logger.error(f"[{task_id}] 抛出: {e}")
            return False

class TaskConsumer:
    def __init__(self, task_queue_client: TaskQueue):
        self.task_queue = task_queue_client
        self._active_task: Optional[str] = None

    async def run(self):
        global status, result, error_msg
        logger.info(f"Worker started (ID: {worker_id})")

        while not STOP_EVENT.is_set():
            task_data = None
            try:
                # 使用较短超时，以便频繁检查 STOP_EVENT 实现快速退出
                task_data = self.task_queue.get_task(timeout=60)
                if not task_data:
                    if STOP_EVENT.is_set():
                        break
                    continue

                task_id = task_data.get("task_id")
                try:
                    payload = task_data.get("payload")
                    # payload 是否是字符串
                    if isinstance(payload,str):
                        payload = json.loads(task_data.get("payload", "{}"))
                        logger.info(f"current playload data:{payload}")
                    else:
                        logger.info(f"current playload data:{payload}, format erro!!!")
                        self.task_queue.callback(
                            task_id=task_id,
                            task_type="unknown",
                            status="fail",
                            err=" payload is not str"
                        )
                        continue
                except json.JSONDecodeError:
                    logger.warning(f"[INVALID_TASK] Bad payload format")
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
                    result = await process_pdf_task(task_data)
                    status = result['status']  # 从结果中获取状态
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
                logger.info(f"[{task_id}] Callback content: {result}")

            except Exception as e:
                # 仅捕获全局崩溃（如网络断开）
                logger.warning(f"SYSTEM FAILURE: {e}", exc_info=True)
                await asyncio.sleep(5)

        # 优雅退出
        if self._active_task:
            logger.info(f"退出前等待当前任务 [{self._active_task}] 完成...")
        logger.info("Worker 已优雅退出")


async def main():
    with open(os.getenv('COREKG_CONFIGPATH', './config/analyser_config.yaml'), 'r', encoding='utf-8') as file:
        position_data = yaml.safe_load(file)
    
    if webhook:
        logger.info(f"Success, 当前处于【 {config.get('ENV')} 】环境，企微报警推送生效中✅")
    else:
        logger.warning(f"企微告警无法开启，webhook->{webhook} ❌")
        
    # Initialize task queue
    task_queue = TaskQueue(
        task_type="ke.prase_pdf_task",
        url=position_data['api_url'],
        apikey=None
    )

    # Create and run task consumer
    consumer = TaskConsumer(task_queue)
    await consumer.run()

    logger.info("✅ All tasks completed, exiting")

