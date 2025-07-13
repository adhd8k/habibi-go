import { Agent } from '../types'
import { formatDistanceToNow } from 'date-fns'

interface AgentSelectorProps {
  agents: Agent[]
  currentAgent?: Agent
  onSelectAgent: (agent: Agent) => void
  onCreateNewAgent: () => void
}

export function AgentSelector({ agents, currentAgent, onSelectAgent, onCreateNewAgent }: AgentSelectorProps) {
  // Sort agents by started_at desc (newest first)
  const sortedAgents = [...agents].sort((a, b) => 
    new Date(b.started_at).getTime() - new Date(a.started_at).getTime()
  )

  return (
    <div className="p-4 border-b">
      <h3 className="text-sm font-semibold mb-3">Claude Sessions</h3>
      
      <div className="space-y-2 max-h-48 overflow-y-auto">
        {sortedAgents.map((agent) => {
          const isActive = currentAgent?.id === agent.id
          const hasSessionId = !!agent.claude_session_id
          
          return (
            <button
              key={agent.id}
              onClick={() => onSelectAgent(agent)}
              className={`w-full text-left p-2 rounded-md transition-colors ${
                isActive 
                  ? 'bg-blue-100 border border-blue-300' 
                  : 'bg-gray-50 hover:bg-gray-100 border border-gray-200'
              }`}
            >
              <div className="flex items-center justify-between">
                <div className="flex-1">
                  <div className="flex items-center gap-2">
                    <span className={`w-2 h-2 rounded-full ${
                      agent.status === 'running' ? 'bg-green-500' :
                      agent.status === 'stopped' ? 'bg-gray-400' :
                      'bg-red-500'
                    }`} />
                    <span className="text-sm font-medium">
                      Agent #{agent.id}
                    </span>
                    {hasSessionId && (
                      <span className="text-xs px-2 py-0.5 bg-purple-100 text-purple-700 rounded">
                        Resumable
                      </span>
                    )}
                  </div>
                  <div className="text-xs text-gray-500 mt-1">
                    Started {formatDistanceToNow(new Date(agent.started_at), { addSuffix: true })}
                  </div>
                </div>
                <div className="text-xs">
                  {agent.status}
                </div>
              </div>
            </button>
          )
        })}
      </div>
      
      <button
        onClick={onCreateNewAgent}
        className="w-full mt-3 p-2 bg-green-600 text-white rounded-md hover:bg-green-700 text-sm font-medium"
      >
        Start New Claude Session
      </button>
    </div>
  )
}