"use client";

import { useState } from "react";
import { Task, TaskStatus } from "@/lib/api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { StatusBadge } from "./StatusBadge";
import { StatusSelector } from "./StatusSelector";
import { formatDate } from "@/lib/utils";

interface TaskDetailProps {
  task: Task;
}

export function TaskDetail({ task: initialTask }: TaskDetailProps) {
  const [task, setTask] = useState<Task>(initialTask);

  function handleStatusChange(newStatus: TaskStatus) {
    setTask((prev) => ({
      ...prev,
      status: newStatus,
      updated_at: new Date().toISOString(),
    }));
  }

  return (
    <Card>
      <CardHeader className="pb-4">
        <div className="flex items-start justify-between gap-4">
          <div className="flex-1 min-w-0">
            <CardTitle className="text-2xl font-bold text-gray-900 mb-3">
              {task.title}
            </CardTitle>
            <StatusSelector
              taskId={task.id}
              currentStatus={task.status}
              onStatusChange={handleStatusChange}
            />
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-6">
        {task.description && (
          <div>
            <h4 className="text-sm font-semibold text-gray-500 uppercase tracking-wide mb-2">
              Description
            </h4>
            <div className="bg-gray-50 rounded-lg p-4">
              <p className="text-gray-700 whitespace-pre-wrap leading-relaxed">
                {task.description}
              </p>
            </div>
          </div>
        )}

        {task.plan && (
          <div>
            <h4 className="text-sm font-semibold text-gray-500 uppercase tracking-wide mb-2">
              Plan
            </h4>
            <div className="bg-gray-900 rounded-lg p-4 overflow-x-auto">
              <pre className="text-gray-100 text-sm font-mono whitespace-pre-wrap">
                {task.plan}
              </pre>
            </div>
          </div>
        )}

        {task.review_notes && (
          <div>
            <h4 className="text-sm font-semibold text-gray-500 uppercase tracking-wide mb-2">
              Review Notes
            </h4>
            <div className="bg-amber-50 border border-amber-200 rounded-lg p-4">
              <p className="text-amber-900 whitespace-pre-wrap leading-relaxed">
                {task.review_notes}
              </p>
            </div>
          </div>
        )}

        <div className="pt-6 border-t border-gray-200">
          <h4 className="text-sm font-semibold text-gray-500 uppercase tracking-wide mb-4">
            Details
          </h4>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div className="bg-gray-50 rounded-lg p-4">
              <p className="text-xs text-gray-500 uppercase tracking-wide mb-1">
                Created At
              </p>
              <p className="text-sm font-medium text-gray-900">
                {formatDate(task.created_at)}
              </p>
            </div>
            <div className="bg-gray-50 rounded-lg p-4">
              <p className="text-xs text-gray-500 uppercase tracking-wide mb-1">
                Last Updated
              </p>
              <p className="text-sm font-medium text-gray-900">
                {formatDate(task.updated_at)}
              </p>
            </div>
            <div className="bg-gray-50 rounded-lg p-4 sm:col-span-2">
              <p className="text-xs text-gray-500 uppercase tracking-wide mb-1">
                Task ID
              </p>
              <p className="text-sm font-mono text-gray-700 break-all">
                {task.id}
              </p>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
