import React, { useState } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AppLayout } from './components/layout'
import { TaskListPage, TaskDetailPage, DagViewPage, SettingsPage } from './pages'

/**
 * 应用主组件
 */
function App() {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)

  return (
    <BrowserRouter>
      <Routes>
        {/* 主布局 */}
        <Route element={<AppLayout />}>
          <Route index element={<Navigate to="/tasks" replace />} />

          {/* 任务列表 */}
          <Route path="/tasks" element={<TaskListPage />} />
          <Route path="/tasks/new" element={<TaskListPage />} />

          {/* 任务详情 */}
          <Route path="/tasks/:taskId" element={<TaskDetailPage />} />

          {/* DAG 视图 */}
          <Route path="/dag" element={<DagViewPage />} />

          {/* 设置 */}
          <Route path="/settings" element={<SettingsPage />} />
        </Route>

        {/* 404 */}
        <Route path="*" element={<Navigate to="/tasks" replace />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App
