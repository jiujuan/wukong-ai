import React from 'react'
import { useLocation } from 'react-router-dom'
import { Bell, User, Search } from 'lucide-react'

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
    <header className="flex h-16 items-center justify-between border-b border-gray-200 bg-white px-6">
      {/* 标题 */}
      <div>
        <h1 className="text-xl font-semibold text-gray-900">{pageTitle}</h1>
      </div>

      {/* 操作区 */}
      <div className="flex items-center gap-4">
        {/* 搜索 */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
          <input
            type="text"
            placeholder="搜索任务..."
            className="h-9 w-64 rounded-lg border border-gray-300 bg-white pl-9 pr-4 text-sm placeholder-gray-500 focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
          />
        </div>

        {/* 通知 */}
        <button
          className="relative rounded-lg p-2 text-gray-500 hover:bg-gray-100 hover:text-gray-700"
          title="通知"
        >
          <Bell className="h-5 w-5" />
          <span className="absolute right-1.5 top-1.5 h-2 w-2 rounded-full bg-red-500"></span>
        </button>

        {/* 用户 */}
        <button
          className="flex items-center gap-2 rounded-lg p-2 text-gray-500 hover:bg-gray-100 hover:text-gray-700"
          title="用户"
        >
          <User className="h-5 w-5" />
          <span className="text-sm">Admin</span>
        </button>
      </div>
    </header>
  )
}
