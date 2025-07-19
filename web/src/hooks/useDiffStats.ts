import { useQuery } from '@tanstack/react-query'
import { sessionsApi, type DiffFile } from '../api/client'

export function useDiffStats(sessionId?: number) {
  return useQuery({
    queryKey: ['session-diffs', sessionId],
    queryFn: async () => {
      if (!sessionId) return null
      const response = await sessionsApi.getDiffs(sessionId)
      // Axios gives us response.data which contains {success: true, data: {files: [...]}}
      const files: DiffFile[] = response.data?.data?.files || []
      
      const stats = {
        filesChanged: files.length,
        additions: files.reduce((sum: number, f: DiffFile) => sum + (f.additions || 0), 0),
        deletions: files.reduce((sum: number, f: DiffFile) => sum + (f.deletions || 0), 0),
      }
      
      return stats
    },
    enabled: !!sessionId,
  })
}