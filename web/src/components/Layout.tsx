import { ReactNode } from 'react'
import { NotificationSettings } from './NotificationSettings'
import { DarkModeToggle } from './DarkModeToggle'

interface LayoutProps {
  children: ReactNode
}

export function Layout({ children }: LayoutProps) {

  return (
    <div className="h-screen flex flex-col bg-gray-50 dark:bg-gray-900">
      {/* Header */}
      <header className="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 px-4 sm:px-6 py-3 sm:py-4 flex-shrink-0">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2 sm:gap-4">
            <h1 className="text-lg sm:text-xl font-semibold text-gray-900 dark:text-gray-100">Habibi-Go</h1>
          </div>
          
          <div className="flex items-center space-x-2">
            <DarkModeToggle />
            <NotificationSettings />
          </div>
        </div>
      </header>

      {/* Content */}
      <main className="flex-1 overflow-hidden flex flex-col">
        {children}
      </main>
    </div>
  )
}