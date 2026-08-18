import time
import threading
from collections import defaultdict
from tqdm import tqdm
from loguru import logger 
from elasticsearch import Elasticsearch, NotFoundError
from elasticsearch.helpers import bulk, streaming_bulk


class ElasticsearchSession:
    """Elasticsearch 单例连接管理器，自动保持激活状态 + 后台心跳检测"""

    _instance = None
    _es = None

    def __new__(cls, *args, **kwargs):
        """保证单例模式"""
        if not cls._instance:
            cls._instance = super().__new__(cls)
        return cls._instance

    def __init__(self, es_host, es_account, es_password, es_pool_size=10, heartbeat_interval=30):
        """初始化参数，只执行一次"""
        if not hasattr(self, "_initialized"):
            # 处理 host 地址
            self.es_host = str(es_host).strip()
            if not self.es_host.startswith(("http://", "https://")):
                self.es_host = f"http://{self.es_host}"
            self.es_account = es_account
            self.es_password = es_password
            self.es_pool_size = es_pool_size
            self.heartbeat_interval = heartbeat_interval
            # 初始化 ES 连接
            self._init_es_client()
            # 启动后台心跳线程
            self._start_heartbeat()
            self._initialized = True

    def _init_es_client(self):
        """初始化 ES 客户端连接"""
        try:
            self._es = Elasticsearch(
                hosts=[self.es_host],
                http_auth=(self.es_account, self.es_password),
                connections_per_node=self.es_pool_size,
                request_timeout=30,
                retry_on_timeout=True,
                max_retries=3,
                verify_certs=False,
                ssl_show_warn=False
            )
            if self._es.ping():
                logger.info("✅ ES 初始化成功")
            else:
                raise ConnectionError("Ping ES 失败")
        except Exception as e:
            raise RuntimeError(f"❌ ES 初始化失败: {e}")

    def _start_heartbeat(self):
        """启动后台心跳线程，定时检测连接是否可用"""
        def heartbeat():
            while True:
                try:
                    if not (self._es and self._es.ping()):
                        logger.error("⚠️ ES 心跳检测失败，尝试重连...")
                        self._init_es_client()
                except Exception:
                    logger.error("⚠️ ES 心跳异常，尝试重连...")
                    self._init_es_client()
                time.sleep(self.heartbeat_interval)

        thread = threading.Thread(target=heartbeat, daemon=True)
        thread.start()

    def get_es(self):
        """获取可用的 ES 客户端"""
        return self._es


class ElasticsearchVectorDBStorage:
    """ES 数据存储操作封装"""

    def __init__(self, index_name, es_session):
        self.index_name = index_name
        self.es_session = es_session
        self.es = self.es_session.get_es()

    def _get_es(self):
        """内部取连接，保证活跃"""
        self.es = self.es_session.get_es()
        return self.es

    async def upsert_chunks(self, datas: dict):
        """插入或更新 chunks 数据（输入数据即为完整文档，不做任何改动，写入后立即可查询，带进度条）"""
        es = self._get_es()

        if len(datas) == 1:
            k, v = next(iter(datas.items()))
            try:
                es.index(index=self.index_name, id=k, body=v, refresh="wait_for")  # 插入或更新
                logger.info(f"✅ 单条插入/更新成功: {k}")
            except Exception as e:
                logger.error(f"❌ 单条插入/更新失败: {k}, 错误: {str(e)}")
        else:
            actions = (
                {
                    "_op_type": "index",  # 插入或更新
                    "_index": self.index_name,
                    "_id": k,
                    "_source": v
                }
                for k, v in datas.items()
            )

            success, failed = 0, 0
            for ok, result in tqdm(
                streaming_bulk(
                    es,
                    actions,
                    refresh="wait_for",
                    raise_on_error=False,
                    max_retries=3,
                ),
                total=len(datas),
                desc="Chunk is inserting ES:"
            ):
                if ok:
                    success += 1
                else:
                    failed += 1
                    # 取出错误原因
                    action, response = result.popitem()
                    error_reason = response.get("error", {}).get("reason", response)
                    logger.error(f"❌ 插入/更新失败: {response.get('_id')}, 错误: {error_reason}")
                    raise

            logger.info(f"✅ 插入/更新完成: 成功 {success} 条，失败 {failed} 条")

    async def del_es_chunk(
        self, 
        forest_id: str, 
        file_id: str, 
        company_id: str, 
        index_name: str = None
    ) -> int:
        """
        根据条件删除 chunk 文档（排除 entity、file_description）
        :return: 实际删除数量
        """
        es = self._get_es()
        index = index_name or self.index_name
        logger.info(f"开始删除索引 {index} 中 forest_id={forest_id}, file_id={file_id}, company_id={company_id} 的所有 chunk 文档...")

        query = {
            "query": {
                "bool": {
                    "must": [
                        {"term": {"forest_id": forest_id}},
                        {"term": {"file_id": file_id}},
                        {"term": {"company_id": company_id}},
                    ],
                    "must_not": [
                        {"terms": {"type": ["entity", "file_description"]}}
                    ]
                }
            }
        }

        try:
            resp = es.delete_by_query(index=index, body=query, conflicts="proceed")
            deleted = resp.get("deleted", 0)
            logger.info(f"删除完成，共删除 {deleted} 条文档")
            return deleted
        except NotFoundError:
            logger.warning(f"索引 {index} 未找到匹配的文档")
            return 0
        except Exception as e:
            logger.error(f"ES 删除失败: {e}", exc_info=True)
            raise RuntimeError(f"ES 删除失败: {e}")
        

    async def upsert_entities(self, datas: dict):
        """插入/更新实体数据"""
        for k, v in datas.items():
            v["forest_id"] = self.forest_id
            v["company_id"] = self.company_id
            v["uin"] = self.uin

        es = self._get_es()
        if len(datas) == 1:
            k, v = next(iter(datas.items()))
            es.index(index=self.index_name, id=k, body=v)
        else:
            actions = [
                {
                    "_op_type": "update",
                    "_index": self.index_name,
                    "_id": str(self.forest_id) + k,
                    "script": {
                        "source": "ctx._source.references.addAll(params.new_references)",
                        "lang": "painless",
                        "params": {"new_references": v["references"]},
                    },
                    "upsert": v,
                }
                for k, v in datas.items()
            ]
            bulk(es, actions)

    def get_by_id(self, id: str):
        """根据 id 获取文档"""
        es = self._get_es()
        try:
            result = es.get(index=self.index_name, id=id)
            doc = result["_source"]
            doc["id"] = doc.get("__id__")
            return doc
        except NotFoundError:
            return []
        except Exception as e:
            raise RuntimeError(f"ES 查询失败: {e}")

    def get_chunks_by_file_id(self, file_id: str):
        """根据文件 ID 获取所有 chunks"""
        es = self._get_es()
        try:
            query = {"query": {"term": {"file_id": file_id}}}
            resp = es.search(index=self.index_name, body=query, scroll="1m", size=1000)
            results = []
            while resp["hits"]["hits"]:
                results.extend([hit["_source"] for hit in resp["hits"]["hits"]])
                resp = es.scroll(scroll_id=resp["_scroll_id"], scroll="1m")
            return results
        except Exception as e:
            raise RuntimeError(f"ES 查询 chunks 失败: {e}")

    def get_entities(self, entity_type: str):
        """获取某类实体"""
        es = self._get_es()
        try:
            query = {"query": {"term": {"entity_type": entity_type}}}
            resp = es.search(index=self.index_name, body=query, scroll="1m", size=1000)
            results = []
            while resp["hits"]["hits"]:
                results.extend([hit["_source"] for hit in resp["hits"]["hits"]])
                resp = es.scroll(scroll_id=resp["_scroll_id"], scroll="1m")
            return results
        except Exception as e:
            raise RuntimeError(f"ES 查询实体失败: {e}")

    def query(self, es_query: dict):
        """执行自定义查询"""
        es = self._get_es()
        try:
            resp = es.search(index=self.index_name, body=es_query, scroll="1m", size=1000)
            results = []
            while resp["hits"]["hits"]:
                results.extend([hit["_source"] for hit in resp["hits"]["hits"]])
                resp = es.scroll(scroll_id=resp["_scroll_id"], scroll="1m")
            return results
        except Exception as e:
            raise RuntimeError(f"ES 查询失败: {e}")
