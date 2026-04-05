import { CheckCircle2, Copy, CheckCheck } from 'lucide-react'
import { useMemo, useState } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import type { Components } from 'react-markdown'
import type { ReactNode } from 'react'
import { Button, Card, CardContent, CardHeader, CardTitle } from '@/components/ui'

interface TaskResultPanelProps {
  content: string
  title?: string
}

function renderInlineBreaks(children: ReactNode): ReactNode {
  if (typeof children === 'string') {
    const parts = children.split(/<br\s*\/?>/gi)
    if (parts.length === 1) {
      return children
    }
    return parts.flatMap((part, index) => (index < parts.length - 1 ? [part, <br key={`br-${index}`} />] : [part]))
  }
  if (Array.isArray(children)) {
    return children.map((child, index) => <span key={index}>{renderInlineBreaks(child)}</span>)
  }
  return children
}

const markdownComponents: Components = {
  h1: ({ children }) => <h1 className="mb-3 mt-4 text-2xl font-bold text-gray-900">{children}</h1>,
  h2: ({ children }) => <h2 className="mb-2 mt-4 text-xl font-bold text-gray-900">{children}</h2>,
  h3: ({ children }) => <h3 className="mb-2 mt-3 text-lg font-semibold text-gray-900">{children}</h3>,
  p: ({ children }) => <p className="text-sm leading-7 text-gray-700">{children}</p>,
  strong: ({ children }) => <strong className="font-bold text-gray-900">{children}</strong>,
  em: ({ children }) => <em className="italic text-gray-800">{children}</em>,
  a: ({ children, href }) => (
    <a href={href} target="_blank" rel="noreferrer" className="text-indigo-600 underline hover:text-indigo-700">
      {children}
    </a>
  ),
  ul: ({ children }) => <ul className="list-disc space-y-1 pl-6 text-sm leading-7 text-gray-700">{children}</ul>,
  ol: ({ children }) => <ol className="list-decimal space-y-1 pl-6 text-sm leading-7 text-gray-700">{children}</ol>,
  blockquote: ({ children }) => <blockquote className="border-l-4 border-gray-300 pl-3 text-sm text-gray-600">{children}</blockquote>,
  hr: () => <hr className="my-3 border-gray-200" />,
  table: ({ children }) => <table className="min-w-full border-collapse text-sm text-gray-700">{children}</table>,
  thead: ({ children }) => <thead className="bg-gray-50">{children}</thead>,
  th: ({ children }) => <th className="border border-gray-200 px-3 py-2 text-left font-medium">{children}</th>,
  td: ({ children }) => <td className="border border-gray-200 px-3 py-2 align-top">{renderInlineBreaks(children)}</td>,
  pre: ({ children }) => <pre className="overflow-x-auto rounded-lg border border-gray-200 bg-gray-100 p-3 text-xs text-gray-700">{children}</pre>,
  code: ({ children }) => <code className="rounded bg-gray-100 px-1 py-0.5 text-xs text-gray-800">{children}</code>,
}

function normalizeStreamingMarkdown(content: string) {
  const normalizedBr = content
    .split('\n')
    .map((line) => {
      if (line.includes('|')) {
        return line
      }
      return line.replace(/<br\s*\/?>/gi, '\n')
    })
    .join('\n')

  const normalized = normalizedBr
    .replace(/\r\n/g, '\n')
    .replace(/([。！？.!?:：\)])\s*(#{1,6}\s)/g, '$1\n$2')
  const fenceCount = (normalized.match(/```/g) || []).length
  if (fenceCount % 2 === 1) {
    return `${normalized}\n\`\`\``
  }
  const boldCount = (normalized.match(/\*\*/g) || []).length
  if (boldCount % 2 === 1) {
    return `${normalized}**`
  }
  return normalized
}

/**
 * 任务结果面板
 */
export function TaskResultPanel({ content, title = '执行结果' }: TaskResultPanelProps) {
  const [copied, setCopied] = useState(false)
  const renderedContent = useMemo(() => normalizeStreamingMarkdown(content), [content])

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(content)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      console.error('Failed to copy:', err)
    }
  }

  return (
    <Card>
      <CardHeader className="flex-row items-center justify-between space-y-0 border-b">
        <div className="flex items-center gap-2">
          <CheckCircle2 className="h-5 w-5 text-emerald-500" />
          <CardTitle className="text-base">{title}</CardTitle>
        </div>
        <Button
          onClick={handleCopy}
          variant="ghost"
          size="sm"
          className="gap-1.5"
        >
          {copied ? (
            <>
              <CheckCheck className="h-4 w-4 text-green-500" />
              已复制
            </>
          ) : (
            <>
              <Copy className="h-4 w-4" />
              复制
            </>
          )}
        </Button>
      </CardHeader>
      <CardContent className="p-6">
        <div className="space-y-1 break-words">
          <ReactMarkdown remarkPlugins={[remarkGfm]} components={markdownComponents}>
            {renderedContent}
          </ReactMarkdown>
        </div>
      </CardContent>
    </Card>
  )
}
