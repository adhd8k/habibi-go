import { useState, useEffect, useRef } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { agentsApi } from '../api/client'
import { wsClient } from '../api/websocket'
import { Agent } from '../types'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { playNotificationSound, getNotificationsEnabled } from '../utils/notifications'

interface Message {
  id: string
  role: 'user' | 'assistant' | 'tool_use' | 'tool_result'
  content: string
  timestamp: Date
  toolName?: string
  toolInput?: any
  toolUseId?: string
  toolContent?: any
}

interface ClaudeChatProps {
  agent: Agent
}

export function ClaudeChat({ agent }: ClaudeChatProps) {
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [isProcessing, setIsProcessing] = useState(false)
  const [isInitialized, setIsInitialized] = useState(false)
  const [showToolMessages, setShowToolMessages] = useState(true)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const lastMessageRef = useRef<string>('')

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
    if (historyData?.messages && !isInitialized) {
      const historicalMessages: Message[] = historyData.messages.map(msg => ({
        id: msg.id.toString(),
        role: msg.role as 'user' | 'assistant' | 'tool_use' | 'tool_result',
        content: msg.content,
        timestamp: new Date(msg.created_at),
        toolName: msg.tool_name,
        toolInput: msg.tool_input ? (() => {
          try { return JSON.parse(msg.tool_input) } 
          catch { return msg.tool_input }
        })() : undefined,
        toolUseId: msg.tool_use_id,
        toolContent: msg.tool_content ? (() => {
          try { return JSON.parse(msg.tool_content) }
          catch { return msg.tool_content }
        })() : undefined
      }))
      setMessages(historicalMessages)
      setIsInitialized(true)
      
      // Store the last message to prevent duplicates
      if (historicalMessages.length > 0) {
        const lastMessage = historicalMessages[historicalMessages.length - 1]
        if (lastMessage.role === 'user') {
          lastMessageRef.current = lastMessage.content
        }
      }
    }
  }, [historyData, isInitialized])

  useEffect(() => {
    // Subscribe to agent output
    wsClient.subscribe(agent.id)
    
    wsClient.on('agent_output', (message) => {
      if (message.agent_id === agent.id && message.data) {
        const data = message.data
        const contentType = data.content_type || 'text'
        const dbMessageId = data.db_message_id
        
        setMessages(prev => {
          switch (contentType) {
            case 'text': {
              const output = data.output
              const isChunk = data.is_chunk
              
              // Skip empty output
              if (!output) {
                return prev
              }
              
              // If this is a chunk, find the message by db_message_id or append to last assistant message
              if (isChunk) {
                // Try to find existing message by db_message_id
                if (dbMessageId) {
                  const existingIndex = prev.findIndex(msg => 
                    msg.role === 'assistant' && msg.id === dbMessageId.toString()
                  )
                  
                  if (existingIndex >= 0) {
                    // Update existing message
                    const updatedMessages = [...prev]
                    updatedMessages[existingIndex] = {
                      ...updatedMessages[existingIndex],
                      content: updatedMessages[existingIndex].content + output
                    }
                    return updatedMessages
                  }
                }
                
                const lastMessage = prev[prev.length - 1]
                
                // If we have an assistant message being built, append seamlessly
                if (lastMessage && lastMessage.role === 'assistant') {
                  return [
                    ...prev.slice(0, -1),
                    {
                      ...lastMessage,
                      content: lastMessage.content + output,
                      id: dbMessageId?.toString() || lastMessage.id
                    }
                  ]
                }
                
                // Start a new assistant message
                return [
                  ...prev,
                  {
                    id: dbMessageId?.toString() || Date.now().toString(),
                    role: 'assistant',
                    content: output,
                    timestamp: new Date()
                  }
                ]
              }
              return prev
            }
            
            case 'tool_use': {
              // Add tool use message
              return [
                ...prev,
                {
                  id: dbMessageId?.toString() || Date.now().toString(),
                  role: 'tool_use',
                  content: '',
                  timestamp: new Date(),
                  toolName: data.tool_name,
                  toolInput: data.tool_input,
                  toolUseId: data.tool_use_id
                }
              ]
            }
            
            case 'tool_result': {
              // Add tool result message
              return [
                ...prev,
                {
                  id: dbMessageId?.toString() || Date.now().toString(),
                  role: 'tool_result',
                  content: '',
                  timestamp: new Date(),
                  toolUseId: data.tool_use_id,
                  toolContent: data.tool_content
                }
              ]
            }
            
            default:
              return prev
          }
        })
      }
    })
    
    // Listen for response completion
    wsClient.on('agent_response_complete', (message) => {
      if (message.agent_id === agent.id) {
        setIsProcessing(false)
        
        // Play notification sound when Claude responds
        if (getNotificationsEnabled()) {
          playNotificationSound()
        }
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
    
    const messageToSend = input.trim()
    
    // Prevent sending the same message twice
    if (messageToSend === lastMessageRef.current) {
      console.log('Preventing duplicate message send')
      return
    }
    
    lastMessageRef.current = messageToSend
    sendMessage.mutate(messageToSend)
    setInput('')
  }

  // Filter messages based on show/hide tool messages setting
  const filteredMessages = messages.filter(msg => 
    showToolMessages || (msg.role !== 'tool_use' && msg.role !== 'tool_result')
  )

  return (
    <div className="flex flex-col h-full w-full max-w-full overflow-hidden" style={{ width: '100%', maxWidth: '100vw' }}>
      <div className="border-b p-2 flex justify-between items-center bg-gray-50">
        <h3 className="font-medium">Claude Chat</h3>
        <button
          onClick={() => setShowToolMessages(!showToolMessages)}
          className={`text-xs px-2 py-1 rounded ${
            showToolMessages 
              ? 'bg-blue-100 text-blue-700' 
              : 'bg-gray-100 text-gray-600'
          }`}
        >
          {showToolMessages ? 'Hide Tools' : 'Show Tools'}
        </button>
      </div>
      
      <div className="flex-1 overflow-y-auto overflow-x-hidden p-4 space-y-4 w-full max-w-full" style={{ width: '100%', maxWidth: '100%' }}>
        {filteredMessages.length === 0 && (
          <div className="text-center text-gray-500 mt-8">
            <p className="text-lg mb-2">Start a conversation with Claude</p>
            <p className="text-sm">Type a message below to begin</p>
          </div>
        )}
        
        {filteredMessages.map((message) => (
          <div
            key={message.id}
            className={`flex min-w-0 w-full max-w-full ${
              message.role === 'user' ? 'justify-end' : 
              message.role === 'tool_use' || message.role === 'tool_result' ? 'justify-center' :
              'justify-start'
            }`}
          >
            <div
              className={`max-w-lg min-w-0 w-auto rounded-lg p-3 overflow-hidden word-wrap break-words ${
                message.role === 'user'
                  ? 'bg-blue-500 text-white'
                  : message.role === 'tool_use'
                  ? 'bg-amber-50 border border-amber-200 text-amber-900'
                  : message.role === 'tool_result'  
                  ? 'bg-green-50 border border-green-200 text-green-900'
                  : 'bg-gray-100 text-gray-900'
              }`}
              style={{ 
                maxWidth: '28rem', 
                width: 'auto', 
                wordBreak: 'break-word', 
                overflowWrap: 'anywhere',
                tableLayout: 'fixed' 
              }}
            >
              {message.role === 'user' ? (
                <div className="whitespace-pre-wrap break-words">{message.content}</div>
              ) : message.role === 'tool_use' ? (
                <div>
                  <div className="flex items-center gap-2 mb-2">
                    <span className="text-xs font-medium bg-amber-200 px-2 py-1 rounded">🔧 Tool Use</span>
                    <span className="font-medium">{message.toolName}</span>
                  </div>
                  {message.toolInput && (
                    <div className="text-xs bg-amber-100 p-2 rounded overflow-hidden">
                      <pre className="whitespace-pre-wrap break-words overflow-wrap-anywhere">
                        {JSON.stringify(message.toolInput, null, 2)}
                      </pre>
                    </div>
                  )}
                </div>
              ) : message.role === 'tool_result' ? (
                <div>
                  <div className="flex items-center gap-2 mb-2">
                    <span className="text-xs font-medium bg-green-200 px-2 py-1 rounded">✅ Tool Result</span>
                    {message.toolUseId && (
                      <span className="text-xs text-green-600">ID: {message.toolUseId}</span>
                    )}
                  </div>
                  {message.toolContent && (
                    <div className="text-xs bg-green-100 p-2 rounded overflow-hidden max-h-48">
                      <pre className="whitespace-pre-wrap break-words overflow-wrap-anywhere overflow-y-auto">
                        {typeof message.toolContent === 'string' 
                          ? message.toolContent 
                          : JSON.stringify(message.toolContent, null, 2)}
                      </pre>
                    </div>
                  )}
                </div>
              ) : (
                <div className="prose prose-sm max-w-none dark:prose-invert overflow-hidden" style={{ wordBreak: 'break-word', overflowWrap: 'anywhere' }}>
                  <ReactMarkdown 
                    remarkPlugins={[remarkGfm]}
                    components={{
                      pre: ({ ...props }) => (
                        <div className="bg-gray-800 text-gray-100 p-2 rounded overflow-hidden text-sm">
                          <pre className="whitespace-pre-wrap text-xs" style={{ wordBreak: 'break-word', overflowWrap: 'anywhere' }} {...props} />
                        </div>
                      ),
                      code: ({ className, children, ...props }) => {
                        const match = /language-(\w+)/.exec(className || '')
                        return match ? (
                          <code className="bg-gray-800 text-gray-100 text-xs" style={{ wordBreak: 'break-word', overflowWrap: 'anywhere' }} {...props}>
                            {children}
                          </code>
                        ) : (
                          <code className="bg-gray-200 px-1 rounded text-xs" style={{ wordBreak: 'break-word', overflowWrap: 'anywhere' }} {...props}>
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
                message.role === 'user' ? 'text-blue-100' : 
                message.role === 'tool_use' ? 'text-amber-600' :
                message.role === 'tool_result' ? 'text-green-600' :
                'text-gray-500'
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