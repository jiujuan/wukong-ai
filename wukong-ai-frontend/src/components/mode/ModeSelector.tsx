import React from 'react'
import { Zap, Brain, Map, Users } from 'lucide-react'
import { useTask } from '@/hooks'
import { calculateMode } from '@/store'

/**
 * 模式选择器
 */
export function ModeSelector() {
  const { modeConfig, toggleThinking, togglePlan, toggleSubagent } = useTask()

  const currentMode = calculateMode(modeConfig)

  const modeOptions = [
    {
      key: 'thinking',
      icon: Brain,
      label: '思考',
      description: '启用思维链',
      enabled: modeConfig.thinking,
      onToggle: toggleThinking,
      mode: 'standard',
    },
    {
      key: 'plan',
      icon: Map,
      label: '计划',
      description: '生成执行计划',
      enabled: modeConfig.plan,
      onToggle: togglePlan,
      mode: 'pro',
    },
    {
      key: 'subagent',
      icon: Users,
      label: '子代理',
      description: '启用子代理',
      enabled: modeConfig.subagent,
      onToggle: toggleSubagent,
      mode: 'ultra',
    },
  ]

  const modeLabels = {
    flash: '快速模式',
    standard: '标准模式',
    pro: '增强模式',
    ultra: '超级模式',
  }

  return (
    <div className="rounded-lg border border-gray-200 bg-white p-4">
      <div className="mb-4 flex items-center justify-between">
        <h3 className="font-medium text-gray-900">执行模式</h3>
        <span className="rounded-full bg-indigo-100 px-3 py-1 text-sm font-medium text-indigo-700">
          {modeLabels[currentMode]}
        </span>
      </div>

      <div className="space-y-3">
        {modeOptions.map((option) => {
          const Icon = option.icon
          return (
            <label
              key={option.key}
              className={`
                flex cursor-pointer items-center justify-between rounded-lg border-2 p-3 transition-all
                ${option.enabled ? 'border-indigo-500 bg-indigo-50' : 'border-gray-200 hover:border-gray-300'}
              `}
            >
              <div className="flex items-center gap-3">
                <div
                  className={`
                    flex h-10 w-10 items-center justify-center rounded-lg
                    ${option.enabled ? 'bg-indigo-500 text-white' : 'bg-gray-100 text-gray-500'}
                  `}
                >
                  <Icon className="h-5 w-5" />
                </div>
                <div>
                  <p className="font-medium text-gray-900">{option.label}</p>
                  <p className="text-sm text-gray-500">{option.description}</p>
                </div>
              </div>
              <input
                type="checkbox"
                checked={option.enabled}
                onChange={option.onToggle}
                className="h-5 w-5 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
              />
            </label>
          )
        })}
      </div>

      <div className="mt-4 rounded-lg bg-gray-50 p-3">
        <div className="flex items-center gap-2">
          <Zap className="h-4 w-4 text-amber-500" />
          <span className="text-sm text-gray-600">
            {currentMode === 'flash' && '极速响应，适合简单任务'}
            {currentMode === 'standard' && '深度思考，适合复杂分析'}
            {currentMode === 'pro' && '规划执行，适合多步骤任务'}
            {currentMode === 'ultra' && '全功能支持，适合大规模任务'}
          </span>
        </div>
      </div>
    </div>
  )
}
