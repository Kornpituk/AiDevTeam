'use client'

import { AgentRunStatus, AgentRunStepStatus, HumanApprovalStatus, AgentToolCallStatus, AgentMessageRole } from '@/lib/api'
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

export function getApprovalStatusVariant(status: HumanApprovalStatus): 'default' | 'success' | 'warning' | 'error' | 'info' {
  switch (status) {
    case 'approved':
      return 'success'
    case 'rejected':
    case 'cancelled':
      return 'error'
    case 'pending':
      return 'warning'
    default:
      return 'default'
  }
}

export function getToolCallStatusVariant(status: AgentToolCallStatus): 'default' | 'success' | 'warning' | 'error' | 'info' {
  switch (status) {
    case 'completed':
    case 'approved':
      return 'success'
    case 'failed':
    case 'rejected':
      return 'error'
    case 'recorded':
      return 'info'
    default:
      return 'default'
  }
}

export function getMessageRoleVariant(role: AgentMessageRole): 'default' | 'success' | 'warning' | 'error' | 'info' {
  switch (role) {
    case 'assistant':
      return 'info'
    case 'user':
      return 'success'
    case 'system':
      return 'warning'
    case 'tool':
      return 'default'
    case 'reviewer':
      return 'success'
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

export function formatApprovalStatus(status: HumanApprovalStatus): string {
  return status
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
}

export function formatToolCallStatus(status: AgentToolCallStatus): string {
  return status
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
}

export function formatMessageRole(role: AgentMessageRole): string {
  return role.charAt(0).toUpperCase() + role.slice(1)
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

interface ApprovalStatusBadgeProps {
  status: HumanApprovalStatus
  className?: string
}

export function ApprovalStatusBadge({ status, className }: ApprovalStatusBadgeProps) {
  return (
    <Badge variant={getApprovalStatusVariant(status)} className={`capitalize ${className || ''}`}>
      {formatApprovalStatus(status)}
    </Badge>
  )
}

interface ToolCallStatusBadgeProps {
  status: AgentToolCallStatus
  className?: string
}

export function ToolCallStatusBadge({ status, className }: ToolCallStatusBadgeProps) {
  return (
    <Badge variant={getToolCallStatusVariant(status)} className={`capitalize ${className || ''}`}>
      {formatToolCallStatus(status)}
    </Badge>
  )
}

interface MessageRoleBadgeProps {
  role: AgentMessageRole
  className?: string
}

export function MessageRoleBadge({ role, className }: MessageRoleBadgeProps) {
  return (
    <Badge variant={getMessageRoleVariant(role)} className={`capitalize ${className || ''}`}>
      {formatMessageRole(role)}
    </Badge>
  )
}
