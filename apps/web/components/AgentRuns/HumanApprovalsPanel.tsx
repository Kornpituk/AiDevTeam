'use client'

import { useState, useEffect } from 'react'
import {
  HumanApproval,
  HumanApprovalStatus,
  getHumanApprovals,
  AgentRunStep,
} from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { ApprovalStatusSelector } from './ApprovalStatusSelector'
import { AddHumanApprovalForm } from './AddHumanApprovalForm'
import { formatDate } from '@/lib/utils'

interface HumanApprovalsPanelProps {
  runId: string
  steps: AgentRunStep[]
  runStatus?: string
}

export function HumanApprovalsPanel({ runId, steps, runStatus }: HumanApprovalsPanelProps) {
  const [approvals, setApprovals] = useState<HumanApproval[]>([])
  const [loading, setLoading] = useState(true)
  const [showAddForm, setShowAddForm] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    async function fetchApprovals() {
      try {
        const data = await getHumanApprovals(runId)
        setApprovals(data)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load approvals')
      } finally {
        setLoading(false)
      }
    }

    fetchApprovals()
  }, [runId])

  // Poll for new approvals while run is running
  useEffect(() => {
    if (runStatus !== 'running') return

    const intervalId = setInterval(async () => {
      try {
        const data = await getHumanApprovals(runId)
        setApprovals(data)
      } catch (err) {
        console.error('Approval polling error:', err)
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

        {approvals.length === 0 ? (
          <div className="py-6 text-center">
            <p className="text-slate-500 text-sm">No approvals yet.</p>
          </div>
        ) : (
          <div className="space-y-3">
            {approvals.map((approval) => (
              <div
              key={approval.id}
              className="p-3 bg-slate-50 rounded-lg border border-slate-200"
            >
              <div className="flex items-start justify-between gap-4">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    <span className="text-sm font-medium text-slate-900">
                      {approval.approval_type}
                    </span>
                  </div>
                  <div className="flex items-center gap-3">
                    <ApprovalStatusSelector
                      approvalId={approval.id}
                      currentStatus={approval.status}
                      onStatusChange={(newStatus) => handleStatusChange(approval.id, newStatus)}
                    />
                  </div>
                  {approval.request_notes && (
                    <p className="text-sm text-slate-600 mt-2">
                      Request: {approval.request_notes}
                    </p>
                  )}
                  {approval.decision_notes && (
                    <p className="text-sm text-slate-600 mt-1">
                      Decision: {approval.decision_notes}
                    </p>
                  )}
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
          ))}
        </div>
      )}
    </CardContent>
  </Card>
  )
}
