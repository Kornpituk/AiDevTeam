'use client'

import { useState } from 'react'
import { AgentRunStep, AgentProfile, AgentRunStepStatus } from '@/lib/api'
import { Card, CardContent } from '@/components/ui/card'
import { StepStatusSelector } from './StepStatusSelector'
import { formatStepStatus } from './AgentRunBadges'
import { formatDate, cn } from '@/lib/utils'

interface RunStepsListProps {
  steps: AgentRunStep[]
  profiles: AgentProfile[]
  onStatusChange?: (stepId: string, newStatus: AgentRunStepStatus) => void
}

function getStepCircleClasses(status: AgentRunStepStatus): string {
  switch (status) {
    case 'completed':
      return 'bg-green-100 border-green-500 text-green-700'
    case 'running':
      return 'bg-blue-100 border-blue-500 text-blue-700 animate-pulse'
    case 'failed':
      return 'bg-red-100 border-red-500 text-red-700'
    case 'waiting_approval':
      return 'bg-amber-100 border-amber-500 text-amber-700'
    case 'pending':
      return 'bg-gray-100 border-gray-300 text-gray-500'
    case 'skipped':
      return 'bg-gray-100 border-gray-400 text-gray-400'
    case 'cancelled':
      return 'bg-gray-100 border-gray-400 text-gray-400'
    default:
      return 'bg-gray-100 border-gray-300 text-gray-500'
  }
}

function getStepContentClasses(status: AgentRunStepStatus): string {
  switch (status) {
    case 'running':
      return 'border-blue-200 bg-blue-50/50'
    case 'failed':
      return 'border-red-200 bg-red-50/50'
    case 'completed':
      return 'border-green-200 bg-green-50/50'
    case 'waiting_approval':
      return 'border-amber-200 bg-amber-50/50'
    default:
      return 'border-gray-200 bg-white'
  }
}

function getStepIcon(status: AgentRunStepStatus, position: number): string {
  switch (status) {
    case 'completed':
      return '✓'
    case 'failed':
      return '✗'
    case 'running':
      return '◉'
    case 'skipped':
      return '—'
    case 'cancelled':
      return '✕'
    case 'waiting_approval':
      return '⏸'
    default:
      return String(position)
  }
}

export function RunStepsList({ steps, profiles, onStatusChange }: RunStepsListProps) {
  const sortedSteps = [...steps].sort((a, b) => a.position - b.position)

  if (sortedSteps.length === 0) {
    return (
      <Card>
        <CardContent className="py-8 text-center">
          <p className="text-slate-500 text-sm">No steps yet. Add a manual step below.</p>
        </CardContent>
      </Card>
    )
  }

  function handleStatusChange(stepId: string, newStatus: AgentRunStepStatus) {
    onStatusChange?.(stepId, newStatus)
  }

  return (
    <div className="relative">
      {sortedSteps.map((step, index) => {
        const profile = profiles.find((p) => p.id === step.profile_id)
        const isLast = index === sortedSteps.length - 1

        return (
          <div key={step.id} className={cn('flex gap-4 relative', !isLast && 'pb-8')}>
            {/* Left: Position Circle + Connecting Line */}
            <div className="flex flex-col items-center">
              <div
                className={cn(
                  'w-8 h-8 rounded-full flex items-center justify-center text-xs font-medium border-2 z-10 transition-all duration-300',
                  getStepCircleClasses(step.status),
                  step.status === 'running' && 'ring-2 ring-blue-300 ring-offset-1'
                )}
              >
                {getStepIcon(step.status, step.position)}
              </div>
              {!isLast && (
                <div className="w-0.5 flex-1 bg-gray-200 min-h-[1rem]" />
              )}
            </div>

            {/* Right: Step Content */}
            <div
              className={cn(
                'flex-1 p-4 rounded-lg border transition-all duration-300',
                getStepContentClasses(step.status)
              )}
            >
              <div className="flex items-start justify-between gap-4 mb-2">
                <div className="flex items-center gap-3 min-w-0 flex-1">
                  <h4 className="font-medium text-slate-900 text-sm truncate">{step.title}</h4>
                  <span className="px-2 py-0.5 bg-slate-100 text-slate-600 rounded text-xs whitespace-nowrap flex-shrink-0">
                    {step.step_type}
                  </span>
                </div>
                <div className="flex-shrink-0">
                  <StepStatusSelector
                    stepId={step.id}
                    currentStatus={step.status}
                    onStatusChange={(newStatus) => handleStatusChange(step.id, newStatus)}
                  />
                </div>
              </div>

              {step.instructions && (
                <p className="text-sm text-slate-600 mb-2 whitespace-pre-wrap">
                  {step.instructions}
                </p>
              )}

              {step.output && (
                <details className="group mb-2">
                  <summary className="text-xs text-slate-500 cursor-pointer hover:text-slate-700 inline-flex items-center gap-1">
                    <svg
                      className="w-3 h-3 transition-transform group-open:rotate-90"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                    </svg>
                    View output
                  </summary>
                  <pre className="mt-2 p-3 bg-white/80 rounded text-xs text-slate-600 overflow-x-auto whitespace-pre-wrap border border-slate-200">
                    {step.output}
                  </pre>
                </details>
              )}

              <div className="flex items-center gap-4 text-xs text-slate-500 mt-2 flex-wrap">
                {profile && (
                  <span className="inline-flex items-center gap-1">
                    <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                    </svg>
                    {profile.name}
                  </span>
                )}
                <span>Created: {formatDate(step.created_at)}</span>
                {step.started_at && (
                  <span>Started: {formatDate(step.started_at)}</span>
                )}
                {step.completed_at && (
                  <span>Completed: {formatDate(step.completed_at)}</span>
                )}
              </div>
            </div>
          </div>
        )
      })}
    </div>
  )
}
