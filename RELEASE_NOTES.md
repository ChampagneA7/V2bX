# ChampagneA7 V2bX compatibility release

- Keep the legacy UniProxy paths, authentication parameters, JSON/MessagePack fields and report bodies.
- Treat panel user IDs as opaque identities so a panel can map each subscription to a stable node_user_id.
- Apply ID / UUID / speed / device-limit changes and handle explicit empty user lists correctly.
- Detect online-device report errors without changing the old panel protocol.
- Keep the original pinned Xray, sing-box and Hysteria2 core dependencies.
- Include docs/XBOARD_SUBSCRIPTIONS.md for the panel implementation; no panel business code is changed.
- Provide Linux amd64, arm64 and s390x packages with SHA256 sidecars.

Existing configurations require no new fields. Upgrade an existing installation with the dedicated upgrade.sh in ChampagneA7/V2bX-script; it preserves configurations, certificates, service definitions and geo databases and restores the previous program on detected startup failure.
