import { useState, useEffect, useRef } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { agentsApi } from '../api/client'
import { wsClient } from '../api/websocket'
import { Agent } from '../types'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'

interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
  timestamp: Date
}

interface ClaudeChatProps {
  agent: Agent
}

export function ClaudeChat({ agent }: ClaudeChatProps) {
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [isProcessing, setIsProcessing] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages])
  
  // Load chat history
  const { data: historyData } = useQuery({
    queryKey: ['chat-history', agent.id],
    queryFn: async () => {
      const response = await agentsApi.chatHistory(agent.id, 100)
      return response.data
    },
    enabled: agent.status === 'running'
  })
  
  // Initialize messages from history
  useEffect(() => {
    if (historyData?.messages) {
      const historicalMessages: Message[] = historyData.messages.map(msg => ({
        id: msg.id.toString(),
        role: msg.role as 'user' | 'assistant',
        content: msg.content,
        timestamp: new Date(msg.created_at)
      }))
      setMessages(historicalMessages)
    }
  }, [historyData])

  useEffect(() => {
    // Subscribe to agent output
    wsClient.subscribe(agent.id)
    
    wsClient.on('agent_output', (message) => {
      if (message.agent_id === agent.id && message.data) {
        const output = message.data.output
        
        // Skip empty lines
        if (!output || output.trim() === '') {
          return
        }
        
        setMessages(prev => {
          const lastMessage = prev[prev.length - 1]
          
          // If there's already an assistant message being built, append to it
          if (lastMessage && lastMessage.role === 'assistant') {
            return [
              ...prev.slice(0, -1),
              {
                ...lastMessage,
                content: lastMessage.content + '\n' + output
              }
            ]
          }
          
          // Otherwise start a new assistant message
          return [
            ...prev,
            {
              id: Date.now().toString(),
              role: 'assistant',
              content: output,
              timestamp: new Date()
            }
          ]
        })
      }
    })
    
    // Listen for response completion
    wsClient.on('agent_response_complete', (message) => {
      if (message.agent_id === agent.id) {
        setIsProcessing(false)
      }
    })

    return () => {
      wsClient.unsubscribe(agent.id)
      wsClient.off('agent_output')
      wsClient.off('agent_response_complete')
    }
  }, [agent.id])

  const sendMessage = useMutation({
    mutationFn: async (message: string) => {
      const response = await agentsApi.execute({ 
        agent_id: agent.id, 
        command: message 
      })
      return response.data
    },
    onMutate: (message) => {
      // Add user message immediately
      setMessages(prev => [
        ...prev,
        {
          id: Date.now().toString(),
          role: 'user',
          content: message,
          timestamp: new Date()
        }
      ])
      setIsProcessing(true)
    },
    onError: () => {
      setIsProcessing(false)
    }
  })

  const handleSend = () => {
    if (!input.trim() || isProcessing) return
    
    sendMessage.mutate(input.trim())
    setInput('')
  }

  return (
    <div className="flex flex-col h-full">
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {messages.length === 0 && (
          <div className="text-center text-gray-500 mt-8">
            <p className="text-lg mb-2">Start a conversation with Claude</p>
            <p className="text-sm">Type a message below to begin</p>
          </div>
        )}
        
        {messages.map((message) => (
          <div
            key={message.id}
            className={`flex ${message.role === 'user' ? 'justify-end' : 'justify-start'}`}
          >
            <div
              className={`max-w-[70%] rounded-lg p-3 ${
                message.role === 'user'
                  ? 'bg-blue-500 text-white'
                  : 'bg-gray-100 text-gray-900'
              }`}
            >
              {message.role === 'user' ? (
                <div className="whitespace-pre-wrap break-words">{message.content}</div>
              ) : (
                <div className="prose prose-sm max-w-none dark:prose-invert">
                  <ReactMarkdown 
                    remarkPlugins={[remarkGfm]}
                    components={{
                      pre: ({ ...props }) => (
                        <pre className="bg-gray-800 text-gray-100 p-2 rounded overflow-x-auto" {...props} />
                      ),
                      code: ({ className, children, ...props }) => {
                        const match = /language-(\w+)/.exec(className || '')
                        return match ? (
                          <code className="bg-gray-800 text-gray-100" {...props}>
                            {children}
                          </code>
                        ) : (
                          <code className="bg-gray-200 px-1 rounded" {...props}>
                            {children}
                          </code>
                        )
                      },
                    }}
                  >
                    {message.content}
                  </ReactMarkdown>
                </div>
              )}
              <div className={`text-xs mt-1 ${
                message.role === 'user' ? 'text-blue-100' : 'text-gray-500'
              }`}>
                {message.timestamp.toLocaleTimeString()}
              </div>
            </div>
          </div>
        ))}
        
        {isProcessing && (
          <div className="flex justify-start">
            <div className="bg-gray-100 text-gray-900 rounded-lg p-3">
              <div className="flex space-x-1">
                <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce"></div>
                <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0.1s' }}></div>
                <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0.2s' }}></div>
              </div>
            </div>
          </div>
        )}
        
        <div ref={messagesEndRef} />
      </div>
      
      <div className="border-t p-4">
        <div className="flex gap-2">
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyPress={(e) => e.key === 'Enter' && !e.shiftKey && handleSend()}
            placeholder="Type your message..."
            disabled={isProcessing || agent.status !== 'running'}
            className="flex-1 p-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
          />
          <button
            onClick={handleSend}
            disabled={!input.trim() || isProcessing || agent.status !== 'running'}
            className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Send
          </button>
        </div>
        
        {agent.status !== 'running' && (
          <p className="text-sm text-red-500 mt-2">
            Agent is not running. Please start the agent to send messages.
          </p>
        )}
      </div>
    </div>
  )
}