import Link from "next/link";
import { Task } from "@/lib/api";
import { Card, CardContent } from "@/components/ui/card";
import { Badge, getStatusVariant, formatStatus } from "@/components/ui/badge";
import { formatDate } from "@/lib/utils";

interface TaskListProps {
  tasks: Task[];
}

export function TaskList({ tasks }: TaskListProps) {
  if (tasks.length === 0) {
    return (
      <Card>
        <CardContent className="py-16 text-center">
          <div className="mx-auto w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mb-4">
            <svg className="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
            </svg>
          </div>
          <h3 className="text-lg font-semibold text-gray-900 mb-2">No tasks yet</h3>
          <p className="text-gray-500 mb-4">Get started by creating your first development task.</p>
          <Link
            href="/tasks/new"
            className="inline-flex items-center px-4 py-2 text-sm font-medium rounded-lg text-white bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-700 hover:to-indigo-700 transition-all"
          >
            <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            Create Task
          </Link>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      {tasks.map((task) => (
        <Link key={task.id} href={`/tasks/${task.id}`} className="block">
          <Card className="hover:shadow-lg transition-all duration-200 group cursor-pointer">
            <CardContent className="py-5">
              <div className="flex items-start justify-between gap-4">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-3 mb-2">
                    <h3 className="text-base font-semibold text-gray-900 truncate group-hover:text-blue-600 transition-colors">
                      {task.title}
                    </h3>
                  </div>
                  {task.description && (
                    <p className="text-sm text-gray-500 line-clamp-2 mb-3">
                      {task.description}
                    </p>
                  )}
                  <div className="flex items-center gap-4 text-xs text-gray-400">
                    <span className="flex items-center gap-1">
                      <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                      {formatDate(task.created_at)}
                    </span>
                    {task.updated_at !== task.created_at && (
                      <span className="flex items-center gap-1">
                        <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                        </svg>
                        Updated {formatDate(task.updated_at)}
                      </span>
                    )}
                  </div>
                </div>
                <div className="flex-shrink-0">
                  <Badge variant={getStatusVariant(task.status)}>
                    {formatStatus(task.status)}
                  </Badge>
                </div>
              </div>
            </CardContent>
          </Card>
        </Link>
      ))}
    </div>
  );
}
