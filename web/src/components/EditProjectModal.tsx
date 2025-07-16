import { useState, useEffect } from 'react'
import { Modal } from './ui/Modal'
import { useUpdateProjectMutation } from '../features/projects/api/projectsApi'
import { Project } from '../shared/types/schemas'

interface EditProjectModalProps {
  isOpen: boolean
  onClose: () => void
  project: Project | null
}

export function EditProjectModal({ isOpen, onClose, project }: EditProjectModalProps) {
  const [formData, setFormData] = useState({
    name: '',
    path: '',
    repository_url: '',
    default_branch: '',
    setup_command: ''
  })
  
  const [updateProject, { isLoading: isUpdating }] = useUpdateProjectMutation()

  useEffect(() => {
    if (project) {
      setFormData({
        name: project.name || '',
        path: project.path || '',
        repository_url: project.repository_url ?? '',
        default_branch: project.default_branch || 'main',
        setup_command: project.setup_command ?? ''
      })
    }
  }, [project])

  const handleSave = async () => {
    if (!project) return

    try {
      await updateProject({
        id: project.id,
        data: {
          name: formData.name,
          path: formData.path,
          repository_url: formData.repository_url || undefined,
          default_branch: formData.default_branch,
          setup_command: formData.setup_command || undefined
        }
      }).unwrap()
      onClose()
    } catch (error) {
      console.error('Failed to update project:', error)
      alert('Failed to update project')
    }
  }

  const handleRunScript = async () => {
    if (!project) return
    
    try {
      const response = await fetch(`/api/projects/${project.id}/run-startup-script`, {
        method: 'POST',
      })
      const data = await response.json()
      
      if (data.success) {
        alert(`Startup script output:\n\n${data.output || 'Script executed successfully'}`)
      } else {
        alert(`Failed to run startup script: ${data.error}`)
      }
    } catch (error) {
      alert(`Failed to run startup script: ${error}`)
    }
  }

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={`Edit Project - ${project?.name || ''}`}
      footer={
        <>
          <button
            onClick={handleSave}
            disabled={isUpdating}
            className="inline-flex justify-center px-4 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent rounded-md hover:bg-blue-700 focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-blue-500"
          >
            {isUpdating ? 'Saving...' : 'Save'}
          </button>
          <button
            onClick={onClose}
            className="inline-flex justify-center px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-gray-500"
          >
            Cancel
          </button>
        </>
      }
    >
      <div className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Project Name
          </label>
          <input
            type="text"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="My Project"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Project Path
          </label>
          <input
            type="text"
            value={formData.path}
            onChange={(e) => setFormData({ ...formData, path: e.target.value })}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="/home/user/projects/my-project"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Repository URL (optional)
          </label>
          <input
            type="text"
            value={formData.repository_url}
            onChange={(e) => setFormData({ ...formData, repository_url: e.target.value })}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="https://github.com/user/repo.git"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Default Branch
          </label>
          <input
            type="text"
            value={formData.default_branch}
            onChange={(e) => setFormData({ ...formData, default_branch: e.target.value })}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="main"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Startup Script (optional)
          </label>
          <p className="text-sm text-gray-600 mb-2">
            This script runs when creating new sessions. Available environment variables:
          </p>
          <ul className="text-xs text-gray-500 list-disc list-inside mb-2">
            <li>$PROJECT_PATH - Main project directory</li>
            <li>$WORKTREE_PATH - Session worktree directory</li>
            <li>$SESSION_NAME - Name of the session</li>
            <li>$BRANCH_NAME - Git branch name</li>
          </ul>
          <textarea
            value={formData.setup_command}
            onChange={(e) => setFormData({ ...formData, setup_command: e.target.value })}
            className="w-full h-32 px-3 py-2 text-sm font-mono border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="#!/bin/bash\n# Example: Install dependencies\nnpm install\n\n# Copy environment file\ncp .env.example .env"
          />
          {formData.setup_command && (
            <button
              type="button"
              onClick={handleRunScript}
              className="mt-2 text-sm text-blue-600 hover:text-blue-800"
            >
              Test Run Script →
            </button>
          )}
        </div>
      </div>
    </Modal>
  )
}