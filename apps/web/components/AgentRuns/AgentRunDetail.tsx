'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import {
  AgentRun,
  AgentRunStep,
  AgentProfile,
  AgentRunStepStatus,
  AgentRunStatus,
  getAgentRun,
  getAgentRunSteps,
  getAgentProfiles,
  getAgentTeams,
  startAgentRun,
  cancelAgentRun,
} from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { RunStatusBadge } from './AgentRunBadges'
import { RunStepsList } from './RunStepsList'
import { AddRunStepForm } from './AddRunStepForm'
import { AgentMessagesPanel } from './AgentMessagesPanel'
import { HumanApprovalsPanel } from './HumanApprovalsPanel'
import { ToolCallsPanel } from './ToolCallsPanel'
import { formatDate } from '@/lib/utils'

interface AgentRunDetailProps {
  runId: string
}

export function AgentRunDetail({ runId }: AgentRunDetailProps) {
  const [run, setRun] = useState<AgentRun | null>(null)
  const [steps, setSteps] = useState<AgentRunStep[]>([])
  const [profiles, setProfiles] = useState<AgentProfile[]>([])
  const [teams, setTeams] = useState<{ [key: string]: { name: string } }>({})
  const [loading, setLoading] = useState(true)
  const [showAddStep, setShowAddStep] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [actionLoading, setActionLoading] = useState<'start' | 'cancel' | null>(null)

  const isRunningStatus = (status: AgentRunStatus): boolean => {
    return status === 'running'
  }

  useEffect(() => {
    async function fetchData() {
      try {
        const [runData, stepsData, profilesData, teamsData] = await Promise.all([
          getAgentRun(runId),
          getAgentRunSteps(runId),
          getAgentProfiles(),
          getAgentTeams(),
        ])

        setRun(runData)
        setSteps(stepsData)
        setProfiles(profilesData)

        const teamsMap: { [key: string]: { name: string } } = {}
        teamsData.forEach((t) => {
          teamsMap[t.id] = { name: t.name }
        })
        setTeams(teamsMap)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load run')
      } finally {
        setLoading(false)
      }
    }

    fetchData()

    let intervalId: NodeJS.Timeout | null = null

    const startPolling = () => {
      intervalId = setInterval(async () => {
        try {
          const [runData, stepsData] = await Promise.all([
            getAgentRun(runId),
            getAgentRunSteps(runId),
          ])
          setRun(runData)
          setSteps(stepsData)

          if (!isRunningStatus(runData.status)) {
            if (intervalId) {
              clearInterval(intervalId)
              intervalId = null
            }
          }
        } catch (err) {
          console.error('Polling error:', err)
        }
      }, 3000)
    }

    if (run && isRunningStatus(run.status)) {
      startPolling()
    }

    return () => {
      if (intervalId) {
        clearInterval(intervalId)
      }
    }
  }, [runId, run])

  async function handleStartRun() {
    if (!run) return
    setActionLoading('start')
    try {
      const updatedRun = await startAgentRun(run.id)
      setRun(updatedRun)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to start run')
    } finally {
      setActionLoading(null)
    }
  }

  async function handleCancelRun() {
    if (!run) return
    setActionLoading('cancel')
    try {
      const updatedRun = await cancelAgentRun(run.id)
      setRun(updatedRun)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to cancel run')
    } finally {
      setActionLoading(null)
    }
  }

  function handleStepAdded(step: AgentRunStep) {
    setSteps((prev) => [...prev, step])
    setShowAddStep(false)
  }

  function handleStatusChange(stepId: string, newStatus: AgentRunStepStatus) {
    setSteps((prev) =>
      prev.map((s) => (s.id === stepId ? { ...s, status: newStatus } : s))
    )
  }

  if (loading) {
    return (
      <div className="space-y-6">
        <Card>
          <CardContent className="py-8 text-center">
            <p className="text-slate-500 text-sm">Loading run...</p>
          </CardContent>
        </Card>
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-6">
        <Card className="border-red-200 bg-red-50">
          <CardContent className="py-5">
            <div className="flex items-start">
              <svg className="w-5 h-5 text-red-500 mr-3 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <div>
                <h3 className="text-sm font-medium text-red-800">
                  Error loading run
                </h3>
                <p className="text-sm text-red-700 mt-1">
                  {error}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  if (!run) {
    return (
      <div className="space-y-6">
        <Card>
          <CardContent className="py-8 text-center">
            <p className="text-slate-500 text-sm">Run not found</p>
          </CardContent>
        </Card>
      </div>
    )
  }

  const team = run.team_id ? teams[run.team_id] : null

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link href={`/tasks/${run.task_id}`} className="text-blue-600 hover:underline text-sm">
          ← Back to Task
        </Link>
      </div>

       <Card>
         <CardHeader className="pb-3">
           <div className="flex items-start justify-between gap-4">
             <div>
               <div className="flex items-center gap-3 mb-2">
                 <RunStatusBadge status={run.status} />
                 <span className="text-sm text-slate-500">
                   Run ID: {run.id}
                 </span>
               </div>
               <p className="text-sm text-slate-500">Task: {run.task_id}</p>
               {team && (
                 <p className="text-sm text-slate-500">Team: {team.name}</p>
               )}
             </div>
             <div className="flex gap-2">
               {(run.status === 'draft' || run.status === 'planned' || run.status === 'waiting_approval' || run.status === 'approved') && (
                 <Button
                   size="sm"
                   onClick={handleStartRun}
                   disabled={actionLoading !== null}
                 >
                   {actionLoading === 'start' ? 'Starting...' : 'Start Run'}
                 </Button>
               )}
               {run.status === 'running' && (
                 <Button
                   size="sm"
                   variant="secondary"
                   onClick={handleCancelRun}
                   disabled={actionLoading !== null}
                 >
                   {actionLoading === 'cancel' ? 'Cancelling...' : 'Cancel Run'}
                 </Button>
               )}
             </div>
           </div>
         </CardHeader>
        <CardContent className="space-y-4">
          {run.goal && (
            <div>
              <h3 className="text-sm font-medium text-slate-700 mb-1">Goal</h3>
              <p className="text-sm text-slate-600 whitespace-pre-wrap">{run.goal}</p>
            </div>
          )}

          {run.summary && (
            <div>
              <h3 className="text-sm font-medium text-slate-700 mb-1">Summary</h3>
              <p className="text-sm text-slate-600 whitespace-pre-wrap">{run.summary}</p>
            </div>
          )}

          <div className="flex gap-6 text-xs text-slate-500 pt-2 border-t border-slate-100">
            <span>Created: {formatDate(run.created_at)}</span>
            <span>Updated: {formatDate(run.updated_at)}</span>
          </div>
        </CardContent>
      </Card>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="space-y-6">
          <Card>
            <CardHeader className="pb-3">
              <div className="flex items-center justify-between">
                <CardTitle className="text-base">
                  Steps ({steps.length})
                </CardTitle>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => setShowAddStep(!showAddStep)}
                >
                  {showAddStep ? 'Cancel' : '+ Add Step'}
                </Button>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              {showAddStep && (
                <AddRunStepForm
                  runId={runId}
                  profiles={profiles}
                  onAdd={handleStepAdded}
                  onCancel={() => setShowAddStep(false)}
                />
              )}

              <RunStepsList
                steps={steps}
                profiles={profiles}
                onStatusChange={handleStatusChange}
              />
            </CardContent>
          </Card>

          <AgentMessagesPanel
            runId={runId}
            steps={steps}
            profiles={profiles}
          />
        </div>

        <div className="space-y-6">
          <HumanApprovalsPanel
            runId={runId}
            steps={steps}
          />

          <ToolCallsPanel
            runId={runId}
            steps={steps}
          />
        </div>
      </div>
    </div>
  )
}
