'use client'

import { useState } from 'react'
import { AgentRunStep, AgentProfile, AgentRunStepStatus } from '@/lib/api'
import { Card, CardContent } from '@/components/ui/card'
import { StepStatusSelector } from './StepStatusSelector'
import { formatStepStatus } from './AgentRunBadges'
import { formatDate } from '@/lib/utils'

interface RunStepsListProps {
  steps: AgentRunStep[]
  profiles: AgentProfile[]
  onStatusChange?: (stepId: string, newStatus: AgentRunStepStatus) => void
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
    <div className="space-y-3">
      {sortedSteps.map((step) => {
        const profile = profiles.find((p) => p.id === step.profile_id)
        return (
          <Card key={step.id} className="hover:shadow-sm transition-shadow">
            <CardContent className="py-4">
              <div className="flex items-start justify-between gap-4">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-3 mb-2">
                    <span className="inline-flex items-center justify-center w-6 h-6 rounded-full bg-slate-100 text-slate-600 text-xs font-medium">
                      {step.position}
                    </span>
                    <h4 className="font-medium text-slate-900 text-sm">{step.title}</h4>
                    <span className="px-2 py-0.5 bg-slate-100 text-slate-600 rounded text-xs">
                      {step.step_type}
                    </span>
                  </div>

                  {step.instructions && (
                    <p className="text-sm text-slate-600 mb-2 ml-9 whitespace-pre-wrap">
                      {step.instructions}
                    </p>
                  )}

                  {step.output && (
                    <details className="ml-9">
                      <summary className="text-xs text-slate-500 cursor-pointer hover:text-slate-700">
                        View output
                      </summary>
                      <pre className="mt-2 p-3 bg-slate-50 rounded text-xs text-slate-600 overflow-x-auto whitespace-pre-wrap">
                        {step.output}
                      </pre>
                    </details>
                  )}

                  <div className="flex items-center gap-4 mt-2 ml-9 text-xs text-slate-500">
                    {profile && (
                      <span>Agent: {profile.name}</span>
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

                <div className="flex-shrink-0">
                  <StepStatusSelector
                    stepId={step.id}
                    currentStatus={step.status}
                    onStatusChange={(newStatus) => handleStatusChange(step.id, newStatus)}
                  />
                </div>
              </div>
            </CardContent>
          </Card>
        )
      })}
    </div>
  )
}
