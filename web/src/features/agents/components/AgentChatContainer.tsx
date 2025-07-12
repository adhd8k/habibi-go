import { useState, useEffect } from 'react'
import { useAppSelector, useAppDispatch } from '../../../app/hooks'
import { useGetAgentsQuery, useCreateAgentMutation } from '../api/agentsApi'
import { selectCurrentAgent, setCurrentAgent } from '../slice/agentsSlice'
import { AgentChatView } from './AgentChatView'
import { LoadingSpinner } from '../../../shared/components/LoadingSpinner'

interface AgentChatContainerProps {
  sessionId: number
}

export function AgentChatContainer({ sessionId }: AgentChatContainerProps) {
  const dispatch = useAppDispatch()
  const currentAgent = useAppSelector(selectCurrentAgent)
  const [isCreatingAgent, setIsCreatingAgent] = useState(false)
  
  const { data: agents, isLoading } = useGetAgentsQuery(sessionId)
  const [createAgent] = useCreateAgentMutation()

  // Set current agent when agents load
  useEffect(() => {
    if (agents && agents.length > 0 && !currentAgent) {
      const activeAgent = agents.find(a => a.status === 'running') || agents[0]
      dispatch(setCurrentAgent(activeAgent))
    }
  }, [agents, currentAgent, dispatch])

  const handleCreateAgent = async () => {
    setIsCreatingAgent(true)
    try {
      const newAgent = await createAgent({
        session_id: sessionId,
        agent_type: 'claude',
        command: 'claude --output-format stream-json',
      }).unwrap()
      dispatch(setCurrentAgent(newAgent))
    } catch (error) {
      console.error('Failed to create agent:', error)
    } finally {
      setIsCreatingAgent(false)
    }
  }

  if (isLoading) {
    return <LoadingSpinner className="mt-8" />
  }

  if (!agents || agents.length === 0) {
    return (
      <div className="h-full flex items-center justify-center">
        <div className="text-center">
          <p className="text-gray-600 mb-4">No assistant available for this session</p>
          <button
            onClick={handleCreateAgent}
            disabled={isCreatingAgent}
            className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
          >
            {isCreatingAgent ? 'Creating...' : 'Start Claude Assistant'}
          </button>
        </div>
      </div>
    )
  }

  if (!currentAgent) {
    return <LoadingSpinner className="mt-8" />
  }

  return <AgentChatView agent={currentAgent} />
}