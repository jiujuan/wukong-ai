import React, { useState } from 'react'
import { Copy, Check } from 'lucide-react'

interface CopyButtonProps {
  text: string
  className?: string
}

/**
 * 复制按钮
 */
export function CopyButton({ text, className = '' }: CopyButtonProps) {
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(text)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      console.error('Failed to copy:', err)
    }
  }

  return (
    <button
      onClick={handleCopy}
      className={`flex items-center gap-1.5 rounded-lg px-2 py-1 text-sm text-gray-500 hover:bg-gray-100 hover:text-gray-700 ${className}`}
      title="复制"
    >
      {copied ? (
        <>
          <Check className="h-4 w-4 text-green-500" />
          已复制
        </>
      ) : (
        <>
          <Copy className="h-4 w-4" />
          复制
        </>
      )}
    </button>
  )
}
