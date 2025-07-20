import { useState } from 'react'
import { useCreateProjectMutation } from '../api/projectsApi'
import { CreateProjectRequest } from '../../../shared/types/schemas'
import { z } from 'zod'
import { ProjectConfigSchema } from '../../../shared/types/schemas'

type ProjectConfig = z.infer<typeof ProjectConfigSchema>

interface AddSSHProjectFormProps {
  onSuccess?: () => void
  onCancel?: () => void
}

export function AddSSHProjectForm({ onSuccess, onCancel }: AddSSHProjectFormProps) {
  const [createProject, { isLoading, error }] = useCreateProjectMutation()
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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    const config: Partial<ProjectConfig> = {
      ssh_host: formData.ssh_host,
      ssh_port: formData.ssh_port,
      ssh_key_path: formData.ssh_key_path,
      remote_project_path: formData.remote_project_path,
      remote_setup_cmd: formData.remote_setup_cmd,
      environment_vars: formData.environment_vars,
    }

    const request: CreateProjectRequest = {
      name: formData.name,
      path: formData.remote_project_path,
      default_branch: formData.default_branch,
      setup_command: formData.remote_setup_cmd,
      config: config as ProjectConfig,
    }

    try {
      await createProject(request).unwrap()
      onSuccess?.()
    } catch (err) {
      console.error('Failed to create SSH project:', err)
    }
  }

  const addEnvironmentVar = () => {
    if (envVarInput.key && envVarInput.value) {
      setFormData({
        ...formData,
        environment_vars: {
          ...formData.environment_vars,
          [envVarInput.key]: envVarInput.value,
        },
      })
      setEnvVarInput({ key: '', value: '' })
    }
  }

  const removeEnvironmentVar = (key: string) => {
    const newVars = { ...formData.environment_vars }
    delete newVars[key]
    setFormData({ ...formData, environment_vars: newVars })
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
          Project Name
        </label>
        <input
          type="text"
          value={formData.name}
          onChange={(e) => setFormData({ ...formData, name: e.target.value })}
          className="w-full p-2 border border-gray-300 dark:border-gray-600 rounded-md focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
          required
        />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            SSH Host
          </label>
          <input
            type="text"
            value={formData.ssh_host}
            onChange={(e) => setFormData({ ...formData, ssh_host: e.target.value })}
            className="w-full p-2 border border-gray-300 dark:border-gray-600 rounded-md focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
            placeholder="example.com"
            required
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            SSH Port
          </label>
          <input
            type="number"
            value={formData.ssh_port}
            onChange={(e) => setFormData({ ...formData, ssh_port: parseInt(e.target.value) })}
            className="w-full p-2 border border-gray-300 dark:border-gray-600 rounded-md focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
            required
          />
        </div>
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
          SSH Key Path
        </label>
        <input
          type="text"
          value={formData.ssh_key_path}
          onChange={(e) => setFormData({ ...formData, ssh_key_path: e.target.value })}
          className="w-full p-2 border border-gray-300 dark:border-gray-600 rounded-md focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
          placeholder="~/.ssh/id_rsa"
          required
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
          Remote Project Path
        </label>
        <input
          type="text"
          value={formData.remote_project_path}
          onChange={(e) => setFormData({ ...formData, remote_project_path: e.target.value })}
          className="w-full p-2 border border-gray-300 dark:border-gray-600 rounded-md focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
          placeholder="/home/user/project"
          required
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
          Default Branch
        </label>
        <input
          type="text"
          value={formData.default_branch}
          onChange={(e) => setFormData({ ...formData, default_branch: e.target.value })}
          className="w-full p-2 border border-gray-300 dark:border-gray-600 rounded-md focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
          Remote Setup Script
        </label>
        <textarea
          value={formData.remote_setup_cmd}
          onChange={(e) => setFormData({ ...formData, remote_setup_cmd: e.target.value })}
          className="w-full p-2 border border-gray-300 dark:border-gray-600 rounded-md focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 font-mono text-sm"
          rows={4}
          placeholder="#!/bin/bash&#10;# Commands to run on remote server when creating session"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
          Environment Variables
        </label>
        <div className="space-y-2">
          {Object.entries(formData.environment_vars).map(([key, value]) => (
            <div key={key} className="flex items-center gap-2">
              <code className="bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-2 py-1 rounded">{key}={value}</code>
              <button
                type="button"
                onClick={() => removeEnvironmentVar(key)}
                className="text-red-600 dark:text-red-400 hover:text-red-700 dark:hover:text-red-300"
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
              placeholder="Key"
              className="flex-1 p-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
            />
            <input
              type="text"
              value={envVarInput.value}
              onChange={(e) => setEnvVarInput({ ...envVarInput, value: e.target.value })}
              placeholder="Value"
              className="flex-1 p-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
            />
            <button
              type="button"
              onClick={addEnvironmentVar}
              className="px-4 py-2 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-200 rounded-md hover:bg-gray-200 dark:hover:bg-gray-600"
            >
              Add
            </button>
          </div>
        </div>
      </div>

      {error && (
        <div className="bg-red-50 dark:bg-red-900 border border-red-200 dark:border-red-700 rounded-md p-3">
          <p className="text-sm text-red-800 dark:text-red-300">Failed to create project</p>
        </div>
      )}

      <div className="flex justify-end gap-3">
        <button
          type="button"
          onClick={onCancel}
          className="px-4 py-2 text-gray-700 dark:text-gray-200 bg-gray-100 dark:bg-gray-700 rounded-md hover:bg-gray-200 dark:hover:bg-gray-600"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={isLoading}
          className="px-4 py-2 bg-blue-600 dark:bg-blue-800 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
        >
          {isLoading ? 'Creating...' : 'Create SSH Project'}
        </button>
      </div>
    </form>
  )
}
