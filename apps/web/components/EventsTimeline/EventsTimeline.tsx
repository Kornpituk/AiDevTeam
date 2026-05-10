import { Event } from "@/lib/api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { formatDate } from "@/lib/utils";
import { cn } from "@/lib/utils";

interface EventsTimelineProps {
  events: Event[];
}

function getEventTypeVariant(eventType: string): "default" | "success" | "warning" | "error" | "info" {
  const typeLower = eventType.toLowerCase();
  if (typeLower.includes("complete") || typeLower.includes("success")) return "success";
  if (typeLower.includes("fail") || typeLower.includes("error")) return "error";
  if (typeLower.includes("warn") || typeLower.includes("review")) return "warning";
  if (typeLower.includes("status") || typeLower.includes("update") || typeLower.includes("progress")) return "info";
  return "default";
}

function formatEventType(eventType: string): string {
  return eventType
    .split("_")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}

export function EventsTimeline({ events }: EventsTimelineProps) {
  if (events.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="text-lg flex items-center gap-2">
            <svg className="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            Events
          </CardTitle>
        </CardHeader>
        <CardContent className="py-8 text-center">
          <div className="mx-auto w-12 h-12 bg-gray-100 rounded-full flex items-center justify-center mb-3">
            <svg className="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z" />
            </svg>
          </div>
          <p className="text-sm text-gray-500">No events yet</p>
          <p className="text-xs text-gray-400 mt-1">Events will appear here as the task progresses</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader className="pb-4">
        <CardTitle className="text-lg flex items-center gap-2">
          <svg className="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          Events
          <span className="ml-2 px-2 py-0.5 bg-gray-100 text-gray-600 text-xs font-medium rounded-full">
            {events.length}
          </span>
        </CardTitle>
      </CardHeader>
      <CardContent className="pt-0">
        <div className="relative">
          <div className="absolute left-4 top-0 bottom-0 w-px bg-gray-200" />
          
          <div className="space-y-0">
            {events.map((event, index) => (
              <div key={event.id} className="relative pl-10 pb-6 last:pb-0">
                <div className={cn(
                  "absolute left-2 w-5 h-5 rounded-full border-4 border-white shadow-sm",
                  index === 0 ? "bg-blue-500" : "bg-gray-300"
                )} />
                
                <div className="py-2">
                  <div className="flex items-center gap-2 mb-2 flex-wrap">
                    <Badge variant={getEventTypeVariant(event.event_type)}>
                      {formatEventType(event.event_type)}
                    </Badge>
                    <span className="text-xs text-gray-400 flex items-center gap-1">
                      <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                      {formatDate(event.created_at)}
                    </span>
                  </div>
                  
                  {event.message && (
                    <p className="text-sm text-gray-700 mb-2">{event.message}</p>
                  )}
                  
                  {event.metadata && Object.keys(event.metadata).length > 0 && (
                    <details className="group">
                      <summary className="text-xs text-gray-500 cursor-pointer hover:text-gray-700 inline-flex items-center gap-1">
                        <svg className="w-3.5 h-3.5 transition-transform group-open:rotate-90" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                        </svg>
                        View metadata
                      </summary>
                      <pre className="mt-2 text-xs text-gray-600 bg-gray-50 p-3 rounded-lg overflow-x-auto">
                        {JSON.stringify(event.metadata, null, 2)}
                      </pre>
                    </details>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
