"use client";

import { cn } from "@/lib/utils";
import { TextareaHTMLAttributes, forwardRef } from "react";

export interface TextareaProps extends TextareaHTMLAttributes<HTMLTextAreaElement> {}

const Textarea = forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ className, ...props }, ref) => (
    <textarea
      ref={ref}
      className={cn(
        "w-full px-4 py-2.5 border border-gray-300 rounded-lg text-gray-900 placeholder-gray-400 text-sm resize-vertical min-h-[120px]",
        "focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent",
        "disabled:bg-gray-50 disabled:text-gray-500 disabled:cursor-not-allowed",
        "transition-all duration-200",
        className
      )}
      {...props}
    />
  )
);

Textarea.displayName = "Textarea";

export { Textarea };
