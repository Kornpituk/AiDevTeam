'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import {
  AgentTeam,
  AgentTeamMember,
  AgentProfile,
  getAgentTeam,
  getAgentTeamMembers,
  getAgentProfiles,
} from '@/lib/api'
import { AddTeamMemberForm } from './AddTeamMemberForm'

interface AgentTeamDetailProps {
  teamId: string
}

export function AgentTeamDetail({ teamId }: AgentTeamDetailProps) {
  const [team, setTeam] = useState<AgentTeam | null>(null)
  const [members, setMembers] = useState<AgentTeamMember[]>([])
  const [profiles, setProfiles] = useState<AgentProfile[]>([])
  const [loading, setLoading] = useState(true)
  const [showAddForm, setShowAddForm] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    async function fetchData() {
      try {
        const [teamData, membersData, profilesData] = await Promise.all([
          getAgentTeam(teamId),
          getAgentTeamMembers(teamId),
          getAgentProfiles(),
        ])

        setTeam(teamData)
        setMembers(membersData)
        setProfiles(profilesData)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load team')
      } finally {
        setLoading(false)
      }
    }

    fetchData()
  }, [teamId])

  function handleMemberAdded(member: AgentTeamMember) {
    setMembers((prev) => [...prev, member])
    setShowAddForm(false)
  }

  if (loading) {
    return <div className="text-slate-500">Loading team...</div>
  }

  if (error) {
    return (
      <div className="p-4 bg-red-50 border border-red-200 rounded-lg text-red-600">
        {error}
      </div>
    )
  }

  if (!team) {
    return <div className="text-slate-500">Team not found</div>
  }

  const availableProfiles = profiles.filter(
    (p) => !members.some((m) => m.profile_id === p.id)
  )

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link href="/agents" className="text-blue-600 hover:underline text-sm">
          ← Back to Agents
        </Link>
      </div>

      <div className="bg-white border border-slate-200 rounded-lg p-6">
        <div className="flex items-start justify-between">
          <div>
            <h1 className="text-xl font-bold text-slate-900">{team.name}</h1>
            {team.description && (
              <p className="mt-2 text-slate-600">{team.description}</p>
            )}
            <p className="mt-3 text-sm text-slate-500">
              Created: {new Date(team.created_at).toLocaleDateString()}
            </p>
          </div>
        </div>
      </div>

      <div className="bg-white border border-slate-200 rounded-lg">
        <div className="px-6 py-4 border-b border-slate-200 flex items-center justify-between">
          <h2 className="text-lg font-semibold text-slate-900">Team Members</h2>
          {availableProfiles.length > 0 && (
            <button
              onClick={() => setShowAddForm(true)}
              className="px-3 py-1.5 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm"
            >
              + Add Member
            </button>
          )}
        </div>

        {showAddForm && (
          <div className="px-6 py-4 border-b border-slate-200">
            <AddTeamMemberForm
              teamId={teamId}
              profiles={availableProfiles}
              onAdd={handleMemberAdded}
              onCancel={() => setShowAddForm(false)}
            />
          </div>
        )}

        <div className="divide-y divide-slate-100">
          {members.length === 0 ? (
            <div className="px-6 py-8 text-center text-slate-500">
              <p>No members yet.</p>
              {availableProfiles.length === 0 && (
                <p className="mt-1 text-sm">
                  No agent profiles available.{' '}
                  <Link href="/agents/profiles/new" className="text-blue-600 hover:underline">
                    Create a profile
                  </Link>
                  .
                </p>
              )}
            </div>
          ) : (
            members.map((member) => {
              const profile = profiles.find((p) => p.id === member.profile_id)
              return (
                <div
                  key={member.id}
                  className="px-6 py-4 flex items-center justify-between hover:bg-slate-50"
                >
                  <div>
                    <div className="font-medium text-slate-900">
                      {profile?.name || member.profile_id}
                    </div>
                    {profile?.description && (
                      <div className="text-sm text-slate-500 mt-1">
                        {profile.description}
                      </div>
                    )}
                  </div>
                  <span
                    className={`px-2 py-1 rounded text-xs font-medium ${
                      member.member_role === 'lead'
                        ? 'bg-purple-100 text-purple-700'
                        : member.member_role === 'reviewer'
                        ? 'bg-amber-100 text-amber-700'
                        : 'bg-slate-100 text-slate-600'
                    }`}
                  >
                    {member.member_role}
                  </span>
                </div>
              )
            })
          )}
        </div>
      </div>
    </div>
  )
}
