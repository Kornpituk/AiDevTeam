"use client";

import { cn } from "@/lib/utils";
import { SelectHTMLAttributes, forwardRef } from "react";

export interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {}

const Select = forwardRef<HTMLSelectElement, SelectProps>(
  ({ className, ...props }, ref) => (
    <select
      ref={ref}
      className={cn(
        "w-full px-4 py-2.5 border border-gray-300 rounded-lg text-gray-900 bg-white text-sm cursor-pointer",
        "focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent",
        "disabled:bg-gray-50 disabled:text-gray-500 disabled:cursor-not-allowed",
        "transition-all duration-200",
        className
      )}
      {...props}
    />
  )
);

Select.displayName = "Select";

export { Select };
