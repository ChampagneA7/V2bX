# v2026.10.28 — 审查修复与配置保留升级

本版本保留原 UniProxy 接口、鉴权参数和 id / uuid / speed_limit / device_limit 格式，旧面板无需同步更新。多订阅由面板将每条订阅的 node_user_id 映射到原 id 字段实现，面板改造见 docs/XBOARD_SUBSCRIPTIONS.md。

修复内容：
- 严格验证完整 JSON / MessagePack 用户列表，拒绝截断数据、重复身份和歧义字段。
- 修复设备限制变更丢失在线计数、内核更新失败后 ETag 阻止重试的问题。
- 隔离 sing-box 适配层中各本地节点的身份映射，避免移除一个节点身份影响其他节点流量。
- 升级等待节点真正初始化完成，再检查服务存活；失败时尝试恢复旧程序并明确报告恢复结果。
- 已有配置、损坏程序和残留安装目录均走配置保留流程。
- 配套程序、管理脚本与升级脚本按固定提交一起打包，提供 SHA256 校验。

内核依赖保持原锁定版本，本次没有升级 Xray、sing-box 或 Hysteria2 内核。

验证：本地回归测试，以及 Ubuntu 22.04 / x86_64 / systemd 上从原 v0.4.0 到审查构建的实机升级。旧面板保持不动，多轮真实流量及在线设备上报正常，原配置、证书、geo 和服务文件指纹一致。其他架构通过交叉构建；OpenRC 和其他协议未做同等实机验收。

从原作者脚本切换，root 执行：

```bash
wget -O upgrade.sh https://raw.githubusercontent.com/ChampagneA7/V2bX-script/master/upgrade.sh && bash upgrade.sh
```

已切换到本仓库维护脚本的安装，以后执行：

```bash
V2bX update
```

升级包含短暂重启。备份位于 /usr/local/V2bX/backups/；不会替换现有配置、证书、路由、geo 或 service 文件。
