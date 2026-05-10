import Link from 'next/link'
import { AgentTeamForm } from '@/components/Agents/AgentTeamForm'
import { Button } from '@/components/ui/button'

export default function NewAgentTeamPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link href="/agents">
          <Button variant="ghost" size="sm">
            <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
            Back to Agents
          </Button>
        </Link>
      </div>

      <div>
        <h1 className="text-2xl font-bold text-gray-900">Create Agent Team</h1>
        <p className="mt-1 text-sm text-gray-500">
          Create a new team of agents
        </p>
      </div>

      <AgentTeamForm />
    </div>
  )
}
