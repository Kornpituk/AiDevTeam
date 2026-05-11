'use client'

import { useState } from 'react'
import { AgentMessage, AgentMessageRole, AGENT_MESSAGE_ROLES, createAgentMessage, AgentRunStep, AgentProfile } from '@/lib/api'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'
import { formatMessageRole } from './AgentRunBadges'

interface AddAgentMessageFormProps {
  runId: string
  steps: AgentRunStep[]
  profiles: AgentProfile[]
  onAdd: (message: AgentMessage) => void
  onCancel: () => void
}

function isValidJSON(str: string): boolean {
  if (!str.trim()) return true
  try {
    JSON.parse(str)
    return true
  } catch {
    return false
  }
}

export function AddAgentMessageForm({ runId, steps, profiles, onAdd, onCancel }: AddAgentMessageFormProps) {
  const [role, setRole] = useState<AgentMessageRole>('user')
  const [content, setContent] = useState('')
  const [stepId, setStepId] = useState('')
  const [profileId, setProfileId] = useState('')
  const [metadataJson, setMetadataJson] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()

    if (!content.trim()) {
      setError('Content is required')
      return
    }

    if (metadataJson.trim() && !isValidJSON(metadataJson)) {
      setError('Metadata must be valid JSON')
      return
    }

    setLoading(true)
    setError(null)

    try {
      let metadata: Record<string, unknown> | undefined
      if (metadataJson.trim()) {
        metadata = JSON.parse(metadataJson) as Record<string, unknown>
      }

      const message = await createAgentMessage(runId, {
        role,
        content: content.trim(),
        step_id: stepId || undefined,
        profile_id: profileId || undefined,
        metadata,
      })

      onAdd(message)
      setRole('user')
      setContent('')
      setStepId('')
      setProfileId('')
      setMetadataJson('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to add message')
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="bg-slate-50 p-4 rounded-lg space-y-4">
      <h3 className="text-sm font-medium text-slate-900">Add Message</h3>

      {error && (
        <div className="p-3 bg-red-50 border border-red-200 rounded text-red-600 text-sm">
          {error}
        </div>
      )}

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">
            Role <span className="text-red-500">*</span>
          </label>
          <select
            value={role}
            onChange={(e) => setRole(e.target.value as AgentMessageRole)}
            className="w-full px-3 py-2 border border-slate-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
            disabled={loading}
          >
            {AGENT_MESSAGE_ROLES.map((r) => (
              <option key={r} value={r}>
                {formatMessageRole(r)}
              </option>
            ))}
          </select>
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
          Content <span className="text-red-500">*</span>
        </label>
        <Textarea
          value={content}
          onChange={(e) => setContent(e.target.value)}
          placeholder="Message content..."
          disabled={loading}
          rows={4}
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-slate-700 mb-1">
          Metadata (optional JSON)
        </label>
        <Textarea
          value={metadataJson}
          onChange={(e) => setMetadataJson(e.target.value)}
          placeholder='{"key": "value"}'
          disabled={loading}
          rows={2}
        />
      </div>

      <div className="flex gap-3">
        <Button type="submit" disabled={loading}>
          {loading ? 'Adding...' : 'Add Message'}
        </Button>
        <Button type="button" variant="outline" onClick={onCancel} disabled={loading}>
          Cancel
        </Button>
      </div>
    </form>
  )
}
