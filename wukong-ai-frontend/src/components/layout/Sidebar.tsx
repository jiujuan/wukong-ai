import React, { useEffect } from 'react'
import { Link, NavLink, useLocation } from 'react-router-dom'
import {
  ListTodo,
  PlusSquare,
  GitBranch,
  History,
  Settings,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react'
import { useTaskStore } from '@/store'

// SVG Icons
const WukongIcon = () => (
  <svg viewBox="0 0 32 32" className="h-8 w-8" fill="none">
    <circle cx="16" cy="16" r="14" fill="#6366f1" />
    <text x="16" y="21" textAnchor="middle" fill="white" fontSize="14" fontWeight="bold">
      W
    </text>
  </svg>
)

interface SidebarProps {
  collapsed?: boolean
  onToggle?: () => void
}

/**
 * 侧边栏导航
 */
export function Sidebar({ collapsed = false, onToggle }: SidebarProps) {
  const location = useLocation()
  const tasks = useTaskStore((state) => state.tasks)
  const fetchTasks = useTaskStore((state) => state.fetchTasks)
  const loading = useTaskStore((state) => state.loading)

  useEffect(() => {
    fetchTasks(1)
  }, [fetchTasks])

  const navItems = [
    { path: '/tasks', label: '任务列表', icon: ListTodo },
    { path: '/tasks/new', label: '新建任务', icon: PlusSquare },
    { path: '/dag', label: 'DAG 视图', icon: GitBranch },
  ]

  const recentTasks = tasks.slice(0, 6)
  const isTasksActive = location.pathname === '/tasks'
  const isNewTaskActive = location.pathname === '/tasks/new'
  const isDagActive = location.pathname.startsWith('/dag')

  return (
    <aside
      className={`flex flex-col border-r border-gray-200 bg-white transition-all duration-300 ${
        collapsed ? 'w-16' : 'w-64'
      }`}
    >
      {/* Logo */}
      <div className="flex h-16 items-center justify-between border-b border-gray-200 px-4">
        {!collapsed && (
          <div className="flex items-center gap-2">
            <WukongIcon />
            <span className="text-lg font-semibold text-gray-900">悟空 AI</span>
          </div>
        )}
        {collapsed && <WukongIcon />}
      </div>

      {/* 导航菜单 */}
      <nav className="flex-1 space-y-1 p-2">
        {navItems.map((item) => {
          const Icon = item.icon
          let isActive = location.pathname.startsWith(item.path)
          if (item.path === '/tasks') isActive = isTasksActive
          if (item.path === '/tasks/new') isActive = isNewTaskActive
          if (item.path === '/dag') isActive = isDagActive

          return (
            <div key={item.path}>
              <NavLink
                to={item.path}
                className={`flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors ${
                  isActive
                    ? 'bg-indigo-50 text-indigo-600'
                    : 'text-gray-700 hover:bg-gray-100 hover:text-gray-900'
                }`}
                title={collapsed ? item.label : undefined}
              >
                <Icon className="h-5 w-5 flex-shrink-0" />
                {!collapsed && <span>{item.label}</span>}
              </NavLink>

              {!collapsed && item.path === '/dag' && (
                <div className="mt-1 space-y-1 pl-8">
                  <div className="flex items-center gap-1.5 px-2 py-1 text-sm font-semibold text-gray-500">
                    <History className="h-4 w-4" />
                    <span>最近任务</span>
                  </div>
                  {loading && recentTasks.length === 0 && (
                    <p className="text-xs text-gray-400">加载最近任务中...</p>
                  )}
                  {!loading && recentTasks.length === 0 && (
                    <p className="text-xs text-gray-400">暂无最近任务</p>
                  )}
                  {recentTasks.map((task) => (
                    <Link
                      key={task.task_id}
                      to={`/tasks/${task.task_id}`}
                      className="block truncate rounded px-2 py-1 text-xs text-gray-600 hover:bg-gray-100 hover:text-gray-900"
                      title={`${task.task_id} ${task.user_input}`}
                    >
                      {task.user_input || task.task_id}
                    </Link>
                  ))}
                </div>
              )}
            </div>
          )
        })}
      </nav>

      <div className="space-y-1 p-2 pt-0">
        <NavLink
          to="/settings"
          className={`flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors ${
            location.pathname.startsWith('/settings')
              ? 'bg-indigo-50 text-indigo-600'
              : 'text-gray-700 hover:bg-gray-100 hover:text-gray-900'
          }`}
          title={collapsed ? '设置' : undefined}
        >
          <Settings className="h-5 w-5 flex-shrink-0" />
          {!collapsed && <span>设置</span>}
        </NavLink>
      </div>

      {/* 折叠按钮 */}
      <div className="border-t border-gray-200 p-2">
        <button
          onClick={onToggle}
          className="flex w-full items-center justify-center rounded-lg p-2 text-gray-500 hover:bg-gray-100 hover:text-gray-700"
          title={collapsed ? '展开' : '折叠'}
        >
          {collapsed ? (
            <ChevronRight className="h-5 w-5" />
          ) : (
            <ChevronLeft className="h-5 w-5" />
          )}
        </button>
      </div>
    </aside>
  )
}
