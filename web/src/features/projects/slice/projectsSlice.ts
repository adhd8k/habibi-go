import { createSlice, PayloadAction } from '@reduxjs/toolkit'
import { Project } from '../../../shared/types/schemas'

interface ProjectsState {
  currentProject: Project | null
  filter: {
    searchTerm: string
    showSSHOnly: boolean
  }
}

const initialState: ProjectsState = {
  currentProject: null,
  filter: {
    searchTerm: '',
    showSSHOnly: false,
  },
}

export const projectsSlice = createSlice({
  name: 'projects',
  initialState,
  reducers: {
    setCurrentProject: (state, action: PayloadAction<Project | null>) => {
      state.currentProject = action.payload
    },
    
    setSearchTerm: (state, action: PayloadAction<string>) => {
      state.filter.searchTerm = action.payload
    },
    
    setShowSSHOnly: (state, action: PayloadAction<boolean>) => {
      state.filter.showSSHOnly = action.payload
    },
    
    clearFilters: (state) => {
      state.filter = initialState.filter
    },
  },
})

export const { 
  setCurrentProject, 
  setSearchTerm, 
  setShowSSHOnly, 
  clearFilters 
} = projectsSlice.actions

export default projectsSlice.reducer

// Selectors
export const selectCurrentProject = (state: { projects: ProjectsState }) => 
  state.projects.currentProject

export const selectProjectsFilter = (state: { projects: ProjectsState }) => 
  state.projects.filter