#!/usr/bin/env python3
"""CoreKG 本地 mock Embedding 服务（OpenAI 兼容 /v1/embeddings）。

为何存在：
  apps/pipeline/corekg_chunk 的向量化（tools/llm_chat.py 的 chat_with_embedding）通过
  openai.AsyncOpenAI 客户端调用 `client.embeddings.create(model=..., input=text)`，
  读取 `resp.data[0].embedding`。本地没有真实向量模型时需要它提供兼容端点，
  否则"拆 chunk -> 向量化 -> ES 入库"链路会因无法调用 Embedding 而中断。

默认行为：返回一个固定维度的确定性伪向量（基于输入文本哈希，保证同一文本向量稳定）。
这样 ES 的 cosineSimilarity 可算、chunk 能入库、RAG 检索能闭环（相似度低但可用）。

用法：
  python3 scripts/mock_embedding.py --port 8031 --dim 1024
  # 兼容 docker-compose.pipeline.yml 中 `mock-llm` 服务的启动方式（--port 8031，/testdata 挂载）

验证：
  curl -s -X POST http://localhost:8031/v1/embeddings \
    -H "Content-Type: application/json" \
    -d '{"model":"mock","input":"你好"}' -H "Authorization: Bearer x"
"""
import argparse
import hashlib
import json
import logging
import random

from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

log = logging.getLogger("mock-embedding")


class EmbeddingHandler(BaseHTTPRequestHandler):
    dim = 768  # 实际值在 main() 中覆盖
    server_version = "MockEmbedding/1.0"

    def _send_json(self, code: int, obj: dict) -> None:
        body = json.dumps(obj, ensure_ascii=False).encode("utf-8")
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def _route(self) -> bool:
        parsed = self.path.split("?")[0].rstrip("/")
        if parsed == "/health":
            self._send_json(200, {"status": "ok"})
            return True
        if parsed == "/v1/embeddings":
            return False
        self._send_json(404, {"error": f"not found: {self.path}"})
        return True

    def do_GET(self):  # noqa: N802
        if self._route():
            return

    def do_POST(self): # noqa: N802
        try:
            if self._route():
                return
            length = int(self.headers.get("Content-Length", 0))
            raw = self.rfile.read(length) if length else b"{}"
            payload = json.loads(raw or b"{}")
            model = payload.get("model", "mock")
            inp = payload.get("input", "")
            texts = inp if isinstance(inp, list) else [inp]
            texts = [t if isinstance(t, str) else json.dumps(t, ensure_ascii=False) for t in texts]
            data = [
                {"index": i, "object": "embedding", "embedding": self._pseudo_vector(text)}
                for i, text in enumerate(texts)
            ]
            self._send_json(200, {
                "id": "embd-mock",
                "object": "list",
                "created": 0,
                "model": model,
                "data": data,
            })
        except Exception as exc:  # noqa: BLE001
            log.exception("mock embedding error")
            self._send_json(500, {"error": str(exc)})

    def _pseudo_vector(self, text: str):
        """确定性伪向量：同一文本返回同一向量，不同文本大概率不同。"""
        seed = int(hashlib.sha256(text.encode("utf-8")).hexdigest()[:8], 16)
        rng = random.Random(seed)
        vec = [rng.uniform(-1, 1) for _ in range(self.dim)]
        norm = (sum(v * v for v in vec) ** 0.5) or 1.0
        return [v / norm for v in vec]

    def log_message(self, fmt: str, *args):  # noqa: A003
        log.info("%s - %s", self.address_string(), fmt % args)


def main() -> None:
    parser = argparse.ArgumentParser(description="CoreKG mock embeddings server")
    parser.add_argument("--port", type=int, default=8031)
    parser.add_argument("--dim", type=int, default=768)
    parser.add_argument("--host", default="0.0.0.0")
    args = parser.parse_args()

    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
    EmbeddingHandler.dim = args.dim
    server = ThreadingHTTPServer((args.host, args.port), EmbeddingHandler)
    log.info("mock-embedding listening on %s:%s (dim=%s)", args.host, args.port, args.dim)
    server.serve_forever()


if __name__ == "__main__":
    main()
