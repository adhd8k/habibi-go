import React, { useState, useEffect, useRef } from 'react'
import { useMutation } from '@tanstack/react-query'
import { wsClient } from '../api/websocket'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { playNotificationSound, getNotificationsEnabled } from '../utils/notifications'
import { useAppStore } from '../store'

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

export function ClaudeChat() {
  const { currentSession } = useAppStore()
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [isProcessing, setIsProcessing] = useState(false)
  const [showToolMessages, setShowToolMessages] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  // Load chat history and setup WebSocket
  useEffect(() => {
    if (!currentSession) return

    // Ensure WebSocket is connected
    console.log('WebSocket readyState:', wsClient.getReadyState())
    if (!wsClient.isConnected()) {
      console.log('WebSocket not connected, connecting...')
      wsClient.connect()
    }

    // Load existing chat history
    const loadChatHistory = async () => {
      try {
        const response = await fetch(`/api/v1/sessions/${currentSession.id}/chat`, {
          headers: {
            'Authorization': 'Basic ' + btoa('moe:jay')
          }
        })
        if (response.ok) {
          const data = await response.json()
          if (data.data && Array.isArray(data.data)) {
            const historyMessages = data.data.map((msg: any) => ({
              id: msg.id?.toString() || Date.now().toString(),
              role: msg.role,
              content: msg.content,
              timestamp: new Date(msg.created_at),
              toolName: msg.tool_name,
              toolInput: msg.tool_input,
              toolUseId: msg.tool_use_id,
              toolContent: msg.tool_content
            }))
            setMessages(historyMessages)
          }
        }
      } catch (error) {
        console.error('Failed to load chat history:', error)
      }
    }

    loadChatHistory()

    const handleClaudeOutput = (message: any) => {
      console.log('handleClaudeOutput called with message:', message)
      if (message.data && message.data.session_id === currentSession.id) {
        console.log('Processing claude_output for session:', currentSession.id)
        const data = message.data
        const contentType = data.content_type || 'text'
        
        setMessages(prev => {
          switch (contentType) {
            case 'text': {
              const output = data.output
              const isChunk = data.is_chunk !== false // Default to true for backward compatibility
              const dbMessageId = data.db_message_id
              
              if (!output) return prev
              
              // If this is a chunk, append to last assistant message
              if (isChunk) {
                const lastMessage = prev[prev.length - 1]
                
                if (lastMessage && lastMessage.role === 'assistant') {
                  return [
                    ...prev.slice(0, -1),
                    {
                      ...lastMessage,
                      content: lastMessage.content + output
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
              } else {
                // Non-chunked message - create or replace
                const existingIndex = prev.findIndex(m => m.id === dbMessageId?.toString())
                if (existingIndex >= 0) {
                  // Update existing message
                  const updated = [...prev]
                  updated[existingIndex] = {
                    ...updated[existingIndex],
                    content: output
                  }
                  return updated
                } else {
                  // Create new message
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
              }
            }
            
            case 'tool_use': {
              return [
                ...prev,
                {
                  id: Date.now().toString(),
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
              return [
                ...prev,
                {
                  id: Date.now().toString(),
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
    }
    
    const handleResponseComplete = (message: any) => {
      if (message.data && message.data.session_id === currentSession?.id) {
        setIsProcessing(false)
        
        // Play notification sound when Claude responds
        if (getNotificationsEnabled()) {
          playNotificationSound()
        }
      }
    }
    
    const handleNewMessage = (message: any) => {
      if (message.data && message.data.session_id === currentSession.id && message.data.message) {
        const { role, content, id, created_at, tool_name, tool_input, tool_use_id, tool_content } = message.data.message
        
        setMessages(prev => {
          // For user messages, check if already exists to prevent duplicates
          if (role === 'user') {
            const exists = prev.some(m => m.content === content && m.role === 'user' && 
              Math.abs(new Date(m.timestamp).getTime() - Date.now()) < 5000) // within 5 seconds
            if (exists) return prev
          }
          
          // For assistant messages, check if we already have it from streaming
          if (role === 'assistant') {
            const exists = prev.some(m => m.role === 'assistant' && m.id === id?.toString())
            if (exists) return prev
          }
          
          return [...prev, {
            id: id?.toString() || Date.now().toString(),
            role: role as 'user' | 'assistant' | 'tool_use' | 'tool_result',
            content,
            timestamp: new Date(created_at || Date.now()),
            toolName: tool_name,
            toolInput: tool_input,
            toolUseId: tool_use_id,
            toolContent: tool_content
          }]
        })
      }
    }
    
    // Register handlers
    console.log('Registering WebSocket handlers for session:', currentSession.id)
    wsClient.on('claude_output', handleClaudeOutput)
    wsClient.on('claude_response_complete', handleResponseComplete)
    wsClient.on('new_chat_message', handleNewMessage)

    return () => {
      wsClient.off('claude_output', handleClaudeOutput)
      wsClient.off('claude_response_complete', handleResponseComplete)
      wsClient.off('new_chat_message', handleNewMessage)
    }
  }, [currentSession?.id])

  const sendMessage = useMutation({
    mutationFn: async (message: string) => {
      if (!currentSession) {
        throw new Error('No session selected')
      }
      
      // Immediately add user message to UI for instant feedback
      const userMessage: Message = {
        id: Date.now().toString(),
        role: 'user',
        content: message,
        timestamp: new Date()
      }
      
      setMessages(prev => [...prev, userMessage])
      
      // Send via WebSocket
      return new Promise((resolve, reject) => {
        const handleResponse = (msg: any) => {
          if (msg.type === 'chat_sent' && msg.data?.session_id === currentSession.id) {
            wsClient.off('chat_sent')
            wsClient.off('error')
            resolve({ success: true })
          } else if (msg.type === 'error') {
            wsClient.off('chat_sent')
            wsClient.off('error')
            reject(new Error(msg.data?.error || 'Failed to send message'))
          }
        }
        
        wsClient.on('chat_sent', handleResponse)
        wsClient.on('error', handleResponse)
        
        // Send chat message
        wsClient.send({
          type: 'session_chat',
          data: {
            session_id: currentSession.id,
            message: message
          }
        })
        
        // Timeout after 10 seconds
        setTimeout(() => {
          wsClient.off('chat_sent')
          wsClient.off('error')
          reject(new Error('Message send timeout'))
        }, 10000)
      })
    },
    onMutate: () => {
      setIsProcessing(true)
    },
    onError: () => {
      setIsProcessing(false)
    }
  })

  const handleSend = () => {
    if (!input.trim() || isProcessing) return
    
    const messageToSend = input.trim()
    sendMessage.mutate(messageToSend)
    setInput('')
  }

  // Filter messages based on show/hide tool messages setting
  const filteredMessages = messages.filter(msg => 
    showToolMessages || (msg.role !== 'tool_use' && msg.role !== 'tool_result')
  )
  
  // Count hidden tool messages between visible messages
  const getHiddenToolCount = (index: number): number => {
    if (showToolMessages) return 0
    
    let count = 0
    const startIndex = index === 0 ? 0 : messages.indexOf(filteredMessages[index - 1]) + 1
    const endIndex = messages.indexOf(filteredMessages[index])
    
    for (let i = startIndex; i < endIndex; i++) {
      if (messages[i].role === 'tool_use' || messages[i].role === 'tool_result') {
        count++
      }
    }
    return count
  }

  if (!currentSession) {
    return (
      <div className="h-full flex items-center justify-center text-gray-500 dark:text-gray-400">
        <div className="text-center">
          <p className="text-lg mb-2">No session selected</p>
          <p className="text-sm">Select a session to start chatting with Claude</p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col h-full w-full max-w-full overflow-hidden">
      <div className="border-b border-gray-200 dark:border-gray-700 p-2 flex justify-between items-center bg-gray-50 dark:bg-gray-800">
        <h3 className="font-medium text-gray-900 dark:text-gray-100">Claude Chat</h3>
        <button
          onClick={() => setShowToolMessages(!showToolMessages)}
          className={`text-xs px-2 py-1 rounded ${
            showToolMessages 
              ? 'bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300' 
              : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300'
          }`}
        >
          {showToolMessages ? 'Hide Tools' : 'Show Tools'}
        </button>
      </div>
      
      <div className="flex-1 overflow-y-auto overflow-x-hidden p-4 space-y-4 w-full max-w-full">
        {filteredMessages.length === 0 && (
          <div className="text-center text-gray-500 dark:text-gray-400 mt-8">
            <p className="text-lg mb-2">Start a conversation with Claude</p>
            <p className="text-sm">Type a message below to begin</p>
          </div>
        )}
        
        {filteredMessages.map((message, index) => {
          const hiddenToolCount = getHiddenToolCount(index)
          
          return (
            <React.Fragment key={message.id}>
              {hiddenToolCount > 0 && !showToolMessages && (
                <div className="flex justify-center my-2">
                  <button
                    onClick={() => setShowToolMessages(true)}
                    className="text-xs text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 px-3 py-1 rounded-full transition-colors"
                  >
                    🔧 {hiddenToolCount} tool {hiddenToolCount === 1 ? 'action' : 'actions'} hidden
                  </button>
                </div>
              )}
              <div
                className={`flex min-w-0 w-full max-w-full ${
                  message.role === 'user' ? 'justify-end' : 
                  message.role === 'tool_use' || message.role === 'tool_result' ? 'justify-center' :
                  'justify-start'
                }`}
              >
                <div
                  className={`max-w-lg min-w-0 w-auto rounded-lg p-3 overflow-hidden word-wrap break-words ${
                    message.role === 'user'
                      ? 'bg-blue-500 dark:bg-blue-600 text-white'
                      : message.role === 'tool_use'
                      ? 'bg-amber-50 dark:bg-amber-900 border border-amber-200 dark:border-amber-700 text-amber-900 dark:text-amber-100'
                      : message.role === 'tool_result'  
                      ? 'bg-green-50 dark:bg-green-900 border border-green-200 dark:border-green-700 text-green-900 dark:text-green-100'
                      : 'bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-gray-100'
                  }`}
                  style={{ 
                    maxWidth: '28rem', 
                    width: 'auto', 
                    wordBreak: 'break-word', 
                    overflowWrap: 'anywhere'
                  }}
                >
                  {message.role === 'user' ? (
                    <div className="whitespace-pre-wrap break-words">{message.content}</div>
                  ) : message.role === 'tool_use' ? (
                    <div>
                      <div className="flex items-center gap-2 mb-2">
                        <span className="text-xs font-medium bg-amber-200 dark:bg-amber-800 text-amber-900 dark:text-amber-100 px-2 py-1 rounded">🔧 Tool Use</span>
                        <span className="font-medium">{message.toolName}</span>
                      </div>
                      {message.toolInput && (
                        <div className="text-xs bg-amber-100 dark:bg-amber-800 text-amber-900 dark:text-amber-100 p-2 rounded overflow-hidden">
                          <pre className="whitespace-pre-wrap break-words overflow-wrap-anywhere">
                            {JSON.stringify(message.toolInput, null, 2)}
                          </pre>
                        </div>
                      )}
                    </div>
                  ) : message.role === 'tool_result' ? (
                    <div>
                      <div className="flex items-center gap-2 mb-2">
                        <span className="text-xs font-medium bg-green-200 dark:bg-green-800 text-green-900 dark:text-green-100 px-2 py-1 rounded">✅ Tool Result</span>
                      </div>
                      {message.toolContent && (
                        <div className="text-xs bg-green-100 dark:bg-green-800 text-green-900 dark:text-green-100 p-2 rounded overflow-hidden max-h-48">
                          <pre className="whitespace-pre-wrap break-words overflow-wrap-anywhere overflow-y-auto">
                            {typeof message.toolContent === 'string' 
                              ? message.toolContent 
                              : JSON.stringify(message.toolContent, null, 2)}
                          </pre>
                        </div>
                      )}
                    </div>
                  ) : (
                    <div className="prose prose-sm max-w-none dark:prose-invert overflow-hidden">
                      <ReactMarkdown 
                        remarkPlugins={[remarkGfm]}
                        components={{
                          pre: ({ ...props }) => (
                            <div className="bg-gray-800 dark:bg-gray-900 text-gray-100 p-2 rounded overflow-hidden text-sm">
                              <pre className="whitespace-pre-wrap text-xs" {...props} />
                            </div>
                          ),
                          code: ({ className, children, ...props }) => {
                            const match = /language-(\w+)/.exec(className || '')
                            return match ? (
                              <code className="bg-gray-800 dark:bg-gray-900 text-gray-100 text-xs" {...props}>
                                {children}
                              </code>
                            ) : (
                              <code className="bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-1 rounded text-xs" {...props}>
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
                    message.role === 'user' ? 'text-blue-100 dark:text-blue-200' : 
                    message.role === 'tool_use' ? 'text-amber-600 dark:text-amber-400' :
                    message.role === 'tool_result' ? 'text-green-600 dark:text-green-400' :
                    'text-gray-500 dark:text-gray-400'
                  }`}>
                    {message.timestamp.toLocaleTimeString()}
                  </div>
                </div>
              </div>
            </React.Fragment>
          )
        })}
        
        {/* Show hidden tools at the end if any */}
        {!showToolMessages && filteredMessages.length > 0 && (() => {
          const lastFilteredIndex = messages.indexOf(filteredMessages[filteredMessages.length - 1])
          let count = 0
          for (let i = lastFilteredIndex + 1; i < messages.length; i++) {
            if (messages[i].role === 'tool_use' || messages[i].role === 'tool_result') {
              count++
            }
          }
          return count > 0 ? (
            <div className="flex justify-center my-2">
              <button
                onClick={() => setShowToolMessages(true)}
                className="text-xs text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 px-3 py-1 rounded-full transition-colors"
              >
                🔧 {count} tool {count === 1 ? 'action' : 'actions'} hidden
              </button>
            </div>
          ) : null
        })()}
        
        {isProcessing && (
          <div className="flex justify-start">
            <div className="bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-gray-100 rounded-lg p-3">
              <div className="flex space-x-1">
                <div className="w-2 h-2 bg-gray-400 dark:bg-gray-500 rounded-full animate-bounce"></div>
                <div className="w-2 h-2 bg-gray-400 dark:bg-gray-500 rounded-full animate-bounce" style={{ animationDelay: '0.1s' }}></div>
                <div className="w-2 h-2 bg-gray-400 dark:bg-gray-500 rounded-full animate-bounce" style={{ animationDelay: '0.2s' }}></div>
              </div>
            </div>
          </div>
        )}
        
        <div ref={messagesEndRef} />
      </div>
      
      <div className="border-t border-gray-200 dark:border-gray-700 p-4">
        <div className="flex gap-2 items-end">
          <textarea
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyPress={(e) => e.key === 'Enter' && !e.shiftKey && handleSend()}
            placeholder="Type your message... (Shift+Enter for new line)"
            disabled={isProcessing}
            rows={1}
            className="flex-1 p-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 disabled:opacity-50 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 resize-none overflow-hidden min-h-[40px] max-h-[200px]"
            style={{
              height: 'auto',
              minHeight: '40px',
              maxHeight: '200px',
            }}
            onInput={(e) => {
              const target = e.target as HTMLTextAreaElement;
              target.style.height = 'auto';
              target.style.height = Math.min(target.scrollHeight, 200) + 'px';
            }}
          />
          <button
            onClick={handleSend}
            disabled={!input.trim() || isProcessing}
            className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed self-end"
          >
            Send
          </button>
        </div>
      </div>
    </div>
  )
}