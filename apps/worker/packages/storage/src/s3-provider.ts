import { S3Client } from "@aws-sdk/client-s3";
import { Upload } from "@aws-sdk/lib-storage";
import { createReadStream } from "node:fs";
import { readdir, stat, writeFile, mkdir } from "node:fs/promises";
import { join, relative, basename } from "node:path";
import mime from "mime-types";
import type { z } from "zod";
import type { S3ConfigSchema } from "@corekg/config";

type S3Config = z.infer<typeof S3ConfigSchema>;

export interface UploadResult {
  url: string;
  key: string;
}

export interface StorageProvider {
  downloadFile(url: string, destDir: string, filename?: string): Promise<string>;
  uploadFile(localPath: string, key: string, bucket?: string): Promise<UploadResult>;
  uploadDirectory(localDir: string, basePath: string, bucket?: string): Promise<UploadResult[]>;
  getEndpoint(): string;
}

export function createS3Provider(config: S3Config): StorageProvider {
  const client = new S3Client({
    endpoint: config.endpointUrl,
    credentials: {
      accessKeyId: config.accessKeyId,
      secretAccessKey: config.secretAccessKey,
    },
    region: config.region,
    forcePathStyle: true,
  });
  const bucket = config.defaultBucket;
  const pubEndpoint = config.publicEndpointUrl || config.endpointUrl;

  return {
    getEndpoint() {
      return config.endpointUrl;
    },

    async downloadFile(url, destDir, filename?) {
      await mkdir(destDir, { recursive: true });
      const name = filename || basename(new URL(url).pathname);
      const dest = join(destDir, name);
      const resp = await fetch(url);
      if (!resp.ok) throw new Error(`Download failed: ${resp.status}`);
      await writeFile(dest, Buffer.from(await resp.arrayBuffer()));
      return dest;
    },

    async uploadFile(localPath, key, overrideBucket?) {
      const ct = mime.lookup(localPath) || "application/octet-stream";
      await new Upload({
        client,
        params: {
          Bucket: overrideBucket || bucket,
          Key: key,
          Body: createReadStream(localPath),
          ContentType: ct,
        },
      }).done();
      return { url: `${pubEndpoint}/${overrideBucket || bucket}/${key}`, key };
    },

    async uploadDirectory(localDir, basePath, overrideBucket?) {
      const results: UploadResult[] = [];
      const entries = await readdir(localDir, { recursive: true });
      for (const entry of entries) {
        const full = join(localDir, entry as string);
        const s = await stat(full);
        if (!s.isFile()) continue;
        const key = `${basePath}/${relative(localDir, full)}`;
        results.push(await this.uploadFile(full, key, overrideBucket));
      }
      return results;
    },
  };
}
