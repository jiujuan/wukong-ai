import React from 'react'
import { FileQuestion } from 'lucide-react'

interface EmptyStateProps {
  icon?: React.ReactNode
  title: string
  description?: string
  action?: React.ReactNode
}

/**
 * 空状态
 */
export function EmptyState({ icon, title, description, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-12">
      <div className="mb-4 text-muted-foreground">
        {icon || <FileQuestion className="h-12 w-12" />}
      </div>
      <h3 className="mb-2 text-lg font-medium text-foreground">{title}</h3>
      {description && (
        <p className="mb-4 max-w-sm text-center text-sm text-muted-foreground">{description}</p>
      )}
      {action && <div>{action}</div>}
    </div>
  )
}
