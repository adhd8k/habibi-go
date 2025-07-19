import { useState } from 'react'
import { useCreateProjectMutation } from '../api/projectsApi'
import { CreateProjectRequest } from '../../../shared/types/schemas'

interface CreateProjectFormProps {
  onSuccess: () => void
  onCancel: () => void
}

export function CreateProjectForm({ onSuccess, onCancel }: CreateProjectFormProps) {
  const [createProject, { isLoading, error }] = useCreateProjectMutation()

  const [formData, setFormData] = useState<CreateProjectRequest>({
    name: '',
    path: '',
    setup_command: '',
  })

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!formData.name || !formData.path) {
      return
    }

    try {
      await createProject(formData).unwrap()
      onSuccess()
    } catch (err) {
      console.error('Failed to create project:', err)
    }
  }

  return (
    <form onSubmit={handleSubmit}>
      {error && (
        <div className="mb-3 p-2 bg-red-100 dark:bg-red-900 border border-red-400 dark:border-red-700 text-red-700 dark:text-red-300 rounded text-sm">
          Failed to create project
        </div>
      )}

      <input
        type="text"
        placeholder="Project name"
        value={formData.name}
        onChange={(e) => setFormData({ ...formData, name: e.target.value })}
        className="w-full p-2 border border-gray-300 dark:border-gray-600 rounded mb-2 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
        required
      />

      <div className="relative mb-2">
        <input
          type="text"
          placeholder="Project path (e.g., /home/user/myproject)"
          value={formData.path}
          onChange={(e) => setFormData({ ...formData, path: e.target.value })}
          className="w-full p-2 pr-10 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
          required
        />
        <button
          type="button"
          onClick={() => {
            alert('Please enter the full path to your project directory.\\n\\nExample:\\n/home/user/projects/myproject\\n\\nNote: Browser security prevents automatic folder browsing.')
          }}
          className="absolute right-2 top-2 p-1 text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200"
          title="Browse for folder"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
          </svg>
        </button>
      </div>

      <div className="mb-2">
        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
          Setup Script (optional)
        </label>
        <textarea
          placeholder="#!/bin/bash\n# This script runs when creating a new session\n# Available variables:\n# $PROJECT_PATH - Main project directory\n# $WORKTREE_PATH - Session worktree directory\n# $SESSION_NAME - Name of the session\n# $BRANCH_NAME - Git branch name\n\n# Example:\ncp $PROJECT_PATH/.env $WORKTREE_PATH/.env"
          value={formData.setup_command || ''}
          onChange={(e) => setFormData({ ...formData, setup_command: e.target.value })}
          className="w-full p-2 border border-gray-300 dark:border-gray-600 rounded font-mono text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
          rows={6}
        />
        <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
          This script will be executed each time a new session is created
        </p>
      </div>

      <div className="flex gap-2">
        <button
          type="submit"
          disabled={isLoading}
          className="px-3 py-1 bg-green-500 dark:bg-green-800 text-white rounded hover:bg-green-600 text-sm disabled:opacity-50"
        >
          {isLoading ? 'Creating...' : 'Create'}
        </button>
        <button
          type="button"
          onClick={onCancel}
          className="px-3 py-1 bg-gray-300 dark:bg-gray-600 text-gray-700 dark:text-gray-200 rounded hover:bg-gray-400 dark:hover:bg-gray-500 text-sm"
        >
          Cancel
        </button>
      </div>
    </form>
  )
}
