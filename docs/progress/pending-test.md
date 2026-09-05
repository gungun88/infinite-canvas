---
title: 待测试
description: 当前版本已实现但仍需人工验证的变更项
---

# 待测试

- 登录注册页新增 Google、DoingFB 登录入口，以及邮箱注册、邮箱验证、忘记密码和重新发送验证邮件流程。
- 登录注册页不再显示 Linux.do 登录入口。
- Docker 部署配置 `APP_BIND_ADDRESS=127.0.0.1`、`APP_PORT=18902` 和独立 `CONTAINER_NAME` 后，应能与旧项目并行运行。
- 模型渠道下拉框应移除 MiniMax & METASO、APIMart、88API、KIE、MiMo，并显示 `doingAI`；选择 `doingAI` 后接口地址应为 `https://ai.doingfb.com/v1`，点击“获取 API Key”应打开 `https://ai.doingfb.com/keys`。
- 配置 `GOOGLE_CLIENT_ID`、`GOOGLE_CLIENT_SECRET`、`GOOGLE_REDIRECT_URI` 后，Google OAuth 应能完成授权回调并登录。
- 配置 `DOINGFB_CLIENT_ID`、`DOINGFB_CLIENT_SECRET`、`DOINGFB_REDIRECT_URI` 后，DoingFB OAuth 应能完成授权回调并登录。
- Google / DoingFB OAuth 登录后退出账号应直接进入干净登录页；即使回调后的用户信息请求未完成，也不应恢复旧会话或跳过登录页。
- 配置 `SMTP_HOST`、`SMTP_FROM` 等邮件变量后，注册验证邮件、重发验证邮件和密码重置邮件应能正常发送；未配置 SMTP 时，原用户名注册仍可直接登录。
- 首页轮播不应再显示 88API 和 MiniMax & METASO 两个宣传轮播项，其他首页媒体应保持正常显示和切换。
