import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { sessionsApi } from '../api/client'
import { useAppStore } from '../store'
import { Session, CreateSessionRequest } from '../types'

export function SessionManager() {
  const queryClient = useQueryClient()
  const { currentProject, currentSession, setCurrentSession } = useAppStore()
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [newSession, setNewSession] = useState({
    session_name: '',
    branch_name: '',
  })

  const { data: sessions, isLoading } = useQuery({
    queryKey: ['sessions', currentProject?.id],
    queryFn: async () => {
      if (!currentProject) return []
      const response = await sessionsApi.list(currentProject.id)
      // Handle the wrapped response format {data: [...], success: true}
      const data = response.data as any
      if (data && data.data && Array.isArray(data.data)) {
        return data.data
      }
      // Fallback to direct array if API format changes
      return Array.isArray(response.data) ? response.data : []
    },
    enabled: !!currentProject,
  })

  const createMutation = useMutation({
    mutationFn: async (data: CreateSessionRequest) => {
      const response = await sessionsApi.create(data)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sessions'] })
      setShowCreateForm(false)
      setNewSession({ session_name: '', branch_name: '' })
    },
  })

  const handleCreateSession = () => {
    if (!currentProject || !newSession.session_name || !newSession.branch_name) return
    
    createMutation.mutate({
      project_name: currentProject.name,
      session_name: newSession.session_name,
      branch_name: newSession.branch_name,
    })
  }

  if (!currentProject) {
    return (
      <div className="p-4 text-gray-500">
        Select a project to view sessions
      </div>
    )
  }

  return (
    <div className="p-4">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-lg font-semibold">Sessions</h2>
        <button
          onClick={() => setShowCreateForm(!showCreateForm)}
          className="px-3 py-1 bg-blue-500 text-white rounded hover:bg-blue-600 text-sm"
        >
          New Session
        </button>
      </div>

      {showCreateForm && (
        <div className="mb-4 p-3 bg-gray-50 rounded-lg">
          <input
            type="text"
            placeholder="Session name"
            value={newSession.session_name}
            onChange={(e) => setNewSession({ ...newSession, session_name: e.target.value })}
            className="w-full p-2 border rounded mb-2"
          />
          <input
            type="text"
            placeholder="Branch name"
            value={newSession.branch_name}
            onChange={(e) => setNewSession({ ...newSession, branch_name: e.target.value })}
            className="w-full p-2 border rounded mb-2"
          />
          <div className="flex gap-2">
            <button
              onClick={handleCreateSession}
              disabled={createMutation.isPending}
              className="px-3 py-1 bg-green-500 text-white rounded hover:bg-green-600 text-sm disabled:opacity-50"
            >
              Create
            </button>
            <button
              onClick={() => {
                setShowCreateForm(false)
                setNewSession({ session_name: '', branch_name: '' })
              }}
              className="px-3 py-1 bg-gray-300 text-gray-700 rounded hover:bg-gray-400 text-sm"
            >
              Cancel
            </button>
          </div>
        </div>
      )}

      {isLoading ? (
        <div className="animate-pulse">
          <div className="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
          <div className="h-4 bg-gray-200 rounded w-1/2"></div>
        </div>
      ) : (
        <div className="space-y-2">
          {sessions?.map((session: Session) => (
            <button
              key={session.id}
              onClick={() => setCurrentSession(session)}
              className={`w-full text-left p-3 rounded-lg transition-colors ${
                currentSession?.id === session.id
                  ? 'bg-green-100 border-green-500 border'
                  : 'bg-gray-50 hover:bg-gray-100 border-gray-200 border'
              }`}
            >
              <div className="flex justify-between items-start">
                <div>
                  <div className="font-medium">{session.name}</div>
                  <div className="text-sm text-gray-600">{session.branch_name}</div>
                </div>
                <span className={`text-xs px-2 py-1 rounded ${
                  session.status === 'active' ? 'bg-green-200 text-green-800' :
                  session.status === 'paused' ? 'bg-yellow-200 text-yellow-800' :
                  'bg-gray-200 text-gray-800'
                }`}>
                  {session.status}
                </span>
              </div>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}