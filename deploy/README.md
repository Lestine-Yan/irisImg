# irisImg 部署包

个人图床 irisImg 的 Linux 部署包。后端 Go 单二进制 + 前端 SPA，Nginx 同域反代。

## 目录结构

```
irisImg/
├── irisImg                # 后端二进制（linux/amd64）
├── config/config.yaml.example    # 配置模板（首次部署 cp 为 config.yaml 再改，勿直接用）
├── web/                          # 前端 SPA 静态产物
├── data/                         # 运行时数据（数据库 + 图片，运行时生成）
├── nginx/                        # Nginx 示例配置模板
│   ├── irisImg.conf.example          # HTTPS 生产形态
│   └── irisImg.http.conf.example     # HTTP 最简版
└── scripts/
    ├── start.sh               # nohup 启动
    ├── stop.sh                # 停止
    └── irisImg.service        # systemd 单元
```

## 前置要求

- Linux x86_64
- Nginx（推荐 1.18+）
- （可选）certbot 申请 HTTPS 证书

## 部署步骤

### 1. 放置目录

解压到 `/opt/irisImg`（Nginx 配置与 systemd 默认指向此路径）：

```bash
sudo mkdir -p /opt && sudo tar -xzf irisImg-0.1.0-linux-amd64.tar.gz -C /opt
```

### 2. 修改配置

release 包只随附 `config.yaml.example` 模板，**首次部署需复制为 `config.yaml` 再编辑**（更新解压只覆盖 `.example`，不会动你改好的 `config.yaml`）：

```bash
cd /opt/irisImg && cp config/config.yaml.example config/config.yaml
```

编辑 `/opt/irisImg/config/config.yaml`，**务必修改**：

- `auth.username` / `auth.password`：登录后台的账号密码（release 模式下默认值 `admin`/`admin123` 会被启动校验拒绝，必须改）
- `auth.jwt.secret`：换成 32 位以上随机字符串（默认占位串同样会被启动校验拒绝）
- `server.host`：Nginx 同机反代保持 `127.0.0.1`，避免后端 8080 直接暴露公网
- `apikey.https_only`：HTTPS 部署保持 `true`；纯 HTTP 临时测试改 `false`

### 3. 启动后端

二选一：

**systemd（生产推荐）**：

```bash
# 创建服务用户并赋权（service 指定 User=irisimg，需对 data/ 可写以落库/落盘）
sudo useradd -r -s /usr/sbin/nologin irisimg
sudo chown -R irisimg:irisimg /opt/irisImg
sudo cp /opt/irisImg/scripts/irisImg.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now irisImg
sudo journalctl -u irisImg -f   # 看日志
```

**nohup 脚本**：

```bash
cd /opt/irisImg && bash scripts/start.sh
# 停止：bash scripts/stop.sh
```

`start.sh` 启动后会轮询进程存活约 2s：若后端因配置不安全（默认口令 / 弱 JWT 密钥 / 通配 CORS / 非法 `trusted_proxies`）或 DB 打不开等 fail-closed 退出，脚本自动回显 `data/server.log` 末尾并 `exit 1`，不再误报 "started"。

验证：`curl http://127.0.0.1:8080/api/v1/ping` 应返回 pong。

### 4. 配置 Nginx

**HTTPS（生产）**：

```bash
sudo cp /opt/irisImg/nginx/irisImg.conf.example /etc/nginx/conf.d/irisImg.conf
# 编辑：全局替换 your-domain.com 为你的域名，确认证书路径
sudo nginx -t && sudo systemctl reload nginx
```

**HTTP（临时测试）**：

```bash
sudo cp /opt/irisImg/nginx/irisImg.http.conf.example /etc/nginx/conf.d/irisImg.conf
sudo nginx -t && sudo systemctl reload nginx
```

访问 `https://your-domain.com`，用配置的账号密码登录。

## 路径约定

- 后端工作目录 = release 根目录（`/opt/irisImg`），`config/data/imgs` 均为相对路径
- Nginx `/api/` -> 后端 `127.0.0.1:8080`（保留 /api 前缀）
- Nginx `/imgs/` -> `/opt/irisImg/data/imgs/`（与 `storage.root_dir` 一致）；**带图片扩展名白名单**（仅放行 `.png/.jpg/.jpeg/.gif/.webp`，与后端 `serveImages` 对齐）：即便 `root_dir` 误配或目录混入非图片文件，也无法经 `/imgs/` 下载 `.yaml/.db/.go` 等敏感文件

## 升级

整包解压覆盖即可，`config/config.yaml` 与 `data/` 不在 release 包内、不会被覆盖：

```bash
sudo tar -xzf irisImg-<新版本>-linux-amd64.tar.gz -C /opt
sudo systemctl restart irisImg   # nohup 方式则 bash scripts/stop.sh && bash scripts/start.sh
```

注意：整包解压会覆盖 `/opt/irisImg` 下的二进制、`web/`、`scripts/`（start.sh/stop.sh/irisImg.service）、`nginx/*.example` 模板与 `README.md`；若曾本地改动过 `scripts/` 下脚本，升级前请备份。`/etc/systemd/system/irisImg.service` 与 `/etc/nginx/conf.d/irisImg.conf` 是首次部署时单独 cp 出去的、不在 release 包内，不会被覆盖；若新版 `scripts/irisImg.service` 或 `nginx/*.example` 有需采纳的修复，手动重新 cp 并执行 `systemctl daemon-reload` / `nginx -t && systemctl reload nginx`。新版本若在 `config.yaml.example` 新增配置键，升级后请 `diff config/config.yaml config/config.yaml.example` 合并，否则缺失键在运行时取零值（关键键缺失会导致启动失败）。

## 排错

| 现象 | 排查 |
| --- | --- |
| 启动失败 | nohup：`start.sh` 启动后探测进程存活，若启动期 fail-closed（默认口令 / 弱 JWT 密钥 / 通配 CORS / 非法 `trusted_proxies` / DB 打不开等）退出，自动回显 `data/server.log` 末尾并 `exit 1`；systemd：`journalctl -u irisImg -n 100` |
| systemctl 反复重启 / failed | 多半漏了步骤 2：确认 `config/config.yaml` 已从 `.example` 复制并改密码/密钥（release 模式默认值会被启动校验拒绝）；`systemctl reset-failed irisImg && systemctl restart irisImg` |
| 图片 404 | 确认 Nginx `/imgs/` alias 路径与 `storage.root_dir` 一致；或请求的末段扩展名不在白名单（`.yaml/.db/.go`/无扩展名/目录均 404，这是 `/imgs/` 扩展名白名单的纵深防御） |
| 图片 403 | 落盘文件权限/属主：新版已 0644；旧版本 0600 历史文件需 `find /opt/irisImg/data/imgs -type f -exec chmod 644 {} \;` 批量补权限；确认 Nginx worker（`www`）可遍历目录（0755） |
| 图片 src 形如 `/img.example.com/imgs/...` | `public_base_url` 配了无协议裸域名；改为 `https://img.example.com` 或留空，并升级到带「自动补协议」的版本 |
| 密钥接口 403 | HTTPS 模式下 `https_only=true` 需 Nginx 透传 `X-Forwarded-Proto` |
| 上传 413 | 调大 Nginx `client_max_body_size`（> `max_upload_size_mb`） |
| SPA 刷新 404 | Nginx `location /` 需 `try_files $uri $uri/ /index.html` |
