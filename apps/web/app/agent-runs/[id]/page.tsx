'use client'

import { AgentRunDetail } from '@/components/AgentRuns/AgentRunDetail'

interface AgentRunDetailPageParams {
  params: {
    id: string
  }
}

export default function AgentRunDetailPage({ params }: AgentRunDetailPageParams) {
  return <AgentRunDetail runId={params.id} />
}
