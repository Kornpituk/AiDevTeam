'use client'

import { useState, useEffect } from 'react'
import {
  AgentToolCall,
  AgentToolCallStatus,
  getAgentToolCalls,
  AgentRunStep,
} from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { ToolCallStatusSelector } from './ToolCallStatusSelector'
import { AddToolCallForm } from './AddToolCallForm'
import { formatDate } from '@/lib/utils'

interface ToolCallsPanelProps {
  runId: string
  steps: AgentRunStep[]
}

export function ToolCallsPanel({ runId, steps }: ToolCallsPanelProps) {
  const [toolCalls, setToolCalls] = useState<AgentToolCall[]>([])
  const [loading, setLoading] = useState(true)
  const [showAddForm, setShowAddForm] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    async function fetchToolCalls() {
      try {
        const data = await getAgentToolCalls(runId)
        setToolCalls(data)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load tool calls')
      } finally {
        setLoading(false)
      }
    }

    fetchToolCalls()
  }, [runId])

  function handleToolCallAdded(toolCall: AgentToolCall) {
    setToolCalls((prev) => [...prev, toolCall])
    setShowAddForm(false)
  }

  function handleStatusChange(toolCallId: string, newStatus: AgentToolCallStatus) {
    setToolCalls((prev) =>
      prev.map((tc) => (tc.id === toolCallId ? { ...tc, status: newStatus } : tc))
    )
  }

  if (loading) {
    return (
      <Card>
        <CardContent className="py-6 text-center">
          <p className="text-slate-500 text-sm">Loading tool calls...</p>
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
          <CardTitle className="text-base">Tool Calls ({toolCalls.length})</CardTitle>
          <Button
            size="sm"
            variant="outline"
            onClick={() => setShowAddForm(!showAddForm)}
          >
            {showAddForm ? 'Cancel' : '+ Record Tool Call'}
          </Button>
        </div>
        </CardHeader>
      <CardContent className="space-y-4">
        {showAddForm && (
          <AddToolCallForm
            runId={runId}
            steps={steps}
            onAdd={handleToolCallAdded}
            onCancel={() => setShowAddForm(false)}
          />
        )}

        {toolCalls.length === 0 ? (
          <div className="py-6 text-center">
            <p className="text-slate-500 text-sm">No tool calls yet.</p>
          </div>
        ) : (
          <div className="space-y-3">
            {toolCalls.map((toolCall) => (
              <div
                key={toolCall.id}
                className="p-3 bg-slate-50 rounded-lg border border-slate-200"
              >
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <span className="text-sm font-medium text-slate-900">
                        {toolCall.tool_name}
                      </span>
                    </div>
                    <div className="flex items-center gap-3">
                      <ToolCallStatusSelector
                        toolCallId={toolCall.id}
                        currentStatus={toolCall.status}
                        onStatusChange={(newStatus) => handleStatusChange(toolCall.id, newStatus)}
                      />
                    </div>
                    {toolCall.input && Object.keys(toolCall.input).length > 0 && (
                      <details className="mt-2">
                        <summary className="text-xs text-slate-500 cursor-pointer hover:text-slate-700">
                          Input
                        </summary>
                        <pre className="mt-1 p-2 bg-slate-100 rounded text-xs text-slate-600 overflow-x-auto">
                          {JSON.stringify(toolCall.input, null, 2)}
                        </pre>
                      </details>
                    )}
                    {toolCall.output && Object.keys(toolCall.output).length > 0 && (
                      <details className="mt-2">
                        <summary className="text-xs text-slate-500 cursor-pointer hover:text-slate-700">
                          Output
                        </summary>
                        <pre className="mt-1 p-2 bg-slate-100 rounded text-xs text-slate-600 overflow-x-auto">
                          {JSON.stringify(toolCall.output, null, 2)}
                        </pre>
                      </details>
                    )}
                     <div className="flex items-center gap-4 mt-2 text-xs text-slate-500">
                       <span>Created: {formatDate(toolCall.created_at)}</span>
                       {toolCall.completed_at && (
                         <span>Completed: {formatDate(toolCall.completed_at)}</span>
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
