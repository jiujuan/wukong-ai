import apiClient from './client'

export interface UploadResult {
  file_name:     string
  attachment_id: number
  success:       boolean
  status?:       string
  error?:        string
}

export interface AttachmentStatus {
  attachment_id:  number
  file_name:      string
  mime_type:      string
  file_size:      number
  extract_status: string   // pending / extracting / done / failed
  is_image:       boolean
  chunk_count:    number
  error_msg?:     string
  upload_time:    string
  extract_time?:  string
}

export interface UploadStatusResponse {
  task_id:     string
  attachments: AttachmentStatus[]
}

export const uploadApi = {
  /** 上传多个文件，关联到指定 task_id */
  upload: async (taskId: string, files: File[]): Promise<UploadResult[]> => {
    const form = new FormData()
    form.append('task_id', taskId)
    files.forEach(f => form.append('files', f))
    const res = await apiClient.post('/api/upload', form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    return res.results ?? []
  },

  /** 查询附件提取状态（附件提取为异步，需轮询） */
  getStatus: (taskId: string): Promise<UploadStatusResponse> =>
    apiClient.get(`/api/upload/status?task_id=${taskId}`),
}
