"use client";

import { useState, useCallback } from "react";
import { Task, TaskStatus, updateTaskPlan, updateTaskReviewNotes } from "@/lib/api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { StatusBadge } from "./StatusBadge";
import { StatusSelector } from "./StatusSelector";
import { formatDate } from "@/lib/utils";

interface TaskDetailProps {
  task: Task;
}

function EditablePlanSection({
  taskId,
  initialPlan,
  onUpdate,
}: {
  taskId: string;
  initialPlan: string;
  onUpdate: (task: Task) => void;
}) {
  const [isEditing, setIsEditing] = useState(false);
  const [planValue, setPlanValue] = useState(initialPlan || "");
  const [isSaving, setIsSaving] = useState(false);

  const handleSave = useCallback(async () => {
    setIsSaving(true);
    try {
      const updated = await updateTaskPlan(taskId, planValue);
      onUpdate(updated);
      setIsEditing(false);
    } catch (e) {
      console.error("Failed to update plan:", e);
    } finally {
      setIsSaving(false);
    }
  }, [taskId, planValue, onUpdate]);

  const handleCancel = useCallback(() => {
    setPlanValue(initialPlan || "");
    setIsEditing(false);
  }, [initialPlan]);

  if (isEditing) {
    return (
      <div>
        <div className="flex items-center justify-between mb-2">
          <h4 className="text-sm font-semibold text-gray-500 uppercase tracking-wide">
            Plan
          </h4>
          <div className="flex gap-2">
            <Button variant="ghost" size="sm" onClick={handleCancel} disabled={isSaving}>
              Cancel
            </Button>
            <Button size="sm" onClick={handleSave} disabled={isSaving}>
              {isSaving ? "Saving..." : "Save"}
            </Button>
          </div>
        </div>
        <textarea
          className="w-full h-48 bg-gray-900 text-gray-100 text-sm font-mono rounded-lg p-4 resize-y focus:outline-none focus:ring-2 focus:ring-blue-500"
          value={planValue}
          onChange={(e) => setPlanValue(e.target.value)}
          placeholder="Enter your plan here..."
        />
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-2">
        <h4 className="text-sm font-semibold text-gray-500 uppercase tracking-wide">
          Plan
        </h4>
        <Button variant="ghost" size="sm" onClick={() => setIsEditing(true)}>
          Edit
        </Button>
      </div>
      {initialPlan ? (
        <div className="bg-gray-900 rounded-lg p-4 overflow-x-auto">
          <pre className="text-gray-100 text-sm font-mono whitespace-pre-wrap">
            {initialPlan}
          </pre>
        </div>
      ) : (
        <div className="bg-gray-50 border border-dashed border-gray-300 rounded-lg p-6 text-center">
          <p className="text-gray-500 text-sm mb-3">No plan yet. Add one to track your implementation strategy.</p>
          <Button variant="outline" size="sm" onClick={() => setIsEditing(true)}>
            Add Plan
          </Button>
        </div>
      )}
    </div>
  );
}

function EditableReviewNotesSection({
  taskId,
  initialNotes,
  onUpdate,
}: {
  taskId: string;
  initialNotes: string;
  onUpdate: (task: Task) => void;
}) {
  const [isEditing, setIsEditing] = useState(false);
  const [notesValue, setNotesValue] = useState(initialNotes || "");
  const [isSaving, setIsSaving] = useState(false);

  const handleSave = useCallback(async () => {
    setIsSaving(true);
    try {
      const updated = await updateTaskReviewNotes(taskId, notesValue);
      onUpdate(updated);
      setIsEditing(false);
    } catch (e) {
      console.error("Failed to update review notes:", e);
    } finally {
      setIsSaving(false);
    }
  }, [taskId, notesValue, onUpdate]);

  const handleCancel = useCallback(() => {
    setNotesValue(initialNotes || "");
    setIsEditing(false);
  }, [initialNotes]);

  if (isEditing) {
    return (
      <div>
        <div className="flex items-center justify-between mb-2">
          <h4 className="text-sm font-semibold text-gray-500 uppercase tracking-wide">
            Review Notes
          </h4>
          <div className="flex gap-2">
            <Button variant="ghost" size="sm" onClick={handleCancel} disabled={isSaving}>
              Cancel
            </Button>
            <Button size="sm" onClick={handleSave} disabled={isSaving}>
              {isSaving ? "Saving..." : "Save"}
            </Button>
          </div>
        </div>
        <textarea
          className="w-full h-32 bg-amber-50 text-amber-900 text-sm rounded-lg p-4 resize-y focus:outline-none focus:ring-2 focus:ring-amber-500 border border-amber-200"
          value={notesValue}
          onChange={(e) => setNotesValue(e.target.value)}
          placeholder="Enter review notes here..."
        />
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-2">
        <h4 className="text-sm font-semibold text-gray-500 uppercase tracking-wide">
          Review Notes
        </h4>
        <Button variant="ghost" size="sm" onClick={() => setIsEditing(true)}>
          Edit
        </Button>
      </div>
      {initialNotes ? (
        <div className="bg-amber-50 border border-amber-200 rounded-lg p-4">
          <p className="text-amber-900 whitespace-pre-wrap leading-relaxed">
            {initialNotes}
          </p>
        </div>
      ) : (
        <div className="bg-amber-50 border border-dashed border-amber-300 rounded-lg p-6 text-center">
          <p className="text-amber-700 text-sm mb-3">No review notes yet. Add them during code review.</p>
          <Button variant="outline" size="sm" onClick={() => setIsEditing(true)}>
            Add Notes
          </Button>
        </div>
      )}
    </div>
  );
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

  const handleTaskUpdate = useCallback((updated: Task) => {
    setTask(updated);
  }, []);

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

        <EditablePlanSection
          taskId={task.id}
          initialPlan={task.plan || ""}
          onUpdate={handleTaskUpdate}
        />

        <EditableReviewNotesSection
          taskId={task.id}
          initialNotes={task.review_notes || ""}
          onUpdate={handleTaskUpdate}
        />

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
