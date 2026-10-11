# 部署与 Release 打包

本文档说明 irisImg 的部署架构、release 打包流程，以及 Nginx 反代约定。各源文件逐文件文档见各自目录；图片反代细节另见 [`IMAGE.md`](./backend/IMAGE.md)。

## 部署架构

irisImg 为前后端分离的个人图床：

- **后端**：Go(Gin) 单二进制，纯 Go SQLite 驱动（`modernc.org/sqlite`，无需 CGO），监听 `0.0.0.0:8080`。业务接口统一挂在 `/api/v1` 下，图片 / 音频静态服务分别挂在 `/imgs`、`/audios`。
- **前端**：Nuxt4 SPA（`ssr:false`），`pnpm generate` 产出纯静态站点，由 Nginx 直接 serve。
- **反代**：Nginx 同域反代，`/api/` 转发后端、`/imgs/` 与 `/audios/` 直接 serve 落盘目录、`/` serve 前端 SPA。

## Release 打包

打包脚本：[`scripts/build-release.sh`](../scripts/build-release.sh)，产物输出到 `dist/`（已 gitignore）。

### GitHub Actions 自动发布

推送 `v*` tag 会触发 [`.github/workflows/release.yml`](../.github/workflows/release.yml)，在 ubuntu-latest 上自动「测试 + 构建 + 发布 GitHub Release」：

```bash
# 本地发布流程（发布说明写在 release commit 的 body 里）
git commit -m "chore(release): v0.1.1" -m "## 新增
- xxx

## 修复
- xxx"
git tag v0.1.1
git push origin <branch> v0.1.1
```

链路与约定：

1. **触发**：仅 `git push origin v<x.y.z>` 形式的 tag 触发，普通 push 不受影响，完全可控。
2. **发布门禁**：先跑 `go vet ./...` + `go test ./...`（后端），任一失败即中止，不发布。前端无独立测试，`pnpm generate` 构建本身即前端门禁（含 MSYS 污染 grep 校验）。
3. **构建**：复用 `scripts/build-release.sh`，版本号取自 tag（`v0.1.1` → `0.1.1`）。ubuntu-latest 上无 MSYS 路径转换 / CRLF 行尾问题，脚本内的 Windows 防护均为无副作用的 no-op。
4. **发布说明**：取 tag 指向 commit 的 **body**（`git log -1 --format=%b`），markdown 原样渲染为 Release 正文；body 为空则 workflow 直接报错中止，不会发出无说明的 Release。说明风格参考 PR squash commit（如 `9586f98`）的分段格式。
5. **产物**：`dist/irisImg-<version>-linux-amd64.tar.gz` 作为 Release Asset 上传（`fail_on_unmatched_files` 保证产物缺失时发布失败）。

工具链：Go 版本读 `backend/go.mod`，pnpm 10 + Node 22（`frontend/package.json` 未声明 `packageManager`，版本钉在 workflow 里）。

### 产物结构

```
irisImg/
├── irisImg                       # 后端二进制（linux/amd64，CGO_ENABLED=0 交叉编译）
├── config/config.yaml.example    # 配置模板（首次部署 cp 为 config.yaml 再改；更新只覆盖 .example）
├── web/                          # 前端 SPA（frontend/.output/public 产物）
├── data/                         # 运行时数据（db + imgs + audios，运行时生成）
├── nginx/                        # Nginx 示例配置模板（.conf.example，HTTPS 生产 + HTTP 最简）
└── scripts/                      # start.sh / stop.sh / irisImg.service
```

最终压缩为 `dist/irisImg-<version>-linux-amd64.tar.gz`，解压后根目录即后端工作目录（cwd），`config/data/imgs` 全用相对路径。配置随附 `config.yaml.example` 模板，首次部署需 `cp config/config.yaml.example config/config.yaml` 并改密码/密钥后即可运行（release 模式下默认口令/密钥会被 `config.Validate` 拒绝启动）。后续整包解压会覆盖二进制、`web/`、`scripts/`、`nginx/*.example` 等随包文件，但 `config/config.yaml` 与 `data/` 不在 release 包内、不会被覆盖；若新版本在 `.example` 新增配置键或调整默认值，升级后请 `diff config/config.yaml config/config.yaml.example` 合并新增键，否则缺失键在运行时取零值（关键键如 `storage.root_dir` 缺失会导致启动失败）。

### 构建流程

1. **交叉编译后端**：`CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o irisImg ./cmd/server`。纯 Go 驱动无需 CGO，可直接交叉编译。
2. **前端 SPA**：`MSYS_NO_PATHCONV=1 NUXT_PUBLIC_API_BASE=/api/v1 pnpm generate`，产物 `frontend/.output/public/` 拷入 `web/`。`MSYS_NO_PATHCONV=1` 用于规避 Git Bash (MSYS2) 路径转换污染（见下文「Git Bash 构建坑」），Linux 构建机上无副作用；脚本另对产物做 grep 校验，命中盘符路径则中止。
3. **组装**：把 `deploy/` 下的配置/脚本/文档模板拷入 `dist/irisImg/` 对应位置。配置与 Nginx 模板一律以 `.example` 后缀随包分发（`config.yaml.example`、`irisImg.conf.example`、`irisImg.http.conf.example`），避免更新解压覆盖用户真实配置。
4. **打包**：`tar -czf dist/irisImg-<version>-linux-amd64.tar.gz irisImg`。

### 同域 apiBase 约定

前端 `apiBase` 在构建时由 `NUXT_PUBLIC_API_BASE` 注入为 `/api/v1`（相对路径），故 **必须同域部署**：Nginx `/api/` 反代到后端并保留 `/api` 前缀。如需独立后端域名，需改 `frontend/nuxt.config.ts` 的 `apiBase` 为绝对地址并重新打包。

### Git Bash (MSYS2) 构建坑

在 Windows 上用 Git Bash 执行 `NUXT_PUBLIC_API_BASE=/api/v1 pnpm generate` 时，MSYS2 会把传给原生 Windows 程序（`node.exe`）的环境变量值做 POSIX->Windows 路径转换：`/api/v1` 被改写成 `D:/Runer/Git/Git/api/v1`（Git 安装根即 MSYS 根）并烤进产物。部署后浏览器请求该非法 URL 直接 `Failed to fetch`，请求不会到达后端，后端日志里看不到 `/api/v1/auth/login`，只剩公网扫描器产生的无关 404（如 `/api/env`、`/api/users/admin/check`）。

防护（`scripts/build-release.sh` 已内置，三层）：

1. 构建命令前加 `MSYS_NO_PATHCONV=1` 关闭转换。（`MSYS2_ARG_CONV_EXCL="*"` 只管命令行参数、对环境变量值无效，勿用。）
2. `frontend/nuxt.config.ts` 的 `sanitizeApiBase()` 对盘符开头的值兜底重置为 `/api/v1`（应用层，跨平台稳定）。
3. 构建后 grep 产物校验「盘符路径 + api/v1」，命中即中止打包。

手动构建的等价规避：在 PowerShell / cmd 里执行（不做路径转换），或 `MSYS_NO_PATHCONV=1 NUXT_PUBLIC_API_BASE=/api/v1 pnpm generate`。

### 随包脚本 CRLF 行尾坑

Windows 开发机 `git core.autocrlf=true` 会让 checkout 出来的 `.sh` / `.service` 变成 CRLF 行尾。若直接打包到 Linux，`#!/usr/bin/env bash\r` 里的 `\r` 被当作命令名的一部分，执行报 `/usr/bin/env: 'bash\r': No such file or directory`；`systemd` 加载 `.service` 也可能因 CRLF 解析异常。

防护（`scripts/build-release.sh` 已内置，三层）：

1. 仓库根 `.gitattributes` 强制 `*.sh` / `*.service` 在 checkout 时保持 `eol=lf`，不受 `autocrlf` 影响。
2. `build-release.sh` 拷贝 `deploy/scripts/{start,stop}.sh` 与 `irisImg.service` 进包时用 `sed 's/\r$//'` 剥 CR，兜底防源文件意外被污染。
3. 打包时 `tar --mode='0755'` 统一强制执行位（脚本与二进制都覆盖）。

手动排查：`file deploy/scripts/*.sh` 若报 `with CRLF line terminators` 即中招；`sed -i 's/\r$//' <file>` 修复。

## Nginx 反代约定

- `location /api/`：`proxy_pass http://127.0.0.1:8080;`（**无尾斜杠**，保留 `/api` 前缀），透传 `X-Forwarded-Proto` 供后端 `apikey.https_only` 校验；`proxy_read_timeout` 放宽到 300s（100MB 音频经慢速上行链路上传可能超 60s）。
- `location /imgs/`：`alias <root>/data/imgs/;`，路径须与 `storage.root_dir` 一致。落盘文件权限须为 0644（新版 `storage.Saver` 已自动 `chmod`），目录 0755，且 Nginx worker 用户（如 `www`）需对 `<root>/data/imgs/` 及其子目录有可遍历位；旧版本落盘为 0600，升级后历史文件需 `find <root>/data/imgs -type f -exec chmod 644 {} \;` 批量补权限，否则跨用户读取 403。
- `location /audios/`：`alias <root>/data/audios/;`，路径须与 `storage.audio_root_dir` 一致，扩展名白名单为 mp3/m4a/flac + 封面图片扩展名（内嵌封面落 `covers/` 子目录，与音频共用前缀）；权限要求同 `/imgs/`。Nginx 原生支持 Range 请求，音频拖动进度条无需额外配置。
- `location /`：`try_files $uri $uri/ /index.html;` 兜底 SPA 路由刷新。
- `client_max_body_size` 需大于 `storage.audio_max_upload_size_mb`（默认 100MB，Nginx 示例取 150m）；图片上传仍由后端 `max_upload_size_mb=20` 精确限制。

示例配置见 `deploy/nginx/irisImg.conf.example`（HTTPS 生产形态，仅监听 443；80->443 跳转不在该模板内，需由主站 Nginx 或宝塔另行处理，见模板内注释）与 `deploy/nginx/irisImg.http.conf.example`（HTTP 最简版）；release 包内同名 `.example` 模板，首次部署 `cp` 去后缀使用。

## 配置要点（生产）

- `server.mode: release`；`server.trusted_proxies` 默认本地回环（同机反代），跨机反代需追加反代所在 CIDR（HTTPSOnly 仅对来自此列表的请求认 `X-Forwarded-Proto`）
- `auth.username` / `auth.password` / `auth.jwt.secret` 必改
- `apikey.https_only: true`（后端据 `server.trusted_proxies` + `X-Forwarded-Proto` 校验 HTTPS；Nginx 仍建议 `proxy_set_header X-Forwarded-Proto $scheme` 覆盖该头作纵深防御）
- `cors.allow_origins` 留空（同域部署无跨域需求，最安全）；`*` 会被 release 模式 `Validate` 拒绝启动
- `storage.root_dir` 与 Nginx `/imgs/` alias 指向同一物理目录，落盘文件 0644、目录 0755 保证 Nginx 跨用户可读
- `storage.public_base_url` 留空（同域 `/imgs/` 反代，最简）或填带协议的绝对前缀（如 `https://img.example.com`）；裸域名会被自动补 `https://`，但不建议依赖
- **音频升级注记**：v0.2 起新增 `storage.audio_*` 配置键与 `/audios/` Nginx location。**旧 config.yaml 缺这些键时 `ApplyDefaults` 自动补默认值（`data/audios` / 100MB / mp3+m4a+flac），升级不会启动失败**；但 Nginx 配置需手动加上 `/audios/` location（拷贝新 `.example` 模板或对照补充），否则音频直链 404。`data/audios/` 目录由后端启动时自动创建。

部署清单与排错见随 release 包分发的 `deploy/README.md`。
