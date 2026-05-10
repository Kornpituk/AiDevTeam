import Link from 'next/link'
import { AgentProfileForm } from '@/components/Agents/AgentProfileForm'
import { Button } from '@/components/ui/button'

export default function NewAgentProfilePage() {
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
        <h1 className="text-2xl font-bold text-gray-900">Create Agent Profile</h1>
        <p className="mt-1 text-sm text-gray-500">
          Define a new agent profile with tools and instructions
        </p>
      </div>

      <AgentProfileForm />
    </div>
  )
}
