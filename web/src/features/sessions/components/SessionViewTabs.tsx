import { TabType } from './SessionViewContainer'

interface DiffStats {
  filesChanged: number
  additions: number
  deletions: number
}

interface SessionViewTabsProps {
  activeTab: TabType
  onTabChange: (tab: TabType) => void
  diffStats: DiffStats | null
}

export function SessionViewTabs({ activeTab, onTabChange, diffStats }: SessionViewTabsProps) {
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
    { id: 'terminal' as TabType, label: 'Terminal', icon: '🖥️' },
    { id: 'manage' as TabType, label: 'Manage Session', icon: '⚙️' },
  ]

  return (
    <div className="border-b border-gray-200 overflow-x-auto">
      <nav className="flex -mb-px min-w-full">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => onTabChange(tab.id)}
            className={`
              px-3 sm:px-6 py-3 text-xs sm:text-sm font-medium transition-colors flex items-center whitespace-nowrap flex-shrink-0
              ${activeTab === tab.id
                ? 'text-blue-600 border-b-2 border-blue-600'
                : 'text-gray-500 hover:text-gray-700 hover:border-gray-300 border-b-2 border-transparent'
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
  )
}