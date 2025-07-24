import { useState, useEffect } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { useAppStore } from '../../../store'
import { AssistantView } from '../../assistant/components/AssistantView'
import { FileDiffs } from '../../git/components/FileDiffs'
import { ManageSession } from './ManageSession'
import { Terminal } from '../../terminal/components/Terminal'
import { useDiffStats } from '../../git/hooks/useDiffStats'
import { useSessionActivity } from '../../../hooks/useSessionActivity'

type TabType = 'assistant' | 'diffs' | 'terminal' | 'manage'

function ProjectReadme({ projectPath }: { projectPath: string }) {
  const [readmeContent, setReadmeContent] = useState<string | null>(null)
  const [readmeFilename, setReadmeFilename] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchReadme = async () => {
      try {
        setLoading(true)
        // Try to read README.md first, then README.txt, then README
        const readmeFiles = ['README.md', 'README.txt', 'README', 'readme.md', 'readme.txt', 'readme']

        for (const filename of readmeFiles) {
          try {
            const response = await fetch(`/api/projects/file?path=${encodeURIComponent(projectPath)}&file=${encodeURIComponent(filename)}`)
            if (response.ok) {
              const content = await response.text()
              setReadmeContent(content)
              setReadmeFilename(filename)
              return
            }
          } catch (error) {
            // Continue to next file
          }
        }

        // No README found
        setReadmeContent(null)
      } catch (error) {
        console.error('Error fetching README:', error)
        setReadmeContent(null)
      } finally {
        setLoading(false)
      }
    }

    fetchReadme()
  }, [projectPath])

  if (loading) {
    return (
      <div className="h-full w-full flex items-center justify-center text-gray-500 dark:text-gray-400">
        <div className="text-center">
          <p>Loading project README...</p>
        </div>
      </div>
    )
  }

  if (!readmeContent) {
    return (
      <div className="h-full w-full flex items-center justify-center text-gray-500 dark:text-gray-400">
        <div className="text-center">
          <p className="text-lg mb-2">No session selected</p>
          <p className="text-sm mb-4">Select or create a session to get started</p>
          <p className="text-xs">No README found in project directory</p>
        </div>
      </div>
    )
  }

  const isMarkdown = readmeFilename?.toLowerCase().endsWith('.md') || false

  return (
    <div className="h-full w-full overflow-auto p-6">
      <div className="max-w-4xl mx-auto">
        <div className="mb-4 text-sm text-gray-500 dark:text-gray-400 border-b border-gray-200 dark:border-gray-700 pb-2">
          📄 {readmeFilename}
        </div>
        {isMarkdown ? (
          <div className="prose prose-sm dark:prose-invert max-w-none">
            <ReactMarkdown remarkPlugins={[remarkGfm]}>
              {readmeContent}
            </ReactMarkdown>
          </div>
        ) : (
          <pre className="whitespace-pre-wrap text-sm text-gray-900 dark:text-gray-100 font-mono bg-gray-50 dark:bg-gray-800 p-4 rounded-lg">
            {readmeContent}
          </pre>
        )}
      </div>
    </div>
  )
}

export function SessionView() {
  const [activeTab, setActiveTab] = useState<TabType>('assistant')
  const { currentSession, currentProject } = useAppStore()
  const { data: diffStats } = useDiffStats(currentSession?.id)

  // Enable real-time session updates
  useSessionActivity()

  if (!currentSession) {
    if (currentProject) {
      return <ProjectReadme projectPath={currentProject.path} />
    }

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
        <nav className="flex min-w-full">
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
