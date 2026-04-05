import React from 'react'
import { Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { Header } from './Header'

interface AppLayoutProps {
  children?: React.ReactNode
}

/**
 * 应用主布局
 */
export function AppLayout({ children }: AppLayoutProps) {
  return (
    <div className="flex h-screen bg-background">
      <Sidebar />
      <div className="flex flex-1 flex-col overflow-hidden">
        <Header />
        <main className="flex-1 overflow-auto p-6">
          {children || <Outlet />}
        </main>
      </div>
    </div>
  )
}
