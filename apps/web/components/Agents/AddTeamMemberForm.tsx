'use client'

import { useState } from 'react'
import { AgentTeamMember, AgentProfile, addAgentTeamMember } from '@/lib/api'

interface AddTeamMemberFormProps {
  teamId: string
  profiles: AgentProfile[]
  onAdd: (member: AgentTeamMember) => void
  onCancel: () => void
}

export function AddTeamMemberForm({
  teamId,
  profiles,
  onAdd,
  onCancel,
}: AddTeamMemberFormProps) {
  const [selectedProfileId, setSelectedProfileId] = useState('')
  const [role, setRole] = useState('member')
  const [position, setPosition] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!selectedProfileId) return

    setLoading(true)
    setError(null)

    try {
      const positionNum = position.trim() !== '' ? parseInt(position, 10) : undefined
      const member = await addAgentTeamMember(teamId, {
        profile_id: selectedProfileId,
        member_role: role,
        position: positionNum,
      })

      onAdd(member)
      setSelectedProfileId('')
      setRole('member')
      setPosition('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to add member')
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="bg-slate-50 p-4 rounded-lg space-y-4">
      <h3 className="text-sm font-medium text-slate-900">Add Team Member</h3>

      {error && (
        <div className="p-3 bg-red-50 border border-red-200 rounded text-red-600 text-sm">
          {error}
        </div>
      )}

      <div className="grid grid-cols-3 gap-4">
        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">
            Agent Profile
          </label>
          <select
            value={selectedProfileId}
            onChange={(e) => setSelectedProfileId(e.target.value)}
            className="w-full px-3 py-2 border border-slate-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
            required
          >
            <option value="">Select a profile...</option>
            {profiles.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">
            Role
          </label>
          <select
            value={role}
            onChange={(e) => setRole(e.target.value)}
            className="w-full px-3 py-2 border border-slate-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
          >
            <option value="lead">Lead</option>
            <option value="member">Member</option>
            <option value="reviewer">Reviewer</option>
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">
            Position
          </label>
          <input
            type="number"
            value={position}
            onChange={(e) => setPosition(e.target.value)}
            placeholder="0"
            className="w-full px-3 py-2 border border-slate-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
          />
        </div>
      </div>

      <div className="flex gap-3">
        <button
          type="submit"
          disabled={loading || !selectedProfileId}
          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed text-sm"
        >
          {loading ? 'Adding...' : 'Add Member'}
        </button>
        <button
          type="button"
          onClick={onCancel}
          className="px-4 py-2 border border-slate-300 text-slate-700 rounded-md hover:bg-slate-50 text-sm"
        >
          Cancel
        </button>
      </div>
    </form>
  )
}
