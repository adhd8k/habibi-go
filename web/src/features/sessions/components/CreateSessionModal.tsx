import { useState } from 'react'
import { useAppDispatch } from '../../../app/hooks'
import { useCreateSessionMutation } from '../api/sessionsApi'
import { setCurrentSession } from '../slice/sessionsSlice'
import { CreateSessionRequest } from '../../../shared/types/schemas'

interface CreateSessionModalProps {
  projectId: number
  defaultBranch: string
  onClose: () => void
}

export function CreateSessionModal({ projectId, defaultBranch, onClose }: CreateSessionModalProps) {
  const dispatch = useAppDispatch()
  const [createSession, { isLoading }] = useCreateSessionMutation()
  
  const [formData, setFormData] = useState({
    name: '',
    branch_name: '',
    base_branch: defaultBranch || 'main',
  })
  const [error, setError] = useState<string | null>(null)

  const generateBranchName = () => {
    if (!formData.name) return
    const branchName = formData.name
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-|-$/g, '')
    setFormData(prev => ({ ...prev, branch_name: branchName }))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)

    const request: CreateSessionRequest = {
      project_id: projectId,
      name: formData.name,
      branch_name: formData.branch_name,
      base_branch: formData.base_branch,
    }

    try {
      const newSession = await createSession(request).unwrap()
      dispatch(setCurrentSession(newSession))
      onClose()
    } catch (err: any) {
      console.error('Failed to create session:', err)
      const errorMessage = err.data?.error || err.data?.message || err.message || 'Failed to create session'
      setError(errorMessage)
    }
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 w-full max-w-md">
        <h2 className="text-xl font-semibold mb-4">Create New Session</h2>
        
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Session Name
            </label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              onBlur={generateBranchName}
              className="w-full p-2 border rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              placeholder="Feature description"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Branch Name
            </label>
            <input
              type="text"
              value={formData.branch_name}
              onChange={(e) => setFormData({ ...formData, branch_name: e.target.value })}
              className="w-full p-2 border rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              placeholder="feature-branch"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Base Branch
            </label>
            <input
              type="text"
              value={formData.base_branch}
              onChange={(e) => setFormData({ ...formData, base_branch: e.target.value })}
              className="w-full p-2 border rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              placeholder="main"
            />
            <p className="text-xs text-gray-500 mt-1">
              Branch to create the new session from
            </p>
          </div>

          {error && (
            <div className="bg-red-50 border border-red-200 rounded-md p-3">
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}

          <div className="flex justify-end gap-3 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isLoading}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isLoading ? 'Creating...' : 'Create Session'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}