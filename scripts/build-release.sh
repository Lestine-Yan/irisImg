#!/usr/bin/env bash
# 一键构建 irisImg release 部署包（linux/amd64）。
# 产物：dist/irisImg/ 目录 + dist/irisImg-<version>-linux-amd64.tar.gz
#
# 用法：  bash scripts/build-release.sh
# 覆盖版本号：VERSION=0.2.0 bash scripts/build-release.sh
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

VERSION="${VERSION:-0.1.0}"
ARCH="amd64"
DIST="$ROOT_DIR/dist"
PKG="$DIST/irisImg"
TARBALL="$DIST/irisImg-${VERSION}-linux-${ARCH}.tar.gz"

echo "==> 清理旧的构建产物"
rm -rf "$PKG"
mkdir -p "$PKG"/{config,web,data,nginx,scripts}

echo "==> 交叉编译后端 (linux/${ARCH})"
cd "$ROOT_DIR/backend"
CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build -trimpath -ldflags="-s -w" -o "$PKG/irisImg" ./cmd/server
# Git Bash / Windows 构建时 go build 产出的二进制不带 Unix 执行位，
# 必须显式补 +x，否则 Linux 解压后无法执行（nohup Permission denied / systemd 203/EXEC）。
# 注:随包脚本(start.sh/stop.sh)的执行位由打包时 tar --mode='0755' 统一强制,无需此处 chmod。
chmod +x "$PKG/irisImg"

echo "==> 构建前端 SPA (apiBase=/api/v1)"
cd "$ROOT_DIR/frontend"
pnpm install --frozen-lockfile
# Git Bash (MSYS2) 会把以 / 开头的环境变量值转换成 Windows 绝对路径，污染
# NUXT_PUBLIC_API_BASE（/api/v1 -> D:/.../Git/api/v1），部署后浏览器请求非法 URL
# 而 Failed to fetch。MSYS_NO_PATHCONV=1 关闭该转换；nuxt.config.ts 另有盘符兜底，双重保险。
# （Linux 构建机上无 MSYS_NO_PATHCONV 此变量，bash 自动忽略，无副作用。）
MSYS_NO_PATHCONV=1 NUXT_PUBLIC_API_BASE=/api/v1 pnpm generate
cp -r "$ROOT_DIR/frontend/.output/public/." "$PKG/web/"

# 校验产物 apiBase 未被 MSYS 污染：产物里不应出现「盘符: 路径 ... api/v1」的 Windows 绝对路径。
# 正则锚定「单字母盘符 + 冒号 + 单斜杠 + 非斜杠字符」，以排除 https:// 这类合法 URL 协议。
if grep -rEq '[A-Za-z]:[\\/][^\\/"][^"]{0,59}api[\\/]v1' "$PKG/web/" 2>/dev/null; then
  echo "!! 前端产物检测到疑似被 MSYS 污染的 apiBase（盘符路径），构建中止" >&2
  grep -rEno '[A-Za-z]:[\\/][^\\/"][^"]{0,59}api[\\/]v1' "$PKG/web/" 2>/dev/null | head -5 >&2
  exit 1
fi
echo "    apiBase 校验通过（无 MSYS 路径污染）"

echo "==> 拷贝部署模板"
cd "$ROOT_DIR"
# 配置与 Nginx 模板一律带 .example 后缀随包分发，避免更新解压时覆盖用户已改好的
# 真实配置（config/config.yaml、nginx 真实 conf）。首次部署需手动 cp 去后缀并修改。
cp deploy/config.yaml.example "$PKG/config/config.yaml.example"
cp deploy/nginx/irisImg.conf.example "$PKG/nginx/"
cp deploy/nginx/irisImg.http.conf.example "$PKG/nginx/"
# 随包脚本必须 LF 行尾:Windows 开发机 git autocrlf=true 会让 working tree 的 .sh/.service
# 变成 CRLF,Linux 上 `#!/usr/bin/env bash\r` 会报 "/usr/bin/env: 'bash\r': No such file or directory"。
# 源文件已由 .gitattributes 钉为 LF,此处拷贝时再剥一次 CR 兜底,防意外污染。
for f in start.sh stop.sh irisImg.service; do
  sed 's/\r$//' "deploy/scripts/$f" > "$PKG/scripts/$f"
done
cp deploy/README.md "$PKG/README.md"
touch "$PKG/data/.gitkeep"

echo "==> 打包 tar.gz"
cd "$DIST"
# Git Bash 下 chmod 对 NTFS 无效（二进制 stat 仍是 0644），无法靠 chmod 补执行位；
# 改用 tar --mode 在打包时强制条目权限为 0755，确保 Linux 解压后二进制可执行。
tar -czf "$TARBALL" --mode='0755' irisImg

echo "==> 完成: $TARBALL"
echo "    解压后根目录: irisImg/  (后端工作目录 cwd)"
