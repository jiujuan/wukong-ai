import React from 'react'
import { AlertCircle, X } from 'lucide-react'

interface ErrorAlertProps {
  message: string
  onDismiss?: () => void
}

/**
 * 错误提示
 */
export function ErrorAlert({ message, onDismiss }: ErrorAlertProps) {
  return (
    <div className="flex items-start gap-3 rounded-lg bg-red-50 p-4">
      <AlertCircle className="h-5 w-5 flex-shrink-0 text-red-500" />
      <div className="flex-1">
        <h4 className="text-sm font-medium text-red-800">错误</h4>
        <p className="mt-1 text-sm text-red-700">{message}</p>
      </div>
      {onDismiss && (
        <button
          onClick={onDismiss}
          className="text-red-500 hover:text-red-700"
        >
          <X className="h-5 w-5" />
        </button>
      )}
    </div>
  )
}
