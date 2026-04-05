import React from 'react'
import { Button, Card, CardContent, CardHeader, CardTitle, Input } from '@/components/ui'

/**
 * 设置页面
 */
export function SettingsPage() {
  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-xl font-semibold text-foreground">设置</h2>
        <p className="mt-1 text-sm text-muted-foreground">管理您的应用设置</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">API 配置</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div>
              <label className="mb-2 block text-sm font-medium text-foreground">
                API 地址
              </label>
              <Input type="text" defaultValue="/api" />
            </div>
            <div>
              <label className="mb-2 block text-sm font-medium text-foreground">
                请求超时 (毫秒)
              </label>
              <Input type="number" defaultValue={30000} />
            </div>
          </div>
          <div className="mt-6 flex justify-end">
            <Button>
            保存设置
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
