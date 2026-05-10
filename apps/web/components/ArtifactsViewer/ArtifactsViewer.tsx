"use client";

import { useState } from "react";
import { Artifact } from "@/lib/api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { formatDate } from "@/lib/utils";
import { cn } from "@/lib/utils";

interface ArtifactsViewerProps {
  artifacts: Artifact[];
}

function getArtifactTypeVariant(artifactType: string): "default" | "success" | "warning" | "error" | "info" {
  const typeLower = artifactType.toLowerCase();
  if (typeLower.includes("diff")) return "info";
  if (typeLower.includes("plan")) return "success";
  if (typeLower.includes("log")) return "warning";
  if (typeLower.includes("error")) return "error";
  return "default";
}

function formatArtifactType(artifactType: string): string {
  return artifactType
    .split("_")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}

interface ArtifactItemProps {
  artifact: Artifact;
}

function ArtifactItem({ artifact }: ArtifactItemProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const hasContent = artifact.content && artifact.content.trim().length > 0;

  return (
    <div className="border border-gray-200 rounded-xl overflow-hidden hover:shadow-sm transition-shadow">
      <button
        type="button"
        onClick={() => hasContent && setIsExpanded(!isExpanded)}
        className={cn(
          "w-full px-5 py-4 flex items-center justify-between text-left",
          hasContent && "cursor-pointer hover:bg-gray-50 transition-colors"
        )}
        disabled={!hasContent}
      >
        <div className="flex items-center gap-3 min-w-0 flex-1">
          <Badge variant={getArtifactTypeVariant(artifact.artifact_type)}>
            {formatArtifactType(artifact.artifact_type)}
          </Badge>
          <div className="min-w-0">
            <p className="text-sm font-medium text-gray-900 truncate">
              {artifact.name}
            </p>
            {artifact.content_type && (
              <p className="text-xs text-gray-400">
                {artifact.content_type}
              </p>
            )}
          </div>
        </div>
        <div className="flex items-center gap-3 flex-shrink-0 ml-4">
          <span className="text-xs text-gray-400 hidden sm:inline">
            {formatDate(artifact.created_at)}
          </span>
          {hasContent && (
            <div className="flex items-center gap-1">
              <span className="text-xs text-blue-600 font-medium hidden sm:inline">
                {isExpanded ? "Collapse" : "View"}
              </span>
              <svg
                className={cn(
                  "w-4 h-4 text-gray-400 transition-transform",
                  isExpanded && "rotate-180"
                )}
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
              </svg>
            </div>
          )}
        </div>
      </button>

      {isExpanded && hasContent && (
        <div className="border-t border-gray-200">
          <div className="p-4 bg-gray-900">
            <pre className="text-sm text-gray-100 font-mono whitespace-pre-wrap overflow-x-auto max-h-96 overflow-y-auto leading-relaxed">
              {artifact.content}
            </pre>
          </div>
          
          {artifact.metadata && Object.keys(artifact.metadata).length > 0 && (
            <div className="px-4 py-3 bg-gray-50 border-t border-gray-200">
              <p className="text-xs font-medium text-gray-500 uppercase tracking-wide mb-2">
                Metadata
              </p>
              <pre className="text-xs text-gray-600 overflow-x-auto">
                {JSON.stringify(artifact.metadata, null, 2)}
              </pre>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export function ArtifactsViewer({ artifacts }: ArtifactsViewerProps) {
  if (artifacts.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="text-lg flex items-center gap-2">
            <svg className="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            Artifacts
          </CardTitle>
        </CardHeader>
        <CardContent className="py-8 text-center">
          <div className="mx-auto w-12 h-12 bg-gray-100 rounded-full flex items-center justify-center mb-3">
            <svg className="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
          </div>
          <p className="text-sm text-gray-500">No artifacts yet</p>
          <p className="text-xs text-gray-400 mt-1">Artifacts like diffs, logs, and plans will appear here</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader className="pb-4">
        <CardTitle className="text-lg flex items-center gap-2">
          <svg className="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          Artifacts
          <span className="ml-2 px-2 py-0.5 bg-gray-100 text-gray-600 text-xs font-medium rounded-full">
            {artifacts.length}
          </span>
        </CardTitle>
      </CardHeader>
      <CardContent className="pt-0 space-y-3">
        {artifacts.map((artifact) => (
          <ArtifactItem key={artifact.id} artifact={artifact} />
        ))}
      </CardContent>
    </Card>
  );
}
