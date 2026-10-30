# 影像處理與點陣字型（Pillow）。版本鎖定，沿用克萊恩英豪 tools/docker/img.Dockerfile。
# 本作與銀色匕首之謎共用此檔；沿用其建出的 image `sosb-img:r1`（sha256:8b9012adb991…）。
FROM python:3.13.16-slim@sha256:cf50a4f6ac4f026458b08d5222023d700ac806410dff5e9ca8728fa44f2b850f
RUN pip install --no-cache-dir pillow==11.0.0 numpy==2.1.3
