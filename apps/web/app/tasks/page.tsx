import Link from "next/link";
import { getTasks, Task } from "@/lib/api";
import { TaskList } from "@/components/TaskList/TaskList";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";

export const revalidate = 0;

export default async function TasksPage() {
  let tasks: Task[] = [];
  let error: string | null = null;
  let isApiDown = false;

  try {
    tasks = await getTasks();
  } catch (e) {
    error = e instanceof Error ? e.message : "Failed to load tasks";
    isApiDown = true;
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Tasks</h1>
          <p className="mt-1 text-sm text-gray-500">
            Manage and track your development tasks
          </p>
        </div>
        <Link href="/tasks/new">
          <Button>
            <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            New Task
          </Button>
        </Link>
      </div>

      {isApiDown && (
        <Card className="border-amber-200 bg-amber-50">
          <CardContent className="py-5">
            <div className="flex items-start">
              <svg className="w-5 h-5 text-amber-500 mr-3 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
              <div>
                <h3 className="text-sm font-medium text-amber-800">
                  Cannot connect to API
                </h3>
                <p className="text-sm text-amber-700 mt-1">
                  Make sure the Go API is running. Start it with:
                </p>
                <code className="mt-2 inline-block bg-amber-100 text-amber-800 text-xs font-mono px-3 py-1.5 rounded">
                  cd services/api && go run cmd/api/main.go
                </code>
                <p className="text-xs text-amber-600 mt-2">
                  Error: {error}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {!isApiDown && <TaskList tasks={tasks} />}
    </div>
  );
}
