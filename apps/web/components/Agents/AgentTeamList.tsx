"use client";

import Link from "next/link";
import { AgentTeam } from "@/lib/api";
import { Card, CardContent } from "@/components/ui/card";
import { formatDate } from "@/lib/utils";

interface AgentTeamListProps {
  teams: AgentTeam[];
}

export function AgentTeamList({ teams }: AgentTeamListProps) {
  if (teams.length === 0) {
    return (
      <Card>
        <CardContent className="py-16 text-center">
          <div className="mx-auto w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mb-4">
            <svg className="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
            </svg>
          </div>
          <h3 className="text-lg font-semibold text-gray-900 mb-2">No agent teams yet</h3>
          <p className="text-gray-500 mb-4">Create a team to group agent profiles for specific workflows.</p>
          <Link
            href="/agents/teams/new"
            className="inline-flex items-center px-4 py-2 text-sm font-medium rounded-lg text-white bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-700 hover:to-indigo-700 transition-all"
          >
            <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            Create Team
          </Link>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      {teams.map((team) => (
        <Link key={team.id} href={`/agents/teams/${team.id}`} className="block">
          <Card className="hover:shadow-lg transition-all duration-200 group cursor-pointer">
            <CardContent className="py-5">
              <div className="flex items-start justify-between gap-4">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-3 mb-2">
                    <div className="w-10 h-10 bg-gradient-to-br from-green-500 to-teal-500 rounded-lg flex items-center justify-center flex-shrink-0">
                      <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                      </svg>
                    </div>
                    <div>
                      <h3 className="text-base font-semibold text-gray-900 group-hover:text-blue-600 transition-colors">
                        {team.name}
                      </h3>
                    </div>
                  </div>
                  {team.description && (
                    <p className="text-sm text-gray-500 line-clamp-2 ml-13">
                      {team.description}
                    </p>
                  )}
                </div>
                <div className="flex-shrink-0 text-xs text-gray-400">
                  {formatDate(team.created_at)}
                </div>
              </div>
            </CardContent>
          </Card>
        </Link>
      ))}
    </div>
  );
}
