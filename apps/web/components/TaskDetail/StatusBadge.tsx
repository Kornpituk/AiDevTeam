"use client";

import { TaskStatus } from "@/lib/api";
import { Badge, getStatusVariant, formatStatus } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

interface StatusBadgeProps {
  status: TaskStatus;
  className?: string;
}

export function StatusBadge({ status, className }: StatusBadgeProps) {
  return (
    <Badge variant={getStatusVariant(status)} className={cn("capitalize", className)}>
      {formatStatus(status)}
    </Badge>
  );
}
