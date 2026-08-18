import os
import cv2
import imagehash
from PIL import Image
import shutil
from loguru import logger


def process_local_video_and_generate_markdown(video_path, final_md_file, image_prefix="frame", threshold=5, frame_interval_seconds=1, *args, **kwargs):
    """
    从本地视频文件中提取关键帧并生成 Markdown 引用文件。

    关键帧提取基于感知哈希（phash）差异，按指定时间间隔采样。
    生成的图片和 Markdown 文件保存至输出目录。

    Args:
        video_path: 本地视频文件路径
        final_md_file: 输出的 Markdown 文件路径
        image_prefix: 图片 URL 前缀，用于 Markdown 中的图片引用
        threshold: 哈希差异阈值，超过此值时认定为关键帧，默认 5
        frame_interval_seconds: 帧采样间隔（秒），默认 1

    Returns:
        (成功标志, 结果信息)
    """
    try:
        markdown_file_path = final_md_file
        output_base_dir = os.path.dirname(final_md_file)
        # 确保输出目录存在
        images_folder = os.path.join(output_base_dir, "images")
        os.makedirs(images_folder, exist_ok=True)

        cap = cv2.VideoCapture(video_path)
        if not cap.isOpened():
            return False, f"无法打开视频文件：{video_path}，请检查路径是否正确。"

        fps = cap.get(cv2.CAP_PROP_FPS)
        if fps == 0:
            cap.release()
            return False, "无法获取视频帧速率，文件可能已损坏或为空。"

        frame_interval = int(fps * frame_interval_seconds)
        if frame_interval == 0:
            frame_interval = 1  # 最小采样间隔为 1 帧

        prev_hash = None
        frame_count = 0
        extracted_frames_info = []  # 记录关键帧信息 (图片相对路径, 帧号, 时间戳, 哈希差值)
        image_prefix = image_prefix.rstrip('/')

        logger.info(f"开始处理本地视频文件：{video_path}")
        logger.info(f"视频帧速率：{fps:.2f} FPS")
        logger.info(f"采样间隔：每 {frame_interval_seconds} 秒一帧")
        logger.info(f"哈希差异阈值：{threshold}")
        logger.info(f"输出文件保存到：{output_base_dir}")

        while True:
            ret, frame = cap.read()
            if not ret:
                break  # 视频结束或读取失败

            # 每隔 frame_interval 帧处理一次
            if frame_count % frame_interval == 0:
                # BGR → RGB 转换后计算感知哈希
                pil_img = Image.fromarray(cv2.cvtColor(frame, cv2.COLOR_BGR2RGB))
                current_hash = imagehash.phash(pil_img)

                hash_diff = 0
                if prev_hash is not None:
                    hash_diff = current_hash - prev_hash

                # 第一帧或哈希差异超过阈值时认定关键帧并保存
                if prev_hash is None or hash_diff > threshold:
                    image_filename = f"{frame_count:06d}.jpg"
                    image_full_path = os.path.join(images_folder, image_filename)
                    pil_img.save(image_full_path, format='JPEG')

                    # 计算时间戳（秒）
                    timestamp_seconds = frame_count / fps

                    # 构建 Markdown 引用的图片相对路径
                    relative_image_path_in_md = os.path.join(f"{image_prefix}/images", image_filename)
                    extracted_frames_info.append({
                        "path": relative_image_path_in_md,
                        "frame_number": frame_count,
                        "timestamp_seconds": timestamp_seconds,
                        "hash_difference": hash_diff
                    })
                    prev_hash = current_hash  # 更新前一帧哈希

            frame_count += 1

    except Exception as e:
        return False, f"视频处理发生错误：{e}"
    finally:
        if 'cap' in locals() and cap.isOpened():
            cap.release()  # 确保释放视频捕获对象

    if not extracted_frames_info:
        return False, "未提取到任何关键帧。"

    logger.info(f"提取了 {len(extracted_frames_info)} 个关键帧，正在生成 Markdown 文件...")

    try:
        with open(markdown_file_path, "w", encoding="utf-8") as md_file:
            for frame_info in extracted_frames_info:
                # 以 yg_pos 标记注入帧号和秒级时间戳
                md_file.write(f"![]({frame_info['path']})<!--yg_pos{frame_info['frame_number']},{int(frame_info['timestamp_seconds'])},0,0,0yg_pos-->\n\n")

        logger.info(f"Markdown 文件已生成：{markdown_file_path}")
        return True, markdown_file_path
    except Exception as e:
        return False, f"生成 Markdown 文件时发生错误：{e}"




def video_process(video_path, final_md_file, image_prefix, threshold=5, frame_interval_seconds=1, *args, **kwargs):
    logger.info(f"收到视频处理请求：路径={video_path}, 输出文件={final_md_file}, "
            f"图片前缀={image_prefix}, 阈值={threshold}, 间隔={frame_interval_seconds}")

    success, result_info = process_local_video_and_generate_markdown(
        video_path=video_path,
        final_md_file=final_md_file,
        image_prefix=image_prefix,
        threshold=threshold,
        frame_interval_seconds=frame_interval_seconds, *args, **kwargs
    )

    logger.info(f"视频处理完成，Markdown文件已生成。{success} {result_info}" )

