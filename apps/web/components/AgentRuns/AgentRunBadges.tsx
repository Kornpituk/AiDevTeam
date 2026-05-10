'use client'

import { AgentRunStatus, AgentRunStepStatus } from '@/lib/api'
import { Badge } from '@/components/ui/badge'

export function getRunStatusVariant(status: AgentRunStatus): 'default' | 'success' | 'warning' | 'error' | 'info' {
  switch (status) {
    case 'completed':
    case 'approved':
      return 'success'
    case 'failed':
    case 'cancelled':
      return 'error'
    case 'running':
      return 'info'
    case 'waiting_approval':
    case 'paused':
      return 'warning'
    default:
      return 'default'
  }
}

export function getStepStatusVariant(status: AgentRunStepStatus): 'default' | 'success' | 'warning' | 'error' | 'info' {
  switch (status) {
    case 'completed':
      return 'success'
    case 'failed':
    case 'cancelled':
      return 'error'
    case 'running':
      return 'info'
    case 'waiting_approval':
      return 'warning'
    default:
      return 'default'
  }
}

export function formatRunStatus(status: AgentRunStatus): string {
  return status
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
}

export function formatStepStatus(status: AgentRunStepStatus): string {
  return status
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
}

interface RunStatusBadgeProps {
  status: AgentRunStatus
  className?: string
}

export function RunStatusBadge({ status, className }: RunStatusBadgeProps) {
  return (
    <Badge variant={getRunStatusVariant(status)} className={`capitalize ${className || ''}`}>
      {formatRunStatus(status)}
    </Badge>
  )
}

interface StepStatusBadgeProps {
  status: AgentRunStepStatus
  className?: string
}

export function StepStatusBadge({ status, className }: StepStatusBadgeProps) {
  return (
    <Badge variant={getStepStatusVariant(status)} className={`capitalize ${className || ''}`}>
      {formatStepStatus(status)}
    </Badge>
  )
}
