import { useState } from 'react'
import { useAppStore } from '../store'
import { AgentControl } from './AgentControl'
import { FileDiffs } from './FileDiffs'
import { ManageSession } from './ManageSession'
import { useDiffStats } from '../hooks/useDiffStats'

type TabType = 'assistant' | 'diffs' | 'manage'

export function SessionView() {
  const [activeTab, setActiveTab] = useState<TabType>('assistant')
  const { currentSession } = useAppStore()
  const { data: diffStats } = useDiffStats(currentSession?.id)

  if (!currentSession) {
    return (
      <div className="h-full flex items-center justify-center text-gray-500">
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
          <span className="text-green-600"> +{diffStats.additions}</span>,
          <span className="text-red-600"> -{diffStats.deletions}</span>)
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
    { id: 'manage' as TabType, label: 'Manage Session', icon: '⚙️' },
  ]

  return (
    <div className="h-full flex flex-col">
      {/* Tab Headers */}
      <div className="border-b border-gray-200">
        <nav className="flex -mb-px">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`
                px-6 py-3 text-sm font-medium transition-colors flex items-center
                ${activeTab === tab.id
                  ? 'text-blue-600 border-b-2 border-blue-600'
                  : 'text-gray-500 hover:text-gray-700 hover:border-gray-300 border-b-2 border-transparent'
                }
              `}
            >
              <span className="mr-2">{tab.icon}</span>
              {tab.label}
            </button>
          ))}
        </nav>
      </div>

      {/* Tab Content */}
      <div className="flex-1 overflow-hidden">
        {activeTab === 'assistant' && <AgentControl />}
        {activeTab === 'diffs' && <FileDiffs />}
        {activeTab === 'manage' && <ManageSession />}
      </div>
    </div>
  )
}