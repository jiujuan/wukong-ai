import apiClient from './client'
import type {
  ConversationDetailResponse,
  ConversationListResponse,
  ConversationRunRequest,
  ConversationRunResponse,
  CreateConversationRequest,
  CreateConversationResponse,
} from '@/types'

export const conversationApi = {
  create: async (req: CreateConversationRequest = {}): Promise<CreateConversationResponse> => {
    return apiClient.post('/api/conversation', req)
  },

  getDetail: async (conversationId: string): Promise<ConversationDetailResponse> => {
    return apiClient.get(`/api/conversation/${conversationId}`)
  },

  list: async (page = 1, size = 50): Promise<ConversationListResponse> => {
    return apiClient.get(`/api/conversation/list?page=${page}&size=${size}`)
  },

  run: async (conversationId: string, req: ConversationRunRequest): Promise<ConversationRunResponse> => {
    return apiClient.post(`/api/conversation/${conversationId}/run`, req)
  },

  findConversationIdByTask: async (taskId: string): Promise<string | null> => {
    const listResp = await conversationApi.list(1, 100)
    for (const conv of listResp.conversations) {
      try {
        const detail = await conversationApi.getDetail(conv.id)
        if (detail.turns.some((turn) => turn.task_id === taskId)) {
          return conv.id
        }
      } catch {
      }
    }
    return null
  },
}

export default conversationApi
