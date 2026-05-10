import { cn } from "@/lib/utils";
import { HTMLAttributes, forwardRef } from "react";
import { TaskStatus } from "@/lib/api";

export interface BadgeProps extends HTMLAttributes<HTMLDivElement> {
  variant?: "default" | "success" | "warning" | "error" | "info";
}

const Badge = forwardRef<HTMLDivElement, BadgeProps>(
  ({ className, variant = "default", ...props }, ref) => {
    const variants = {
      default: "bg-gray-100 text-gray-800",
      success: "bg-green-100 text-green-800",
      warning: "bg-orange-100 text-orange-800",
      error: "bg-red-100 text-red-800",
      info: "bg-blue-100 text-blue-800",
    };

    return (
      <div
        ref={ref}
        className={cn(
          "inline-flex items-center px-3 py-1 rounded-full text-xs font-semibold",
          variants[variant],
          className
        )}
        {...props}
      />
    );
  }
);

Badge.displayName = "Badge";

export function getStatusVariant(status: TaskStatus): BadgeProps["variant"] {
  switch (status) {
    case "completed":
    case "approved":
      return "success";
    case "failed":
      return "error";
    case "planning":
    case "in_progress":
      return "info";
    case "reviewing":
      return "warning";
    default:
      return "default";
  }
}

export function formatStatus(status: TaskStatus): string {
  return status
    .split("_")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}

export { Badge };
