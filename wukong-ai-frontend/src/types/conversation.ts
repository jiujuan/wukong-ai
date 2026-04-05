import type { RunTaskResponse } from './task'

export interface Conversation {
  id: string
  title: string
  summary?: string
  turn_count: number
  create_time: string
  update_time: string
}

export interface ConversationTurn {
  id: number
  conversation_id: string
  task_id?: string
  turn_index: number
  role: 'user' | 'assistant'
  content: string
  full_output?: string
  create_time: string
}

export interface ConversationDetailResponse {
  conversation: Conversation
  turns: ConversationTurn[]
}

export interface ConversationListResponse {
  total: number
  page: number
  size: number
  conversations: Conversation[]
}

export interface CreateConversationRequest {
  title?: string
}

export interface CreateConversationResponse {
  conversation_id: string
  title: string
  create_time: string
}

export interface ConversationRunRequest {
  user_input: string
  thinking_enabled?: boolean
  plan_enabled?: boolean
  subagent_enabled?: boolean
  llm_provider?: string
  max_sub_agents?: number
  timeout_seconds?: number
}

export type ConversationRunResponse = RunTaskResponse
