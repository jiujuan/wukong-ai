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
import { Button } from '@/components/ui'

// SVG Icons
const WukongIcon = () => (
  <svg viewBox="0 0 32 32" className="h-8 w-8" fill="none">
    <circle cx="16" cy="16" r="14" fill="#3b82f6" />
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
      className={`flex flex-col border-r bg-card transition-all duration-300 ${
        collapsed ? 'w-16' : 'w-60'
      }`}
    >
      <div className="flex h-16 items-center justify-between border-b border-border/60 px-4">
        {!collapsed && (
          <div className="flex items-center gap-2">
            <WukongIcon />
            <span className="text-lg font-semibold text-foreground">悟空 AI</span>
          </div>
        )}
        {collapsed && <WukongIcon />}
      </div>
      <nav className="flex-1 space-y-2 p-3">
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
                    ? 'border border-primary/25 bg-card text-foreground shadow-sm'
                    : 'text-muted-foreground hover:bg-accent/70 hover:text-foreground'
                }`}
                title={collapsed ? item.label : undefined}
              >
                <Icon className="h-5 w-5 flex-shrink-0" />
                {!collapsed && <span>{item.label}</span>}
              </NavLink>

              {!collapsed && item.path === '/dag' && (
                <div className="mt-1 space-y-1 pl-8">
                  <div className="flex items-center gap-1.5 px-2 py-1 text-sm font-semibold text-muted-foreground">
                    <History className="h-4 w-4" />
                    <span>最近任务</span>
                  </div>
                  {loading && recentTasks.length === 0 && (
                    <p className="text-xs text-muted-foreground">加载最近任务中...</p>
                  )}
                  {!loading && recentTasks.length === 0 && (
                    <p className="text-xs text-muted-foreground">暂无最近任务</p>
                  )}
                  {recentTasks.map((task) => (
                    <Link
                      key={task.task_id}
                      to={`/tasks/${task.task_id}`}
                      className="block truncate rounded px-2 py-1 text-xs text-muted-foreground hover:bg-accent hover:text-accent-foreground"
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

      <div className="space-y-1 p-3 pt-0">
        <NavLink
          to="/settings"
          className={`flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors ${
            location.pathname.startsWith('/settings')
              ? 'border border-primary/25 bg-card text-foreground shadow-sm'
              : 'text-muted-foreground hover:bg-accent/70 hover:text-foreground'
          }`}
          title={collapsed ? '设置' : undefined}
        >
          <Settings className="h-5 w-5 flex-shrink-0" />
          {!collapsed && <span>设置</span>}
        </NavLink>
      </div>
      <div className="border-t border-border/60 p-2">
        <Button
          onClick={onToggle}
          variant="ghost"
          className="w-full"
          title={collapsed ? '展开' : '折叠'}
        >
          {collapsed ? (
            <ChevronRight className="h-5 w-5" />
          ) : (
            <ChevronLeft className="h-5 w-5" />
          )}
        </Button>
      </div>
    </aside>
  )
}
