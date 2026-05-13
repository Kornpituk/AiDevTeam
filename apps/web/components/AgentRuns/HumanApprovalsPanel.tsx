'use client'

import { useState, useEffect } from 'react'
import {
  HumanApproval,
  HumanApprovalStatus,
  getHumanApprovals,
  AgentToolCall,
  getAgentToolCalls,
  updateHumanApprovalStatus,
  AgentRunStep,
} from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { ApprovalStatusSelector } from './ApprovalStatusSelector'
import { ApprovalStatusBadge } from './AgentRunBadges'
import { AddHumanApprovalForm } from './AddHumanApprovalForm'
import { formatDate } from '@/lib/utils'

interface HumanApprovalsPanelProps {
  runId: string
  steps: AgentRunStep[]
  runStatus?: string
}

function formatJson(str: string): string {
  try {
    return JSON.stringify(JSON.parse(str), null, 2)
  } catch {
    return str
  }
}

function extractToolName(approvalType: string): string {
  if (approvalType.startsWith('tool:')) {
    return approvalType.slice(5)
  }
  return approvalType
}

function Spinner() {
  return (
    <svg className="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
    </svg>
  )
}

export function HumanApprovalsPanel({ runId, steps, runStatus }: HumanApprovalsPanelProps) {
  const [approvals, setApprovals] = useState<HumanApproval[]>([])
  const [toolCallMap, setToolCallMap] = useState<Record<string, AgentToolCall>>({})
  const [loading, setLoading] = useState(true)
  const [showAddForm, setShowAddForm] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [actionLoadingId, setActionLoadingId] = useState<string | null>(null)
  const [actionTarget, setActionTarget] = useState<HumanApprovalStatus | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)

  useEffect(() => {
    async function fetchData() {
      try {
        const [approvalsData, toolCallsData] = await Promise.all([
          getHumanApprovals(runId),
          getAgentToolCalls(runId),
        ])
        setApprovals(approvalsData)
        const map: Record<string, AgentToolCall> = {}
        toolCallsData.forEach((tc) => {
          map[tc.id] = tc
        })
        setToolCallMap(map)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load approvals')
      } finally {
        setLoading(false)
      }
    }

    fetchData()
  }, [runId])

  // Poll for new approvals while run is running
  useEffect(() => {
    if (runStatus !== 'running') return

    const intervalId = setInterval(async () => {
      try {
        const [approvalsData, toolCallsData] = await Promise.all([
          getHumanApprovals(runId),
          getAgentToolCalls(runId),
        ])
        setApprovals(approvalsData)
        const map: Record<string, AgentToolCall> = {}
        toolCallsData.forEach((tc) => {
          map[tc.id] = tc
        })
        setToolCallMap(map)
      } catch (err) {
        console.error('Polling error:', err)
      }
    }, 3000)

    return () => clearInterval(intervalId)
  }, [runId, runStatus])

  function handleApprovalAdded(approval: HumanApproval) {
    setApprovals((prev) => [...prev, approval])
    setShowAddForm(false)
  }

  function handleStatusChange(approvalId: string, newStatus: HumanApprovalStatus) {
    setApprovals((prev) =>
      prev.map((a) => (a.id === approvalId ? { ...a, status: newStatus } : a))
    )
  }

  async function handleAction(approvalId: string, status: HumanApprovalStatus) {
    setActionLoadingId(approvalId)
    setActionTarget(status)
    setActionError(null)
    try {
      await updateHumanApprovalStatus(approvalId, status)
      setApprovals((prev) =>
        prev.map((a) => (a.id === approvalId ? { ...a, status } : a))
      )
    } catch (err) {
      setActionError(err instanceof Error ? err.message : 'Failed to update status')
    } finally {
      setActionLoadingId(null)
      setActionTarget(null)
    }
  }

  if (loading) {
    return (
      <Card>
        <CardContent className="py-6 text-center">
          <p className="text-slate-500 text-sm">Loading approvals...</p>
        </CardContent>
      </Card>
    )
  }

  if (error) {
    return (
      <Card className="border-red-200 bg-red-50">
        <CardContent className="py-4">
          <p className="text-red-600 text-sm">{error}</p>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-base">Approvals ({approvals.length})</CardTitle>
          <Button
            size="sm"
            variant="outline"
            onClick={() => setShowAddForm(!showAddForm)}
          >
            {showAddForm ? 'Cancel' : '+ Request Approval'}
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {showAddForm && (
          <AddHumanApprovalForm
            runId={runId}
            steps={steps}
            onAdd={handleApprovalAdded}
            onCancel={() => setShowAddForm(false)}
          />
        )}

        {actionError && (
          <div className="p-3 bg-red-50 border border-red-200 rounded text-red-600 text-sm">
            {actionError}
          </div>
        )}

        {approvals.length === 0 ? (
          <div className="py-6 text-center">
            <p className="text-slate-500 text-sm">No approvals yet.</p>
          </div>
        ) : (
          <div className="space-y-3">
            {approvals.map((approval) => {
              const toolName = extractToolName(approval.approval_type)
              const toolCall = approval.tool_call_id
                ? toolCallMap[approval.tool_call_id]
                : undefined
              const isLoading = actionLoadingId === approval.id

              return (
                <div
                  key={approval.id}
                  className="p-3 bg-slate-50 rounded-lg border border-slate-200"
                >
                  <div className="flex items-start justify-between gap-4">
                    <div className="flex-1 min-w-0">
                      {/* Header: tool name + status */}
                      <div className="flex items-center gap-2 mb-1">
                        <span className="text-sm font-medium text-slate-900">
                          {toolName}
                        </span>
                        <ApprovalStatusBadge status={approval.status} />
                      </div>

                      {/* Pending: big action buttons */}
                      {approval.status === 'pending' && (
                        <div className="flex items-center gap-2 mt-2">
                          <Button
                            size="sm"
                            variant="primary"
                            className="bg-green-600 hover:bg-green-700 text-white"
                            onClick={() => handleAction(approval.id, 'approved')}
                            disabled={isLoading}
                          >
                            {isLoading && actionTarget === 'approved' ? (
                              <span className="flex items-center gap-1">
                                <Spinner /> Approving...
                              </span>
                            ) : (
                              '✅ Approve'
                            )}
                          </Button>
                          <Button
                            size="sm"
                            variant="primary"
                            className="bg-red-600 hover:bg-red-700 text-white"
                            onClick={() => handleAction(approval.id, 'rejected')}
                            disabled={isLoading}
                          >
                            {isLoading && actionTarget === 'rejected' ? (
                              <span className="flex items-center gap-1">
                                <Spinner /> Rejecting...
                              </span>
                            ) : (
                              '❌ Reject'
                            )}
                          </Button>

                          {/* Also keep the dropdown for changing to cancelled */}
                          <ApprovalStatusSelector
                            approvalId={approval.id}
                            currentStatus={approval.status}
                            onStatusChange={(newStatus) =>
                              handleStatusChange(approval.id, newStatus)
                            }
                          />
                        </div>
                      )}

                      {/* Non-pending: show status and dropdown */}
                      {approval.status !== 'pending' && (
                        <div className="flex items-center gap-3 mt-1">
                          <ApprovalStatusSelector
                            approvalId={approval.id}
                            currentStatus={approval.status}
                            onStatusChange={(newStatus) =>
                              handleStatusChange(approval.id, newStatus)
                            }
                          />
                        </div>
                      )}

                      {/* Request notes as expandable JSON */}
                      {approval.request_notes && (
                        <details className="mt-2">
                          <summary className="text-xs text-slate-500 cursor-pointer hover:text-slate-700">
                            Tool Input
                          </summary>
                          <pre className="mt-1 p-2 bg-slate-100 rounded text-xs text-slate-600 overflow-x-auto">
                            {formatJson(approval.request_notes)}
                          </pre>
                        </details>
                      )}

                      {/* Decision notes */}
                      {approval.decision_notes && (
                        <p className="text-xs text-slate-600 mt-2">
                          Decision: {approval.decision_notes}
                        </p>
                      )}

                      {/* Tool call result (for approved with tool_call_id) */}
                      {approval.status === 'approved' && toolCall && (
                        <details className="mt-2">
                          <summary className="text-xs text-slate-500 cursor-pointer hover:text-slate-700">
                            Tool Result
                          </summary>
                          <pre className="mt-1 p-2 bg-green-50 rounded text-xs text-slate-600 overflow-x-auto">
                            {toolCall.output
                              ? JSON.stringify(toolCall.output, null, 2)
                              : 'No output'}
                          </pre>
                        </details>
                      )}

                      {/* Dates and metadata */}
                      <div className="flex items-center gap-4 mt-2 text-xs text-slate-500">
                        <span>Created: {formatDate(approval.created_at)}</span>
                        {approval.decided_at && (
                          <span>Decided: {formatDate(approval.decided_at)}</span>
                        )}
                        {approval.requested_by && (
                          <span>Requested by: {approval.requested_by}</span>
                        )}
                        {approval.decided_by && (
                          <span>Decided by: {approval.decided_by}</span>
                        )}
                      </div>
                    </div>
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
