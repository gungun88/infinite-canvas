# 本地渠道选择器覆盖

## 目的

`custom/auth-methods` 分支的“新增渠道”选择器是本地产品配置，不跟随上游协议注册表展示全部渠道。

新建渠道时固定显示并按以下顺序排列：

1. `DoingAI`
2. `OpenAI`
3. `Gemini`
4. `自定义`

`自定义` 使用 OpenAI 兼容请求格式，Base URL 由用户填写。

## 实现位置

- 可见选项和历史协议兼容逻辑：[custom-channel-policy.ts](../../web/src/lib/custom-channel-policy.ts)
- 管理后台新增/编辑渠道：[settings/page.tsx](../../web/src/app/(admin)/admin/settings/page.tsx)
- 用户本地直连渠道：[app-config-modal.tsx](../../web/src/components/layout/app-config-modal.tsx)

上游协议注册表仍保留在 `model-channel.ts`，用于已有渠道的运行时识别、模型能力判断和历史配置读取，不应直接作为新增渠道下拉框的选项来源。

## 上游合并规则

后续同步上游时，必须保留 `custom-channel-policy.ts` 及两个入口对它的引用。若上游修改渠道注册表或下拉框：

- 新增渠道的默认选项仍以本文件约定为准。
- 旧渠道协议只在编辑对应旧配置时临时显示，不能重新加入新建渠道的默认列表。
- `自定义` 的协议值保持为 `custom`，Base URL 不自动填充。
- 合并完成后检查 `web/src/lib/custom-channel-policy.test.ts`。

该文件是本地定制存档，不因上游新增或调整渠道而删除、替换或改回全量协议列表。
