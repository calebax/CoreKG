export interface AlgoSplitOptions {
  uin: string;
  companyId: string;
  forestId: string;
  fileId: string;
  content: string;
  esIndex: string;
  fileExt: string;
}

export async function algoSplit(
  algoUrl: string,
  opts: AlgoSplitOptions,
): Promise<string> {
  const body = new URLSearchParams({
    uin: opts.uin,
    company_id: opts.companyId,
    forest_id: opts.forestId,
    file_id: opts.fileId,
    content: opts.content,
    es_index: opts.esIndex,
    file_ext: opts.fileExt,
  });

  const resp = await fetch(`${algoUrl}/split`, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: body.toString(),
  });
  if (!resp.ok) throw new Error(`Algo split error: ${resp.status}`);
  return resp.text();
}

export interface AlgoIndexOptions {
  uin: string;
  companyId: string;
  forestId: string;
  fileId: string;
  esIndex: string;
}

export async function algoIndex(
  algoUrl: string,
  opts: AlgoIndexOptions,
): Promise<string> {
  const body = new URLSearchParams({
    uin: opts.uin,
    company_id: opts.companyId,
    forest_id: opts.forestId,
    file_id: opts.fileId,
    es_index: opts.esIndex,
  });

  const resp = await fetch(`${algoUrl}/index`, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: body.toString(),
  });
  if (!resp.ok) throw new Error(`Algo index error: ${resp.status}`);
  return resp.text();
}
