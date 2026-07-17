#!/usr/bin/env bash
# irisImg 后台启动脚本（nohup 方式）。
# 生产建议改用 systemd（见 scripts/irisImg.service）。
set -euo pipefail

# 切到 release 根目录（本脚本位于 scripts/ 下）
cd "$(dirname "$0")/.."

# 若以 root 调用,降权到专用服务用户 irisimg(与 systemd User=irisimg 对齐),
# 避免后端以 root 运行(落库/落盘/静态服务一旦有文件写或路径穿越漏洞即危及宿主)。
if [[ "${EUID}" -eq 0 ]]; then
  if ! id -u irisimg >/dev/null 2>&1; then
    echo "错误:以 root 运行但服务用户 irisimg 不存在。请先执行:" >&2
    echo "  sudo useradd -r -s /usr/sbin/nologin irisimg" >&2
    echo "  sudo chown -R irisimg:irisimg \"$(pwd)\"" >&2
    exit 1
  fi
  chown -R irisimg:irisimg . 2>/dev/null || true
  exec runuser -u irisimg -- "$0" "$@"
fi

export IRIS_CONFIG="${IRIS_CONFIG:-config/config.yaml}"
# release 包只随附 config.yaml.example 模板，首次部署需复制为 config.yaml 并改密码/密钥；
# 这样后续更新解压不会覆盖已改好的真实配置。此处检测并给出明确指引，避免后端读不到配置。
if [[ ! -f "$IRIS_CONFIG" ]]; then
  echo "错误：配置文件 $IRIS_CONFIG 不存在" >&2
  if [[ -f config/config.yaml.example ]]; then
    echo "首次部署请先复制并修改模板：cp config/config.yaml.example config/config.yaml" >&2
  fi
  exit 1
fi
mkdir -p data

if [[ -f data/irisImg.pid ]] && kill -0 "$(cat data/irisImg.pid)" 2>/dev/null; then
  echo "irisImg already running, pid=$(cat data/irisImg.pid)"
  exit 1
fi

nohup ./irisImg > data/server.log 2>&1 &
echo $! > data/irisImg.pid
echo "irisImg started, pid=$(cat data/irisImg.pid), log=data/server.log"
