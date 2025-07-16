import { useState, useEffect } from 'react'
import { Modal } from './ui/Modal'
import { useUpdateProjectMutation } from '../features/projects/api/projectsApi'
import { Project } from '../shared/types/schemas'

interface EditStartupScriptModalProps {
  isOpen: boolean
  onClose: () => void
  project: Project | null
}

export function EditStartupScriptModal({ isOpen, onClose, project }: EditStartupScriptModalProps) {
  const [script, setScript] = useState('')
  const [updateProject, { isLoading: isUpdating }] = useUpdateProjectMutation()

  useEffect(() => {
    if (project) {
      setScript(project.setup_command ?? '')
    }
  }, [project])

  const handleSave = async () => {
    if (!project) return

    try {
      await updateProject({
        id: project.id,
        data: { setup_command: script }
      }).unwrap()
      onClose()
    } catch (error) {
      console.error('Failed to update startup script:', error)
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
      title={`Edit Startup Script - ${project?.name || ''}`}
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
            onClick={handleRunScript}
            className="inline-flex justify-center px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-gray-500"
          >
            Run Script
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
          <p className="text-sm text-gray-600 mb-2">
            This script runs when creating new sessions. Available environment variables:
          </p>
          <ul className="text-xs text-gray-500 list-disc list-inside mb-4">
            <li>$PROJECT_PATH - Main project directory</li>
            <li>$WORKTREE_PATH - Session worktree directory</li>
            <li>$SESSION_NAME - Name of the session</li>
            <li>$BRANCH_NAME - Git branch name</li>
          </ul>
        </div>
        <textarea
          value={script}
          onChange={(e) => setScript(e.target.value)}
          className="w-full h-64 px-3 py-2 text-sm font-mono border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="#!/bin/bash&#10;# Example: Install dependencies&#10;npm install&#10;&#10;# Copy environment file&#10;cp .env.example .env"
        />
      </div>
    </Modal>
  )
}