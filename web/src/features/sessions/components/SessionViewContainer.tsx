import { useState } from 'react'
import { useAppSelector } from '../../../app/hooks'
import { selectCurrentSession } from '../slice/sessionsSlice'
import { useGetSessionDiffsQuery } from '../api/sessionsApi'
import { useSessionActivity } from '../hooks/useSessionActivity'
import { SessionViewTabs } from './SessionViewTabs'
import { AgentChatContainer } from '../../agents/components/AgentChatContainer'
import { SessionDiffsContainer } from './SessionDiffsContainer'
import { SessionTerminalContainer } from './SessionTerminalContainer'
import { SessionManageContainer } from './SessionManageContainer'

export type TabType = 'assistant' | 'diffs' | 'terminal' | 'manage'

export function SessionViewContainer() {
  const [activeTab, setActiveTab] = useState<TabType>('assistant')
  const currentSession = useAppSelector(selectCurrentSession)
  
  // Enable real-time session updates
  useSessionActivity()
  
  const { data: diffs } = useGetSessionDiffsQuery(currentSession?.id ?? 0, {
    skip: !currentSession,
  })

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

  const diffStats = diffs ? {
    filesChanged: diffs.files?.length || 0,
    additions: diffs.files?.reduce((sum: number, file: any) => sum + (file.additions || 0), 0) || 0,
    deletions: diffs.files?.reduce((sum: number, file: any) => sum + (file.deletions || 0), 0) || 0,
  } : null

  return (
    <div className="h-full flex flex-col">
      <SessionViewTabs
        activeTab={activeTab}
        onTabChange={setActiveTab}
        diffStats={diffStats}
      />
      
      <div className="flex-1 overflow-hidden">
        {activeTab === 'assistant' && <AgentChatContainer sessionId={currentSession.id} />}
        {activeTab === 'diffs' && <SessionDiffsContainer sessionId={currentSession.id} />}
        {activeTab === 'terminal' && <SessionTerminalContainer sessionId={currentSession.id} />}
        {activeTab === 'manage' && <SessionManageContainer session={currentSession} />}
      </div>
    </div>
  )
}