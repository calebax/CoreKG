export interface DaemonPdfOptions {
  filePath: string;
  targetDir: string;
  publicPath: string;
}

export async function daemonProcessPdf(
  daemonUrl: string,
  opts: DaemonPdfOptions,
): Promise<void> {
  const resp = await fetch(daemonUrl, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      pdf_file_name: opts.filePath,
      target_path: opts.targetDir,
      public_path: opts.publicPath,
    }),
  });
  if (!resp.ok) throw new Error(`Daemon PDF error: ${resp.status}`);
}

export interface DaemonVideoOptions {
  videoPath: string;
  outputDir: string;
  imagePrefix: string;
}

export async function daemonProcessVideo(
  daemonUrl: string,
  opts: DaemonVideoOptions,
): Promise<void> {
  const resp = await fetch(daemonUrl, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      video_path: opts.videoPath,
      output_base_dir: opts.outputDir,
      image_prefix: opts.imagePrefix,
    }),
  });
  if (!resp.ok) throw new Error(`Daemon video error: ${resp.status}`);
}
