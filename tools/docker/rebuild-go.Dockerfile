# 與已驗Go工具鏈相同的1.24.13系列。既有映像可沿用；新機器由此建置。
FROM golang:1.24.13-bookworm@sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac
RUN apt-get update && apt-get install -y --no-install-recommends \
    libx11-dev libxrandr-dev libxcursor-dev libxinerama-dev libxi-dev libxxf86vm-dev \
    libgl1-mesa-dev libasound2-dev pkg-config xvfb xauth libgl1 libglx-mesa0 \
    && rm -rf /var/lib/apt/lists/*
