import { useState, useEffect, useCallback } from 'react'
import { useAppSelector, useAppDispatch } from '../../../app/hooks'
import { useGetAgentsQuery, useCreateAgentMutation, useGetAgentMessagesQuery, useSendAgentMessageMutation } from '../api/agentsApi'
import { selectCurrentAgent, setCurrentAgent } from '../slice/agentsSlice'
import { AgentChatView } from './AgentChatView'
import { LoadingSpinner } from '../../../shared/components/LoadingSpinner'

interface AgentChatContainerProps {
  sessionId: number
}

export function AgentChatContainer({ sessionId }: AgentChatContainerProps) {
  const dispatch = useAppDispatch()
  const currentAgent = useAppSelector(selectCurrentAgent)
  const [pendingMessage, setPendingMessage] = useState<string | null>(null)
  
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

  // Create agent and send message
  const handleSendMessage = useCallback(async (message: string) => {
    if (!agents || agents.length === 0) {
      // No agent exists, create one first
      console.log('Creating new Claude agent with command: claude --output-format stream-json')
      try {
        const newAgent = await createAgent({
          session_id: sessionId,
          agent_type: 'claude',
          command: 'claude --output-format stream-json',
        }).unwrap()
        dispatch(setCurrentAgent(newAgent))
        
        // Send the message to the new agent
        console.log(`Sending message to agent ${newAgent.id}:`, message)
        await sendMessage({ agentId: newAgent.id, message }).unwrap()
      } catch (error) {
        console.error('Failed to create agent or send message:', error)
      }
    } else if (currentAgent) {
      // Agent exists, just send the message
      console.log(`Sending message to agent ${currentAgent.id}:`, message)
      try {
        await sendMessage({ agentId: currentAgent.id, message }).unwrap()
      } catch (error) {
        console.error('Failed to send message:', error)
      }
    }
  }, [agents, currentAgent, sessionId, createAgent, sendMessage, dispatch])

  // Handle pending message after agent is created
  useEffect(() => {
    if (pendingMessage && currentAgent) {
      handleSendMessage(pendingMessage)
      setPendingMessage(null)
    }
  }, [pendingMessage, currentAgent, handleSendMessage])

  if (isLoading) {
    return <LoadingSpinner className="mt-8" />
  }

  // Always show the chat interface, even without an agent
  // The agent will be created on first message
  return (
    <AgentChatView 
      agent={currentAgent || null}
      messages={messages || []}
      onSendMessage={handleSendMessage}
    />
  )
}