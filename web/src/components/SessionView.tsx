import { useState } from 'react'
import { useAppStore } from '../store'
import { AssistantView } from './AssistantView'
import { FileDiffs } from './FileDiffs'
import { ManageSession } from './ManageSession'
import { Terminal } from './Terminal'
import { useDiffStats } from '../hooks/useDiffStats'
import { useSessionActivity } from '../hooks/useSessionActivity'

type TabType = 'assistant' | 'diffs' | 'terminal' | 'manage'

export function SessionView() {
  const [activeTab, setActiveTab] = useState<TabType>('assistant')
  const { currentSession } = useAppStore()
  const { data: diffStats } = useDiffStats(currentSession?.id)

  // Enable real-time session updates
  useSessionActivity()

  if (!currentSession) {
    return (
      <div className="h-full w-full flex items-center justify-center text-gray-500 dark:text-gray-400">
        <div className="text-center">
          <p className="text-lg mb-2">No session selected</p>
          <p className="text-sm">Select or create a session to get started</p>
        </div>
      </div>
    )
  }

  const getDiffLabel = () => {
    if (!diffStats || diffStats.filesChanged === 0) {
      return 'Branch Changes'
    }
    return (
      <span>
        Branch Changes
        <span className="ml-2 text-xs">
          ({diffStats.filesChanged} {diffStats.filesChanged === 1 ? 'file' : 'files'},
          <span className="text-green-600 dark:text-green-400"> +{diffStats.additions}</span>,
          <span className="text-red-600 dark:text-red-400"> -{diffStats.deletions}</span>)
        </span>
      </span>
    )
  }

  const tabs = [
    { id: 'assistant' as TabType, label: 'Assistant', icon: '💬' },
    {
      id: 'diffs' as TabType,
      label: getDiffLabel(),
      icon: '📝',
      hasChanges: diffStats && diffStats.filesChanged > 0
    },
    { id: 'terminal' as TabType, label: 'Terminal', icon: '🖥️' },
    { id: 'manage' as TabType, label: 'Manage Session', icon: '⚙️' },
  ]

  return (
    <div className="h-full w-full flex flex-col">
      {/* Tab Headers */}
      <div className="border-b border-gray-200 dark:border-gray-700 overflow-x-auto">
        <nav className="flex -mb-px min-w-full">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`
                px-3 sm:px-6 py-3 text-xs sm:text-sm font-medium transition-colors flex items-center whitespace-nowrap flex-shrink-0
                ${activeTab === tab.id
                  ? 'text-blue-600 dark:text-blue-400 border-b-2 border-blue-600 dark:border-blue-400'
                  : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 hover:border-gray-300 dark:hover:border-gray-600 border-b-2 border-transparent'
                }
              `}
            >
              <span className="mr-1 sm:mr-2">{tab.icon}</span>
              <span className="hidden sm:inline">{tab.label}</span>
              <span className="sm:hidden">
                {tab.id === 'assistant' && 'Chat'}
                {tab.id === 'diffs' && 'Diffs'}
                {tab.id === 'terminal' && 'Term'}
                {tab.id === 'manage' && 'Manage'}
              </span>
            </button>
          ))}
        </nav>
      </div>

      {/* Tab Content */}
      <div className="flex-1 overflow-hidden">
        {activeTab === 'assistant' && <AssistantView />}
        {activeTab === 'diffs' && <FileDiffs />}
        {activeTab === 'terminal' && <Terminal />}
        {activeTab === 'manage' && <ManageSession />}
      </div>
    </div>
  )
}
