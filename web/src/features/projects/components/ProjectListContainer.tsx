import { useAppDispatch, useAppSelector } from '../../../app/hooks'
import { useGetProjectsQuery, useDeleteProjectMutation } from '../api/projectsApi'
import { setCurrentProject, selectCurrentProject } from '../slice/projectsSlice'
import { clearCurrentSession } from '../../sessions/slice/sessionsSlice'
import { ProjectListView } from './ProjectListView'

export function ProjectListContainer() {
  const dispatch = useAppDispatch()
  const currentProject = useAppSelector(selectCurrentProject)
  
  const { data: projects, isLoading, error } = useGetProjectsQuery()
  const [deleteProject, { isLoading: isDeleting }] = useDeleteProjectMutation()
  
  const handleSelectProject = (project: typeof projects[0]) => {
    dispatch(setCurrentProject(project))
    dispatch(clearCurrentSession())
  }
  
  const handleDeleteProject = async (id: number) => {
    const project = projects?.find(p => p.id === id)
    if (!project) return
    
    if (confirm(`Are you sure you want to delete project "${project.name}"? This will also delete all sessions.`)) {
      try {
        await deleteProject(id).unwrap()
        
        // Clear current project if it was deleted
        if (currentProject?.id === id) {
          dispatch(setCurrentProject(null))
          dispatch(clearCurrentSession())
        }
      } catch (error) {
        console.error('Failed to delete project:', error)
      }
    }
  }
  
  return (
    <ProjectListView
      projects={projects || []}
      currentProject={currentProject}
      isLoading={isLoading}
      error={error}
      isDeleting={isDeleting}
      onSelectProject={handleSelectProject}
      onDeleteProject={handleDeleteProject}
    />
  )
}