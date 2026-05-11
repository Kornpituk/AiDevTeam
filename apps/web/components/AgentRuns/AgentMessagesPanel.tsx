'use client'

import { useState, useEffect } from 'react'
import {
  AgentMessage,
  getAgentMessages,
  AgentRunStep,
  AgentProfile,
} from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { MessageRoleBadge, formatMessageRole } from './AgentRunBadges'
import { AddAgentMessageForm } from './AddAgentMessageForm'
import { formatDate } from '@/lib/utils'

interface AgentMessagesPanelProps {
  runId: string
  steps: AgentRunStep[]
  profiles: AgentProfile[]
}

export function AgentMessagesPanel({ runId, steps, profiles }: AgentMessagesPanelProps) {
  const [messages, setMessages] = useState<AgentMessage[]>([])
  const [loading, setLoading] = useState(true)
  const [showAddForm, setShowAddForm] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [refreshKey, setRefreshKey] = useState(0)

  useEffect(() => {
    async function fetchMessages() {
      setLoading(true)
      try {
        const data = await getAgentMessages(runId)
        setMessages(data)
        setError(null)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load messages')
      } finally {
        setLoading(false)
      }
    }
    fetchMessages()
  }, [runId, refreshKey])

  function handleRefresh() {
    setRefreshKey((k) => k + 1)
  }

  function handleMessageAdded(message: AgentMessage) {
    setMessages((prev) => [...prev, message])
    setShowAddForm(false)
  }

  if (loading) {
    return (
      <Card>
        <CardContent className="py-6 text-center">
          <p className="text-slate-500 text-sm">Loading messages...</p>
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
        <div className="flex items-center justify-between gap-2">
          <CardTitle className="text-base">Messages ({messages.length})</CardTitle>
          <div className="flex gap-2">
            <Button size="sm" variant="outline" onClick={handleRefresh} disabled={loading}>
              Refresh
            </Button>
            <Button
              size="sm"
              variant="outline"
              onClick={() => setShowAddForm(!showAddForm)}
            >
              {showAddForm ? 'Cancel' : '+ Add Message'}
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {showAddForm && (
          <AddAgentMessageForm
            runId={runId}
            steps={steps}
            profiles={profiles}
            onAdd={handleMessageAdded}
            onCancel={() => setShowAddForm(false)}
          />
        )}

        {messages.length === 0 ? (
          <div className="py-6 text-center">
            <p className="text-slate-500 text-sm">No messages yet.</p>
          </div>
        ) : (
          <div className="space-y-3">
            {messages.map((message) => {
              const profile = profiles.find((p) => p.id === message.profile_id)
              return (
                <div
                  key={message.id}
                  className="p-3 bg-slate-50 rounded-lg border border-slate-200"
                >
                  <div className="flex items-start justify-between gap-2 mb-2">
                    <div className="flex items-center gap-2">
                      <MessageRoleBadge role={message.role} />
                      {profile && (
                        <span className="text-xs text-slate-500">
                          Agent: {profile.name}
                        </span>
                      )}
                    </div>
                    <span className="text-xs text-slate-400">
                      {formatDate(message.created_at)}
                    </span>
                  </div>
                  <div className="text-sm text-slate-700 whitespace-pre-wrap">
                    {message.content}
                  </div>
                  {message.metadata && Object.keys(message.metadata).length > 0 && (
                    <details className="mt-2">
                      <summary className="text-xs text-slate-500 cursor-pointer hover:text-slate-700">
                        View metadata
                      </summary>
                      <pre className="mt-2 p-2 bg-slate-100 rounded text-xs text-slate-600 overflow-x-auto">
                        {JSON.stringify(message.metadata, null, 2)}
                      </pre>
                    </details>
                  )}
                </div>
              )
            })}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
