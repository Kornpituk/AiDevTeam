import Link from "next/link";
import { notFound } from "next/navigation";
import { getTask, getEvents, getArtifacts, Task, Event, Artifact } from "@/lib/api";
import { TaskDetail } from "@/components/TaskDetail/TaskDetail";
import { EventsTimeline } from "@/components/EventsTimeline/EventsTimeline";
import { ArtifactsViewer } from "@/components/ArtifactsViewer/ArtifactsViewer";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";

export const revalidate = 0;

interface TaskDetailPageParams {
  params: {
    id: string;
  };
}

export default async function TaskDetailPage({ params }: TaskDetailPageParams) {
  const taskId = params.id;

  let task: Task | null = null;
  let events: Event[] = [];
  let artifacts: Artifact[] = [];
  let error: string | null = null;
  let isApiDown = false;
  let isNotFound = false;

  try {
    [task, events, artifacts] = await Promise.all([
      getTask(taskId),
      getEvents(taskId),
      getArtifacts(taskId),
    ]);
  } catch (e) {
    error = e instanceof Error ? e.message : "Failed to load task";
    isApiDown = true;
    
    if (error.toLowerCase().includes("not found") || error.includes("404")) {
      isNotFound = true;
    }
  }

  if (isNotFound) {
    notFound();
  }

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
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {!isApiDown && task && (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2 space-y-6">
            <TaskDetail task={task} />
            <EventsTimeline events={events} />
          </div>
          <div className="space-y-6">
            <ArtifactsViewer artifacts={artifacts} />
          </div>
        </div>
      )}
    </div>
  );
}
