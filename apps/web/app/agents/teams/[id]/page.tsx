'use client'

import { AgentTeamDetail } from '@/components/Agents/AgentTeamDetail'

interface AgentTeamDetailPageParams {
  params: {
    id: string
  }
}

export default function AgentTeamDetailPage({ params }: AgentTeamDetailPageParams) {
  return <AgentTeamDetail teamId={params.id} />
}
