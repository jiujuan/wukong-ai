import { useRef, useState } from 'react'
import { uploadApi, UploadResult } from '../../api/uploadApi'

const ACCEPTED_TYPES: Record<string, string[]> = {
  'text/plain':       ['.txt'],
  'text/markdown':    ['.md'],
  'text/csv':         ['.csv'],
  'application/pdf':  ['.pdf'],
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document': ['.docx'],
  'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet':       ['.xlsx'],
  'image/png':  ['.png'],
  'image/jpeg': ['.jpg', '.jpeg'],
  'image/webp': ['.webp'],
}

const MAX_SIZE_MB = 50
const ACCEPT_STR = Object.values(ACCEPTED_TYPES).flat().join(',')

interface Props {
  taskId: string
  onUploaded?: (results: UploadResult[]) => void
}

export default function FileUploadZone({ taskId, onUploaded }: Props) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [dragging, setDragging] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [results, setResults] = useState<UploadResult[]>([])

  const handleFiles = async (files: FileList | null) => {
    if (!files || files.length === 0) return
    const validFiles: File[] = []
    for (const f of Array.from(files)) {
      if (f.size > MAX_SIZE_MB * 1024 * 1024) {
        alert(`文件 ${f.name} 超过 ${MAX_SIZE_MB}MB 限制`)
        continue
      }
      validFiles.push(f)
    }
    if (validFiles.length === 0) return

    setUploading(true)
    try {
      const res = await uploadApi.upload(taskId, validFiles)
      setResults(prev => [...prev, ...res])
      onUploaded?.(res)
    } catch (e) {
      console.error('upload error', e)
    } finally {
      setUploading(false)
    }
  }

  return (
    <div className="space-y-3">
      <div
        className={`border-2 border-dashed rounded-lg p-6 text-center cursor-pointer transition-colors
          ${dragging ? 'border-blue-500 bg-blue-50' : 'border-gray-300 hover:border-blue-400'}`}
        onDragOver={e => { e.preventDefault(); setDragging(true) }}
        onDragLeave={() => setDragging(false)}
        onDrop={e => { e.preventDefault(); setDragging(false); handleFiles(e.dataTransfer.files) }}
        onClick={() => inputRef.current?.click()}
      >
        <input
          ref={inputRef}
          type="file"
          multiple
          accept={ACCEPT_STR}
          className="hidden"
          onChange={e => handleFiles(e.target.files)}
        />
        {uploading ? (
          <p className="text-blue-500 text-sm">上传中...</p>
        ) : (
          <>
            <p className="text-gray-500 text-sm">拖拽文件到此处，或点击选择文件</p>
            <p className="text-gray-400 text-xs mt-1">
              支持 PDF / DOCX / XLSX / CSV / TXT / MD / PNG / JPG / WEBP，最大 {MAX_SIZE_MB}MB
            </p>
          </>
        )}
      </div>

      {results.length > 0 && (
        <ul className="space-y-1">
          {results.map((r, i) => (
            <li key={i} className="flex items-center gap-2 text-sm">
              <span className={r.success ? 'text-green-600' : 'text-red-500'}>
                {r.success ? '✓' : '✗'}
              </span>
              <span className="flex-1 truncate">{r.file_name}</span>
              <span className="text-gray-400 text-xs">
                {r.success ? r.status : r.error}
              </span>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
