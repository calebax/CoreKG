import yaml
from loguru import logger 
from openai import AsyncOpenAI

# ===================== 加载配置 =====================
with open("./config/chunk_config.yaml", "r", encoding="utf-8") as file:
    config = yaml.safe_load(file)

# ===================== LLM 调用 =====================
async def chat_with_llm(prompt: str,
                   model: str,
                   api_key: str,
                   base_url: str,
                   system_prompt: str | None = None,
                   history: list[dict] | None = None,
                   temperature: float = 1,
                   **kwargs) -> str:
    """单次调用 LLM 接口"""
    history = history or []

    client = AsyncOpenAI(api_key=api_key, base_url=base_url, timeout=config['Concurrency']['LLM_TIMEOUT'])

    messages = []
    if system_prompt:
        messages.append({"role": "system", "content": system_prompt})
    if history:
        messages.extend(history)
    messages.append({"role": "user", "content": prompt})

    try:
        resp = await client.chat.completions.create(
            model=model,
            messages=messages,
            temperature=temperature,
            extra_body={"enable_thinking":False,
                        "chat_template_kwargs": {"enable_thinking": False,
                                                 "budget_thoughts_token": 50}},
            **kwargs,
        )
        result = resp.choices[0].message.content
        logger.info(f"LLM 调用成功, model={model}, prompt_length={len(prompt)}, response={result}")
        return result
    except Exception as e:
        logger.error(f"LLM 调用失败: {e}")
        raise Exception(f"LLM 调用失败: {e}")


# ===================== Embedding 调用 =====================
async def chat_with_embedding(text: str,
                         model: str,
                         api_key: str,
                         base_url: str) -> list[float] | None:
    """单次调用 Embedding 接口"""
    if not text.strip():
        return None

    client = AsyncOpenAI(api_key=api_key, base_url=base_url, timeout=config['Concurrency']['LLM_TIMEOUT'])

    try:
        resp = await client.embeddings.create(
            model=model,
            input=text,
        )
        return resp.data[0].embedding
    except Exception as e:
        logger.error(f"Embedding 调用失败: {e}")
        raise Exception(f"Embedding 调用失败: {e}")
