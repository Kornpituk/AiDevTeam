"use client";

import { cn } from "@/lib/utils";
import { InputHTMLAttributes, forwardRef } from "react";

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {}

const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, ...props }, ref) => (
    <input
      ref={ref}
      className={cn(
        "w-full px-4 py-2.5 border border-gray-300 rounded-lg text-gray-900 placeholder-gray-400 text-sm",
        "focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent",
        "disabled:bg-gray-50 disabled:text-gray-500 disabled:cursor-not-allowed",
        "transition-all duration-200",
        className
      )}
      {...props}
    />
  )
);

Input.displayName = "Input";

export { Input };
