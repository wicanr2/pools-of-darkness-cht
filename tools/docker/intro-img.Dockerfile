# 片頭動畫材質對位；r2 沿用 img.Dockerfile 的 Python、Pillow、NumPy 版本。
# 建置：docker build -f tools/docker/intro-img.Dockerfile -t pod-intro-img:r2 .
FROM python:3.13.16-slim@sha256:cf50a4f6ac4f026458b08d5222023d700ac806410dff5e9ca8728fa44f2b850f
RUN pip install --no-cache-dir pillow==11.0.0 numpy==2.1.3 \
    'opencv-python-headless @ https://files.pythonhosted.org/packages/89/53/e19c21e0c4eb1275c3e2c97b081103b6dfb3938172264d283a519bf728b9/opencv_python_headless-4.12.0.88-cp37-abi3-manylinux2014_x86_64.manylinux_2_17_x86_64.whl#sha256=236c8df54a90f4d02076e6f9c1cc763d794542e886c576a6fee46ec8ff75a7a9'
