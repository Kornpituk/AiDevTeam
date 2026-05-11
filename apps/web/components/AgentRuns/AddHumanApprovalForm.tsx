'use client'

import { useState } from 'react'
import { HumanApproval, createHumanApproval, AgentRunStep } from '@/lib/api'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'

interface AddHumanApprovalFormProps {
  runId: string
  steps: AgentRunStep[]
  onAdd: (approval: HumanApproval) => void
  onCancel: () => void
}

export function AddHumanApprovalForm({ runId, steps, onAdd, onCancel }: AddHumanApprovalFormProps) {
  const [approvalType, setApprovalType] = useState('')
  const [stepId, setStepId] = useState('')
  const [requestNotes, setRequestNotes] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()

    if (!approvalType.trim()) {
      setError('Approval type is required')
      return
    }

    setLoading(true)
    setError(null)

    try {
      const approval = await createHumanApproval(runId, {
        approval_type: approvalType.trim(),
        step_id: stepId || undefined,
        request_notes: requestNotes.trim() || undefined,
        status: 'pending',
      })

      onAdd(approval)
      setApprovalType('')
      setStepId('')
      setRequestNotes('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create approval')
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="bg-slate-50 p-4 rounded-lg space-y-4">
      <h3 className="text-sm font-medium text-slate-900">Request Approval</h3>

      {error && (
        <div className="p-3 bg-red-50 border border-red-200 rounded text-red-600 text-sm">
          {error}
        </div>
      )}

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">
            Approval Type <span className="text-red-500">*</span>
          </label>
          <Input
            value={approvalType}
            onChange={(e) => setApprovalType(e.target.value)}
            placeholder="e.g., step_execution, run_completion"
            disabled={loading}
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">
            Step
          </label>
          <select
            value={stepId}
            onChange={(e) => setStepId(e.target.value)}
            className="w-full px-3 py-2 border border-slate-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
            disabled={loading}
          >
            <option value="">No specific step</option>
            {steps.map((s) => (
              <option key={s.id} value={s.id}>
                {s.title}
              </option>
            ))}
          </select>
        </div>
      </div>

      <div>
        <label className="block text-sm font-medium text-slate-700 mb-1">
          Request Notes
        </label>
        <Textarea
          value={requestNotes}
          onChange={(e) => setRequestNotes(e.target.value)}
          placeholder="What requires approval?"
          disabled={loading}
          rows={3}
        />
      </div>

      <div className="flex gap-3">
        <Button type="submit" disabled={loading}>
          {loading ? 'Creating...' : 'Request Approval'}
        </Button>
        <Button type="button" variant="outline" onClick={onCancel} disabled={loading}>
          Cancel
        </Button>
      </div>
    </form>
  )
}
