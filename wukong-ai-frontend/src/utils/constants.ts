/**
 * 任务状态常量
 */
export const TASK_STATUS = {
  PENDING: 'pending',
  QUEUED: 'queued',
  RUNNING: 'running',
  SUCCESS: 'success',
  FAILED: 'failed',
} as const

/**
 * 任务模式常量
 */
export const TASK_MODE = {
  FLASH: 'flash',
  STANDARD: 'standard',
  PRO: 'pro',
  ULTRA: 'ultra',
} as const

/**
 * DAG 节点类型
 */
export const DAG_NODE_TYPE = {
  USER_INPUT: 'user_input',
  INTENTION_ANALYSIS: 'intention_analysis',
  PLAN_GENERATION: 'plan_generation',
  TASK_EXECUTION: 'task_execution',
  SUBAGENT: 'subagent',
  SYNTHESIS: 'synthesis',
  MEMORY: 'memory',
  TOOLS: 'tools',
} as const

/**
 * DAG 节点状态
 */
export const DAG_NODE_STATUS = {
  PENDING: 'pending',
  RUNNING: 'running',
  SUCCESS: 'success',
  FAILED: 'failed',
} as const

/**
 * API 配置
 */
export const API_CONFIG = {
  BASE_URL: import.meta.env.VITE_API_BASE_URL || '/api',
  TIMEOUT: 30000,
  RETRY_COUNT: 3,
} as const

/**
 * SSE 配置
 */
export const SSE_CONFIG = {
  RECONNECT_DELAY: 3000,
  MAX_RECONNECT_ATTEMPTS: 10,
} as const

/**
 * 分页配置
 */
export const PAGINATION = {
  DEFAULT_PAGE_SIZE: 10,
  PAGE_SIZE_OPTIONS: [10, 20, 50, 100],
} as const

/**
 * 模式配置
 */
export const MODE_CONFIG = {
  FLASH: {
    label: '快速模式',
    thinking: false,
    plan: false,
    subagent: false,
  },
  STANDARD: {
    label: '标准模式',
    thinking: true,
    plan: false,
    subagent: false,
  },
  PRO: {
    label: '增强模式',
    thinking: true,
    plan: true,
    subagent: false,
  },
  ULTRA: {
    label: '超级模式',
    thinking: true,
    plan: true,
    subagent: true,
  },
} as const
