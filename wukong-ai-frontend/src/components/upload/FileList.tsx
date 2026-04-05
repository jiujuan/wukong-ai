import { AttachmentStatus } from '../../api/uploadApi'

const STATUS_LABEL: Record<string, string> = {
  pending:    '等待解析',
  extracting: '解析中...',
  done:       '解析完成',
  failed:     '解析失败',
}

const STATUS_COLOR: Record<string, string> = {
  pending:    'text-gray-400',
  extracting: 'text-blue-500',
  done:       'text-green-600',
  failed:     'text-red-500',
}

interface Props {
  attachments: AttachmentStatus[]
}

export default function FileList({ attachments }: Props) {
  if (attachments.length === 0) return null
  return (
    <div className="border rounded-lg divide-y text-sm">
      {attachments.map(a => (
        <div key={a.attachment_id} className="flex items-center gap-3 px-3 py-2">
          <span className="text-lg">{a.is_image ? '🖼️' : '📄'}</span>
          <span className="flex-1 truncate">{a.file_name}</span>
          <span className="text-gray-400 text-xs">
            {(a.file_size / 1024).toFixed(1)} KB
          </span>
          {a.extract_status === 'done' && a.chunk_count > 0 && (
            <span className="text-gray-400 text-xs">{a.chunk_count} 块</span>
          )}
          <span className={`text-xs ${STATUS_COLOR[a.extract_status] ?? 'text-gray-400'}`}>
            {STATUS_LABEL[a.extract_status] ?? a.extract_status}
          </span>
        </div>
      ))}
    </div>
  )
}
