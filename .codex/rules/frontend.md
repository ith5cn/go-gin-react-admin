---
description: React 前端的组件、状态、路由和 API 约定
paths: web/**
---

# 前端规则

- 基于 React 19、React Router 7、Zustand 5、Ant Design 6 和 Vite 8 开发；不要按旧版 API 假设实现。
- API 调用集中在 `src/api/` 并复用 `src/utils/request.ts`，页面组件不得自行创建 Axios 实例。
- 响应拦截器统一处理业务错误、安装页跳转、401 和 access token 刷新；组件 catch 时避免重复弹出相同错误。
- 认证、菜单、动态路由和标签页状态由 `src/store/auth.ts` 管理。渲染时使用 selector；命令式回调和拦截器边界可使用 `getState()`。
- 用户上下文由 `Layout` 初始化；登录页不得重复调用 `initUserContext`。清除会话必须同时清 Zustand 状态与持久化存储。
- 后端菜单先经 `normalizeBackendMenuTree`，再由转换函数生成侧栏和路由；新增静态布局路由修改 `staticRoutes.tsx`，公开路由修改 `publicRouters.tsx`。
- 页面放在 `src/pages/<module>/`，通用组件放 `src/components/`，模块私有组件放页面目录的 `components/`。
- 样式优先使用既有 Tailwind 工具类、Ant Design token 和组件 API，避免散落的硬编码主题色与内联样式。
- 所有新增代码通过 TypeScript strict、`pnpm lint` 和 `pnpm build`；注意 Ant Design、React Router 的实际主版本类型签名。
