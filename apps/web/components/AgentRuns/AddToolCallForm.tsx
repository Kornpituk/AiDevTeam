'use client'

import { useState } from 'react'
import { AgentToolCall, createAgentToolCall, AgentRunStep } from '@/lib/api'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'

interface AddToolCallFormProps {
  runId: string
  steps: AgentRunStep[]
  onAdd: (toolCall: AgentToolCall) => void
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

export function AddToolCallForm({ runId, steps, onAdd, onCancel }: AddToolCallFormProps) {
  const [toolName, setToolName] = useState('')
  const [stepId, setStepId] = useState('')
  const [inputJson, setInputJson] = useState('')
  const [outputJson, setOutputJson] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()

    if (!toolName.trim()) {
      setError('Tool name is required')
      return
    }

    if (inputJson.trim() && !isValidJSON(inputJson)) {
      setError('Input must be valid JSON')
      return
    }

    if (outputJson.trim() && !isValidJSON(outputJson)) {
      setError('Output must be valid JSON')
      return
    }

    setLoading(true)
    setError(null)

    try {
      let input: Record<string, unknown> | undefined
      let output: Record<string, unknown> | undefined

      if (inputJson.trim()) {
        input = JSON.parse(inputJson) as Record<string, unknown>
      }
      if (outputJson.trim()) {
        output = JSON.parse(outputJson) as Record<string, unknown>
      }

      const toolCall = await createAgentToolCall(runId, {
        tool_name: toolName.trim(),
        step_id: stepId || undefined,
        input,
        output,
        status: 'recorded',
      })

      onAdd(toolCall)
      setToolName('')
      setStepId('')
      setInputJson('')
      setOutputJson('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to record tool call')
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="bg-slate-50 p-4 rounded-lg space-y-4">
      <h3 className="text-sm font-medium text-slate-900">Record Tool Call</h3>

      {error && (
        <div className="p-3 bg-red-50 border border-red-200 rounded text-red-600 text-sm">
          {error}
        </div>
      )}

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">
            Tool Name <span className="text-red-500">*</span>
          </label>
          <Input
            value={toolName}
            onChange={(e) => setToolName(e.target.value)}
            placeholder="e.g., get_file, edit_file, bash"
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
          Input (optional JSON)
        </label>
        <Textarea
          value={inputJson}
          onChange={(e) => setInputJson(e.target.value)}
          placeholder='{"path": "test.go", "content": "package main"}'
          disabled={loading}
          rows={3}
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-slate-700 mb-1">
          Output (optional JSON)
        </label>
        <Textarea
          value={outputJson}
          onChange={(e) => setOutputJson(e.target.value)}
          placeholder='{"success": true, "output": "..."}'
          disabled={loading}
          rows={3}
        />
      </div>

      <div className="flex gap-3">
        <Button type="submit" disabled={loading}>
          {loading ? 'Recording...' : 'Record Tool Call'}
        </Button>
        <Button type="button" variant="outline" onClick={onCancel} disabled={loading}>
          Cancel
        </Button>
      </div>
    </form>
  )
}
