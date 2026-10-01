# NightReaper AI 更新日志

## v1.0.0-nightreaper（2026-10-01）

基于上游 CyberStrikeAI v1.7.20 的深度定制版本。

### 品牌（55 处）
- 全平台 `CyberStrikeAI` → `NightReaper AI`（UI、登录页、API 文档、终端欢迎语、system prompt、i18n 中英文全树）
- 系统 prompt 声明身份：NightReaper AI 由 yyyr 独立打造
- 主页 ASCII logo / 图标：兜帽黑客动漫人物（荧光绿眼 + 代码雨 + 镰刀）
- 全站配色：原「AI 蓝 + 紫渐变」→ 荧光绿黑客风（400+ 颜色点重新映射）
- Web 端开机动画：矩阵雨 + NIGHTREAPER 渐现 + ACCESS GRANTED（每会话一次，可点击跳过）

### 新增
- 全局模型调用并发闸（默认 4 路排队，流式响应持坑到 EOF/Close，环境变量 `NRAI_MODEL_MAX_CONCURRENCY` 可调）——根除余额不足时 DeepSeek 并发限制引发的 429 风暴
- 19 个火力全开攻击角色（全面击破 / 内网杀手 / 密码爆破机 / 提权专家 / C2 指挥官 / 免杀投递手 + 13 个重写的基础角色）
- 反向 Shell MCP 服务接入（开监听 / 接 shell / 远程执行，自动批准）
- 关于作者弹窗 + GitHub 对接

### 修复
- Token 用量页 500（原作按天分组 SQL 未处理 NULL 行）
- config.yaml 缺 `roles_dir` 导致角色页面永远为空
- GitHub 外链损坏（URL 含空格）
- 登录页 / 授权弹窗 / API 文档残留英文品牌文案（中英文 i18n 32 处）

### 工程化
- SQLite 驱动 `mattn/go-sqlite3`（cgo）→ `modernc.org/sqlite`（纯 Go）：全平台编译不再需要 gcc
- 桌面一键启停启动器（黑客风开机 / 红色终止动画，自动拉起服务并开浏览器）
