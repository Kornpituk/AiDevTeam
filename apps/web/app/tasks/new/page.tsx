import Link from "next/link";
import { TaskForm } from "@/components/TaskForm/TaskForm";
import { Button } from "@/components/ui/button";

export default function NewTaskPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link href="/tasks">
          <Button variant="ghost" size="sm">
            <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
            Back to Tasks
          </Button>
        </Link>
      </div>

      <div>
        <h1 className="text-2xl font-bold text-gray-900">Create New Task</h1>
        <p className="mt-1 text-sm text-gray-500">
          Fill in the details to create a new development task
        </p>
      </div>

      <TaskForm />
    </div>
  );
}
