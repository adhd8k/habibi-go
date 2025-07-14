import { useEffect } from 'react'
import { useAppSelector } from '../../app/hooks'
import { selectCurrentProject } from '../../features/projects/slice/projectsSlice'
import { selectCurrentSession } from '../../features/sessions/slice/sessionsSlice'
import { useAppStore } from '../../store'
import type { Project, Session } from '../../types'

/**
 * Syncs Redux state with Zustand store for legacy components
 * This is a temporary solution while migrating to Redux
 */
export function StoreSync() {
  const reduxProject = useAppSelector(selectCurrentProject)
  const reduxSession = useAppSelector(selectCurrentSession)
  
  const setCurrentProject = useAppStore(state => state.setCurrentProject)
  const setCurrentSession = useAppStore(state => state.setCurrentSession)

  // Helper to convert null to undefined for legacy types
  const nullToUndefined = <T extends Record<string, any>>(obj: T | null): T | null => {
    if (!obj) return obj
    const result = { ...obj }
    for (const key in result) {
      if (result[key] === null) {
        result[key] = undefined as any
      } else if (typeof result[key] === 'object' && result[key] !== null && !Array.isArray(result[key])) {
        result[key] = nullToUndefined(result[key]) as any
      }
    }
    return result
  }

  // Sync Redux project to Zustand
  useEffect(() => {
    const zustandProject = nullToUndefined(reduxProject) as Project | null
    setCurrentProject(zustandProject)
  }, [reduxProject, setCurrentProject])

  // Sync Redux session to Zustand
  useEffect(() => {
    const zustandSession = nullToUndefined(reduxSession) as Session | null
    if (zustandSession) {
      console.log('Syncing session to Zustand:', zustandSession)
    }
    setCurrentSession(zustandSession)
  }, [reduxSession, setCurrentSession])

  return null
}