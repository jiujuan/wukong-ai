import React from 'react'
import { AlertCircle, X } from 'lucide-react'
import { Alert, AlertDescription, AlertTitle, Button } from '@/components/ui'

interface ErrorAlertProps {
  message: string
  onDismiss?: () => void
}

/**
 * 错误提示
 */
export function ErrorAlert({ message, onDismiss }: ErrorAlertProps) {
  return (
    <Alert variant="destructive" className="flex items-start gap-3">
      <AlertCircle className="h-5 w-5 flex-shrink-0" />
      <div className="flex-1">
        <AlertTitle>错误</AlertTitle>
        <AlertDescription>{message}</AlertDescription>
      </div>
      {onDismiss && (
        <Button
          onClick={onDismiss}
          variant="ghost"
          size="icon"
          className="h-8 w-8 text-destructive hover:text-destructive"
        >
          <X className="h-5 w-5" />
        </Button>
      )}
    </Alert>
  )
}
