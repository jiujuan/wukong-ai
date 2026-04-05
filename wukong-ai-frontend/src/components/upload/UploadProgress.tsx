import { useEffect, useState } from 'react'
import { uploadApi, AttachmentStatus } from '../../api/uploadApi'

interface Props {
  taskId: string
  onAllReady?: () => void
}

/** 轮询附件提取状态，全部 done 后回调 onAllReady */
export default function UploadProgress({ taskId, onAllReady }: Props) {
  const [attachments, setAttachments] = useState<AttachmentStatus[]>([])

  useEffect(() => {
    if (!taskId) return
    const timer = setInterval(async () => {
      try {
        const res = await uploadApi.getStatus(taskId)
        setAttachments(res.attachments ?? [])
        const allDone = res.attachments.every(
          a => a.extract_status === 'done' || a.extract_status === 'failed'
        )
        if (allDone && res.attachments.length > 0) {
          clearInterval(timer)
          onAllReady?.()
        }
      } catch (e) {
        console.error(e)
      }
    }, 2000)
    return () => clearInterval(timer)
  }, [taskId])

  if (attachments.length === 0) return null

  const doneCount = attachments.filter(a => a.extract_status === 'done').length
  const total = attachments.length

  return (
    <div className="text-sm text-gray-500">
      {doneCount < total
        ? `正在解析附件（${doneCount}/${total}），解析完成后可提交...`
        : `附件解析完成（${doneCount}/${total}）`}
    </div>
  )
}
