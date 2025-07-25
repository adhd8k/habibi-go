import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { projectsApi } from '../../../api/client'
import { CreateProjectRequest, ProjectConfig } from '../../../types'
import { Modal } from '../../../components/ui/Modal'

interface CreateSSHProjectModalProps {
  isOpen: boolean
  onClose: () => void
  onSuccess?: () => void
}

export function CreateSSHProjectModal({ isOpen, onClose, onSuccess }: CreateSSHProjectModalProps) {
  const queryClient = useQueryClient()
  const [formData, setFormData] = useState({
    name: '',
    ssh_host: '',
    ssh_port: 22,
    ssh_key_path: '~/.ssh/id_rsa',
    remote_project_path: '',
    default_branch: 'main',
    remote_setup_cmd: '',
    environment_vars: {} as Record<string, string>,
  })
  const [envVarInput, setEnvVarInput] = useState({ key: '', value: '' })

  const createProject = useMutation({
    mutationFn: async () => {
      const config: ProjectConfig = {
        ssh_host: formData.ssh_host,
        ssh_port: formData.ssh_port,
        ssh_key_path: formData.ssh_key_path,
        remote_project_path: formData.remote_project_path,
        remote_setup_cmd: formData.remote_setup_cmd,
        environment_vars: formData.environment_vars,
      }

      const request: CreateProjectRequest = {
        name: formData.name,
        path: formData.remote_project_path, // Use remote path as the path
        default_branch: formData.default_branch,
        setup_command: formData.remote_setup_cmd, // Also store in top-level for compatibility
      }

      const response = await projectsApi.create({
        ...request,
        config,
      })
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] })
      handleClose()
      onSuccess?.()
    },
  })

  const handleSubmit = () => {
    createProject.mutate()
  }

  const handleClose = () => {
    setFormData({
      name: '',
      ssh_host: '',
      ssh_port: 22,
      ssh_key_path: '~/.ssh/id_rsa',
      remote_project_path: '',
      default_branch: 'main',
      remote_setup_cmd: '',
      environment_vars: {},
    })
    setEnvVarInput({ key: '', value: '' })
    onClose()
  }

  const addEnvironmentVar = () => {
    if (envVarInput.key && envVarInput.value) {
      setFormData(prev => ({
        ...prev,
        environment_vars: {
          ...prev.environment_vars,
          [envVarInput.key]: envVarInput.value,
        },
      }))
      setEnvVarInput({ key: '', value: '' })
    }
  }

  const removeEnvironmentVar = (key: string) => {
    setFormData(prev => {
      const { [key]: _, ...rest } = prev.environment_vars
      return {
        ...prev,
        environment_vars: rest,
      }
    })
  }

  const setupCommandHint = `Example: source ~/.bashrc && cd $WORKTREE_PATH && npm install
Available variables:
- $PROJECT_PATH: Local project path
- $WORKTREE_PATH: Remote worktree path
- $REMOTE_PROJECT_PATH: Remote project base path
- Plus any custom environment variables you define`

  const footer = (
    <>
      <button
        type="button"
        onClick={handleClose}
        className="px-4 py-2 text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-600"
      >
        Cancel
      </button>
      <button
        type="button"
        onClick={handleSubmit}
        disabled={createProject.isPending || !formData.name || !formData.ssh_host || !formData.remote_project_path}
        className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
      >
        {createProject.isPending ? 'Adding...' : 'Add SSH Project'}
      </button>
    </>
  )

  return (
    <Modal isOpen={isOpen} onClose={handleClose} title="Add SSH Project" footer={footer}>
      <div className="space-y-4 max-h-96 overflow-y-auto">
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Project Name</label>
          <input
            type="text"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
            required
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">SSH Host</label>
          <input
            type="text"
            value={formData.ssh_host}
            onChange={(e) => setFormData({ ...formData, ssh_host: e.target.value })}
            placeholder="user@hostname"
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
            required
          />
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">Format: user@hostname</p>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">SSH Port</label>
            <input
              type="number"
              value={formData.ssh_port}
              onChange={(e) => setFormData({ ...formData, ssh_port: parseInt(e.target.value) || 22 })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">SSH Key Path</label>
            <input
              type="text"
              value={formData.ssh_key_path}
              onChange={(e) => setFormData({ ...formData, ssh_key_path: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
              required
            />
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Remote Project Path</label>
          <input
            type="text"
            value={formData.remote_project_path}
            onChange={(e) => setFormData({ ...formData, remote_project_path: e.target.value })}
            placeholder="/home/user/projects/myproject"
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
            required
          />
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">Full path to the project on the remote server</p>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Default Branch</label>
          <input
            type="text"
            value={formData.default_branch}
            onChange={(e) => setFormData({ ...formData, default_branch: e.target.value })}
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Environment Variables</label>
          <div className="mt-1 space-y-2">
            {Object.entries(formData.environment_vars).map(([key, value]) => (
              <div key={key} className="flex items-center gap-2">
                <span className="text-sm font-mono bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-2 py-1 rounded">{key}={value}</span>
                <button
                  type="button"
                  onClick={() => removeEnvironmentVar(key)}
                  className="text-red-600 dark:text-red-400 hover:text-red-700 dark:hover:text-red-300 text-sm"
                >
                  Remove
                </button>
              </div>
            ))}
            <div className="flex gap-2">
              <input
                type="text"
                value={envVarInput.key}
                onChange={(e) => setEnvVarInput({ ...envVarInput, key: e.target.value })}
                placeholder="KEY"
                className="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
              />
              <input
                type="text"
                value={envVarInput.value}
                onChange={(e) => setEnvVarInput({ ...envVarInput, value: e.target.value })}
                placeholder="value"
                className="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
              />
              <button
                type="button"
                onClick={addEnvironmentVar}
                className="px-3 py-2 bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-200 rounded-md hover:bg-gray-300 dark:hover:bg-gray-600"
              >
                Add
              </button>
            </div>
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Remote Setup Command</label>
          <textarea
            value={formData.remote_setup_cmd}
            onChange={(e) => setFormData({ ...formData, remote_setup_cmd: e.target.value })}
            rows={3}
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
            placeholder="Commands to run when setting up a new session"
          />
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400 whitespace-pre-wrap">{setupCommandHint}</p>
        </div>
      </div>
    </Modal>
  )
}