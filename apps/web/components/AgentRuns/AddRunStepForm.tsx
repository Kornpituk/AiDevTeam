'use client'

import { useState } from 'react'
import { AgentProfile, createAgentRunStep, AgentRunStep, AgentRunStepStatus } from '@/lib/api'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'
import { Select } from '@/components/ui/select'

interface AddRunStepFormProps {
  runId: string
  profiles: AgentProfile[]
  onAdd: (step: AgentRunStep) => void
  onCancel: () => void
}

export function AddRunStepForm({ runId, profiles, onAdd, onCancel }: AddRunStepFormProps) {
  const [title, setTitle] = useState('')
  const [stepType, setStepType] = useState('')
  const [profileId, setProfileId] = useState('')
  const [instructions, setInstructions] = useState('')
  const [position, setPosition] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()

    if (!title.trim()) {
      setError('Title is required')
      return
    }
    if (!stepType.trim()) {
      setError('Step type is required')
      return
    }

    setLoading(true)
    setError(null)

    try {
      const positionNum = position.trim() !== '' ? parseInt(position, 10) : undefined
      const step = await createAgentRunStep(runId, {
        title: title.trim(),
        step_type: stepType.trim(),
        profile_id: profileId || undefined,
        instructions: instructions.trim() || undefined,
        position: positionNum,
        status: 'pending',
      })

      onAdd(step)
      setTitle('')
      setStepType('')
      setProfileId('')
      setInstructions('')
      setPosition('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to add step')
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="bg-slate-50 p-4 rounded-lg space-y-4">
      <h3 className="text-sm font-medium text-slate-900">Add Manual Step</h3>

      {error && (
        <div className="p-3 bg-red-50 border border-red-200 rounded text-red-600 text-sm">
          {error}
        </div>
      )}

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">
            Title <span className="text-red-500">*</span>
          </label>
          <Input
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="e.g., Review pull request"
            disabled={loading}
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">
            Step Type <span className="text-red-500">*</span>
          </label>
          <Input
            value={stepType}
            onChange={(e) => setStepType(e.target.value)}
            placeholder="e.g., review, plan, implement"
            disabled={loading}
          />
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">
            Agent Profile
          </label>
          <select
            value={profileId}
            onChange={(e) => setProfileId(e.target.value)}
            className="w-full px-3 py-2 border border-slate-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
            disabled={loading}
          >
            <option value="">No specific agent</option>
            {profiles.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">
            Position
          </label>
          <Input
            type="number"
            value={position}
            onChange={(e) => setPosition(e.target.value)}
            placeholder="0"
            disabled={loading}
          />
        </div>
      </div>

      <div>
        <label className="block text-sm font-medium text-slate-700 mb-1">
          Instructions
        </label>
        <Textarea
          value={instructions}
          onChange={(e) => setInstructions(e.target.value)}
          placeholder="Step-by-step instructions for this step..."
          disabled={loading}
          rows={3}
        />
      </div>

      <div className="flex gap-3">
        <Button type="submit" disabled={loading}>
          {loading ? 'Adding...' : 'Add Step'}
        </Button>
        <Button type="button" variant="outline" onClick={onCancel} disabled={loading}>
          Cancel
        </Button>
      </div>
    </form>
  )
}
