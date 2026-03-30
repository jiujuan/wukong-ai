---
AIGC:
    ContentProducer: Minimax Agent AI
    ContentPropagator: Minimax Agent AI
    Label: AIGC
    ProduceID: "00000000000000000000000000000000"
    PropagateID: "00000000000000000000000000000000"
    ReservedCode1: 30450221008e169cfe0f6fc14dc2663bedac788217adeaf5fc6096fb54adc506abc56f5c5b0220448774c87b58fe76fcd9a0921b05501c82bdde1d3cf789c9b17cfcb0cd5b7c03
    ReservedCode2: 304502206f13f21b877161215225f18a666ce00538abd38aff8f907d04728246dc492950022100f06dfaa89128c67274be70911ed63d66924b830d0eb418bfe8b324c7eb16ddc2
---

# 悟空 AI 前端

基于 React + TypeScript + Vite 的悟空 AI 任务执行平台前端应用。

## 技术栈

- **框架**: React 18 + TypeScript
- **构建工具**: Vite
- **状态管理**: Zustand
- **路由**: React Router 6
- **HTTP 客户端**: Axios
- **样式**: Tailwind CSS
- **图标**: Lucide React

## 功能特性

- 任务列表管理（创建、运行、取消、续跑）
- 任务详情查看
- DAG 工作流可视化
- 四种执行模式：快速、标准、增强、超级
- 实时任务进度（通过 SSE）
- 响应式设计

## 开发

```bash
# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 构建生产版本
npm run build

# 预览生产版本
npm run preview
```

## 项目结构

```
src/
├── api/          # API 客户端
├── components/   # React 组件
│   ├── common/   # 通用组件
│   ├── dag/      # DAG 可视化组件
│   ├── layout/   # 布局组件
│   ├── mode/     # 模式选择器
│   └── task/     # 任务相关组件
├── hooks/        # 自定义 Hooks
├── pages/        # 页面组件
├── store/        # Zustand 状态管理
├── types/        # TypeScript 类型定义
└── utils/        # 工具函数
```

## 环境变量

```env
VITE_API_BASE_URL=/api
```
