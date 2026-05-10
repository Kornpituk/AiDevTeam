"use client";

import { useState } from "react";
import { TaskStatus, TASK_STATUSES, updateTaskStatus } from "@/lib/api";
import { Select } from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Badge, getStatusVariant, formatStatus } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

interface StatusSelectorProps {
  taskId: string;
  currentStatus: TaskStatus;
  onStatusChange?: (newStatus: TaskStatus) => void;
}

export function StatusSelector({ taskId, currentStatus, onStatusChange }: StatusSelectorProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [selectedStatus, setSelectedStatus] = useState<TaskStatus>(currentStatus);
  const [isLoading, setIsLoading] = useState(false);

  async function handleSave() {
    if (selectedStatus === currentStatus) {
      setIsEditing(false);
      return;
    }

    setIsLoading(true);
    try {
      await updateTaskStatus(taskId, selectedStatus);
      onStatusChange?.(selectedStatus);
      setIsEditing(false);
    } catch (e) {
      console.error("Failed to update status:", e);
      setSelectedStatus(currentStatus);
    } finally {
      setIsLoading(false);
    }
  }

  function handleCancel() {
    setSelectedStatus(currentStatus);
    setIsEditing(false);
  }

  if (!isEditing) {
    return (
      <button
        onClick={() => setIsEditing(true)}
        className="group inline-flex items-center"
        title="Click to change status"
      >
        <Badge variant={getStatusVariant(currentStatus)} className="capitalize group-hover:opacity-80 transition-opacity">
          {formatStatus(currentStatus)}
        </Badge>
        <svg className="w-4 h-4 ml-1.5 text-gray-400 group-hover:text-gray-600 transition-colors" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
        </svg>
      </button>
    );
  }

  return (
    <div className={cn("flex items-center gap-2 flex-wrap")}>
      <Select
        value={selectedStatus}
        onChange={(e) => setSelectedStatus(e.target.value as TaskStatus)}
        disabled={isLoading}
        className="w-auto min-w-36 text-sm"
      >
        {TASK_STATUSES.map((status) => (
          <option key={status} value={status}>
            {formatStatus(status)}
          </option>
        ))}
      </Select>
      <Button size="sm" onClick={handleSave} disabled={isLoading}>
        {isLoading ? (
          <svg className="animate-spin h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
          </svg>
        ) : (
          "Save"
        )}
      </Button>
      <Button size="sm" variant="ghost" onClick={handleCancel} disabled={isLoading}>
        Cancel
      </Button>
    </div>
  );
}
