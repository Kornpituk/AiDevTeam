import type { Metadata } from "next";
import Link from "next/link";
import "./globals.css";

export const metadata: Metadata = {
  title: "AiDevT - AI Dev Task Dashboard",
  description: "Control plane for AI-assisted software development",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className="min-h-screen bg-gray-50">
        <div className="min-h-screen flex flex-col">
          <nav className="bg-white border-b border-gray-200 sticky top-0 z-50">
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
              <div className="flex justify-between h-16">
                <div className="flex items-center space-x-8">
                  <Link href="/tasks" className="flex items-center space-x-2">
                    <div className="w-8 h-8 bg-gradient-to-br from-blue-500 to-indigo-600 rounded-lg flex items-center justify-center">
                      <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
                      </svg>
                    </div>
                    <span className="text-xl font-bold text-gray-900">AiDevT</span>
                  </Link>
                   <div className="hidden md:flex items-center space-x-1">
                     <Link
                       href="/tasks"
                       className="px-3 py-2 rounded-md text-sm font-medium text-gray-700 hover:text-gray-900 hover:bg-gray-100 transition-colors"
                     >
                       Tasks
                     </Link>
                     <Link
                       href="/tasks/new"
                       className="px-3 py-2 rounded-md text-sm font-medium text-gray-700 hover:text-gray-900 hover:bg-gray-100 transition-colors"
                     >
                       New Task
                     </Link>
                     <Link
                       href="/agents"
                       className="px-3 py-2 rounded-md text-sm font-medium text-gray-700 hover:text-gray-900 hover:bg-gray-100 transition-colors"
                     >
                       Agents
                     </Link>
                   </div>
                </div>
                <div className="flex items-center">
                  <Link
                    href="/tasks/new"
                    className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-lg text-white bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-700 hover:to-indigo-700 transition-all shadow-sm hover:shadow"
                  >
                    <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                    </svg>
                    New Task
                  </Link>
                </div>
              </div>
            </div>
          </nav>
          <main className="flex-1">
            <div className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8 px-4">
              {children}
            </div>
          </main>
          <footer className="bg-white border-t border-gray-200 mt-auto">
            <div className="max-w-7xl mx-auto px-4 py-4 sm:px-6 lg:px-8">
              <p className="text-sm text-gray-500 text-center">
                AiDevT - AI Dev Task Dashboard
              </p>
            </div>
          </footer>
        </div>
      </body>
    </html>
  );
}
