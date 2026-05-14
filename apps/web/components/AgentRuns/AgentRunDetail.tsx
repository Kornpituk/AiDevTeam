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
  resumeAgentRun,
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
  const [actionLoading, setActionLoading] = useState<'start' | 'cancel' | 'resume' | null>(null)

  useEffect(() => {
    async function fetchInitialData() {
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

    fetchInitialData()
  }, [runId])

  useEffect(() => {
    if (run?.status !== 'running' && run?.status !== 'paused') {
      return
    }

    let intervalId: NodeJS.Timeout | null = null

    async function pollData() {
      try {
        const [runData, stepsData] = await Promise.all([
          getAgentRun(runId),
          getAgentRunSteps(runId),
        ])
        setRun(runData)
        setSteps(stepsData)
      } catch (err) {
        console.error('Polling error:', err)
      }
    }

    intervalId = setInterval(pollData, 3000)

    return () => {
      if (intervalId) {
        clearInterval(intervalId)
      }
    }
  }, [runId, run?.status])

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

  async function handleResumeRun() {
    if (!run) return
    setActionLoading('resume')
    try {
      const updatedRun = await resumeAgentRun(run.id)
      setRun(updatedRun)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to resume run')
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
                  <p className="text-sm font-medium text-slate-700">{team.name}</p>
                )}
                {steps.length > 0 && (
                  <div className="flex items-center gap-4 text-xs text-slate-500 mt-2">
                    <span className="inline-flex items-center gap-1.5">
                      <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                      </svg>
                      Steps: {steps.filter(s => s.status === 'completed').length} / {steps.length} completed
                    </span>
                  </div>
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
                {run.status === 'paused' && (
                  <Button
                    size="sm"
                    className="bg-amber-600 hover:bg-amber-700 text-white"
                    onClick={handleResumeRun}
                    disabled={actionLoading !== null}
                  >
                    {actionLoading === 'resume' ? 'Resuming...' : '▶ Resume Run'}
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
            runStatus={run?.status}
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
