"use client";

import Link from "next/link";
import { AgentProfile } from "@/lib/api";
import { Card, CardContent } from "@/components/ui/card";
import { formatDate } from "@/lib/utils";

interface AgentProfileListProps {
  profiles: AgentProfile[];
}

export function AgentProfileList({ profiles }: AgentProfileListProps) {
  if (profiles.length === 0) {
    return (
      <Card>
        <CardContent className="py-16 text-center">
          <div className="mx-auto w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mb-4">
            <svg className="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
            </svg>
          </div>
          <h3 className="text-lg font-semibold text-gray-900 mb-2">No agent profiles yet</h3>
          <p className="text-gray-500 mb-4">Create your first agent profile to define reusable agent roles.</p>
          <Link
            href="/agents/profiles/new"
            className="inline-flex items-center px-4 py-2 text-sm font-medium rounded-lg text-white bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-700 hover:to-indigo-700 transition-all"
          >
            <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            Create Profile
          </Link>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      {profiles.map((profile) => (
        <Card key={profile.id} className="hover:shadow-lg transition-all duration-200">
          <CardContent className="py-5">
            <div className="flex items-start justify-between gap-4">
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-3 mb-2">
                  <div className="w-10 h-10 bg-gradient-to-br from-purple-500 to-pink-500 rounded-lg flex items-center justify-center flex-shrink-0">
                    <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                    </svg>
                  </div>
                  <div>
                    <h3 className="text-base font-semibold text-gray-900">
                      {profile.name}
                    </h3>
                    <p className="text-sm text-purple-600 font-medium">
                      {profile.role}
                    </p>
                  </div>
                </div>
                {profile.description && (
                  <p className="text-sm text-gray-500 line-clamp-2 mb-3 ml-13">
                    {profile.description}
                  </p>
                )}
                {profile.default_model && (
                  <p className="text-xs text-gray-400 ml-13">
                    Default model: {profile.default_model}
                  </p>
                )}
              </div>
              <div className="flex-shrink-0 text-xs text-gray-400">
                {formatDate(profile.created_at)}
              </div>
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}
