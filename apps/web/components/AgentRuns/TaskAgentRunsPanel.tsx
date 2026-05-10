'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { AgentRun, AgentTeam, getAgentRuns, getAgentTeams, createAgentRun, AgentRunStatus } from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'
import { RunStatusBadge } from './AgentRunBadges'
import { formatDate } from '@/lib/utils'

interface TaskAgentRunsPanelProps {
  taskId: string
}

export function TaskAgentRunsPanel({ taskId }: TaskAgentRunsPanelProps) {
  const [runs, setRuns] = useState<AgentRun[]>([])
  const [teams, setTeams] = useState<AgentTeam[]>([])
  const [loading, setLoading] = useState(true)
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const [selectedTeamId, setSelectedTeamId] = useState('')
  const [goal, setGoal] = useState('')
  const [creating, setCreating] = useState(false)

  useEffect(() => {
    async function fetchData() {
      try {
        const [runsData, teamsData] = await Promise.all([
          getAgentRuns(taskId),
          getAgentTeams(),
        ])
        setRuns(runsData)
        setTeams(teamsData)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load runs')
      } finally {
        setLoading(false)
      }
    }

    fetchData()
  }, [taskId])

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()

    setCreating(true)
    setError(null)

    try {
      const newRun = await createAgentRun(taskId, {
        team_id: selectedTeamId || undefined,
        goal: goal.trim() || undefined,
        status: 'draft',
      })

      setRuns((prev) => [newRun, ...prev])
      setShowCreateForm(false)
      setSelectedTeamId('')
      setGoal('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create run')
    } finally {
      setCreating(false)
    }
  }

  if (loading) {
    return (
      <Card>
        <CardContent className="py-8 text-center">
          <p className="text-slate-500 text-sm">Loading runs...</p>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-base">Agent Runs</CardTitle>
          <Button
            size="sm"
            variant="outline"
            onClick={() => setShowCreateForm(!showCreateForm)}
          >
            {showCreateForm ? 'Cancel' : '+ New Run'}
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {error && (
          <div className="p-3 bg-red-50 border border-red-200 rounded text-red-600 text-sm">
            {error}
          </div>
        )}

        {showCreateForm && (
          <form onSubmit={handleCreate} className="p-4 bg-slate-50 rounded-lg space-y-4">
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">
                Team
              </label>
              <select
                value={selectedTeamId}
                onChange={(e) => setSelectedTeamId(e.target.value)}
                className="w-full px-3 py-2 border border-slate-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                disabled={creating}
              >
                <option value="">No team</option>
                {teams.map((t) => (
                  <option key={t.id} value={t.id}>
                    {t.name}
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">
                Goal
              </label>
              <Textarea
                value={goal}
                onChange={(e) => setGoal(e.target.value)}
                placeholder="What should this run accomplish?"
                disabled={creating}
                rows={3}
              />
            </div>

            <div className="flex gap-3">
              <Button type="submit" disabled={creating}>
                {creating ? 'Creating...' : 'Create Run'}
              </Button>
              <Button
                type="button"
                variant="outline"
                onClick={() => setShowCreateForm(false)}
                disabled={creating}
              >
                Cancel
              </Button>
            </div>
          </form>
        )}

        {runs.length === 0 ? (
          <div className="py-6 text-center">
            <p className="text-slate-500 text-sm">No agent runs yet. Create one to get started.</p>
          </div>
        ) : (
          <div className="space-y-2">
            {runs.map((run) => {
              const team = teams.find((t) => t.id === run.team_id)
              return (
                <Link
                  key={run.id}
                  href={`/agent-runs/${run.id}`}
                  className="block p-3 border border-slate-200 rounded-lg hover:bg-slate-50 hover:border-slate-300 transition-colors"
                >
                  <div className="flex items-start justify-between gap-4">
                    <div className="min-w-0">
                      <div className="flex items-center gap-2 mb-1">
                        <RunStatusBadge status={run.status} />
                        {team && (
                          <span className="text-xs text-slate-500">
                            Team: {team.name}
                          </span>
                        )}
                      </div>
                      {run.goal && (
                        <p className="text-sm text-slate-600 line-clamp-2">{run.goal}</p>
                      )}
                      <p className="text-xs text-slate-400 mt-1">
                        {formatDate(run.created_at)}
                      </p>
                    </div>
                    <svg className="w-4 h-4 text-slate-400 mt-1 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                    </svg>
                  </div>
                </Link>
              )
            })}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
