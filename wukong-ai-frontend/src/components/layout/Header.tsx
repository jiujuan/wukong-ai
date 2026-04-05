import React from 'react'
import { useLocation } from 'react-router-dom'
import { Bell, User, Search } from 'lucide-react'
import { Button, Input } from '@/components/ui'

// 页面标题映射
const pageTitles: Record<string, string> = {
  '/tasks': '任务列表',
  '/tasks/new': '新建任务',
  '/dag': 'DAG 视图',
  '/settings': '设置',
}

interface HeaderProps {
  title?: string
}

/**
 * 头部导航
 */
export function Header({ title }: HeaderProps) {
  const location = useLocation()
  const pageTitle = title || pageTitles[location.pathname] || '悟空 AI'

  return (
    <header className="flex h-16 items-center justify-between border-b border-border/60 bg-card px-6">
      <div>
        <h1 className="text-xl font-semibold text-foreground">{pageTitle}</h1>
      </div>
      <div className="flex items-center gap-4">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            type="text"
            placeholder="搜索任务..."
            className="h-9 w-64 rounded-full border-border/80 bg-background pl-9"
          />
        </div>
        <Button
          variant="ghost"
          size="icon"
          className="relative"
          title="通知"
        >
          <Bell className="h-5 w-5" />
          <span className="absolute right-1.5 top-1.5 h-2 w-2 rounded-full bg-red-500"></span>
        </Button>
        <Button
          variant="ghost"
          className="gap-2"
          title="用户"
        >
          <User className="h-5 w-5" />
          <span className="text-sm">Admin</span>
        </Button>
      </div>
    </header>
  )
}
