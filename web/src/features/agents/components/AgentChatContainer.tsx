import { useState, useEffect, useCallback } from 'react'
import { useAppSelector, useAppDispatch } from '../../../app/hooks'
import { useGetAgentsQuery, useCreateAgentMutation, useGetAgentMessagesQuery, useSendAgentMessageMutation } from '../api/agentsApi'
import { selectCurrentAgent, setCurrentAgent } from '../slice/agentsSlice'
import { AgentChatView } from './AgentChatView'
import { LoadingSpinner } from '../../../shared/components/LoadingSpinner'
import { getErrorMessage } from '../../../utils/errorHandling'

interface AgentChatContainerProps {
  sessionId: number
}

export function AgentChatContainer({ sessionId }: AgentChatContainerProps) {
  const dispatch = useAppDispatch()
  const currentAgent = useAppSelector(selectCurrentAgent)
  const [isCreatingAgent, setIsCreatingAgent] = useState(false)
  
  const { data: agents, isLoading } = useGetAgentsQuery(sessionId)
  const [createAgent] = useCreateAgentMutation()
  const [sendMessage] = useSendAgentMessageMutation()
  
  // Load messages for current agent
  const { data: messages } = useGetAgentMessagesQuery(currentAgent?.id ?? 0, {
    skip: !currentAgent,
  })

  // Set current agent when agents load
  useEffect(() => {
    if (agents && agents.length > 0 && !currentAgent) {
      const activeAgent = agents.find(a => a.status === 'running') || agents[0]
      dispatch(setCurrentAgent(activeAgent))
    }
  }, [agents, currentAgent, dispatch])

  // Create agent handler
  const handleCreateAgent = useCallback(async () => {
    setIsCreatingAgent(true)
    try {
      const newAgent = await createAgent({
        session_id: sessionId,
        agent_type: 'claude-code',
        command: 'claude --output-format stream-json',
      }).unwrap()
      dispatch(setCurrentAgent(newAgent))
    } catch (error) {
      console.error('Failed to create agent:', error)
      alert(`Failed to create agent: ${getErrorMessage(error)}`)
    } finally {
      setIsCreatingAgent(false)
    }
  }, [sessionId, createAgent, dispatch])

  // Send message handler
  const handleSendMessage = useCallback(async (message: string) => {
    if (!currentAgent) {
      console.error('No agent available to send message')
      return
    }

    console.log(`Sending message to agent ${currentAgent.id}:`, message)
    try {
      await sendMessage({ agentId: currentAgent.id, message }).unwrap()
    } catch (error) {
      console.error('Failed to send message:', error)
      alert(`Failed to send message: ${getErrorMessage(error)}`)
    }
  }, [currentAgent, sendMessage])

  if (isLoading) {
    return <LoadingSpinner className="mt-8" />
  }

  // No agent exists, show create button
  if (!agents || agents.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-full p-8">
        <div className="text-center mb-8">
          <h3 className="text-xl font-semibold mb-2">No Claude Agent</h3>
          <p className="text-gray-600">Create a Claude agent to start chatting</p>
        </div>
        <button
          onClick={handleCreateAgent}
          disabled={isCreatingAgent}
          className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isCreatingAgent ? 'Creating Agent...' : 'Create Claude Agent'}
        </button>
      </div>
    )
  }

  // Show chat interface with existing agent
  return (
    <AgentChatView 
      agent={currentAgent}
      messages={messages || []}
      onSendMessage={handleSendMessage}
    />
  )
}