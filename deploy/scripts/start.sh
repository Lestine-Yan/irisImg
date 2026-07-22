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
pid=$!
echo "$pid" > data/irisImg.pid

# 启动存活探测：后端启动期 fail-closed（配置读/解析失败、release 模式默认口令/弱 JWT 密钥/
# 通配 CORS、非法 server.trusted_proxies、DB 打不开等）会 log.Fatalf + Exit 1。nohup & 异步
# 启动，此处轮询最多 ~2s：进程在此期间退出则回显日志尾部并 exit 1，避免误报 "started"
# 的假成功窗口（错误原本只进 data/server.log，终端看不到）。
for _ in $(seq 1 20); do
  kill -0 "$pid" 2>/dev/null || {
    echo "错误：irisImg 启动失败（进程已退出），日志末尾如下：" >&2
    tail -n 30 data/server.log >&2
    rm -f data/irisImg.pid
    exit 1
  }
  sleep 0.1
done
echo "irisImg started, pid=$pid, log=data/server.log"
