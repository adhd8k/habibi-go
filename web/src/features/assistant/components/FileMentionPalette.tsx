import { useState, useEffect, useRef } from 'react'
import { File, Folder } from 'lucide-react'
import { FileMention } from '../types'
import { fileApi } from '../api/client'

interface FileMentionPaletteProps {
  sessionId: number
  searchTerm: string
  onSelect: (file: FileMention) => void
  onClose: () => void
}

export default function FileMentionPalette({ sessionId, searchTerm, onSelect, onClose }: FileMentionPaletteProps) {
  const [selectedIndex, setSelectedIndex] = useState(0)
  const [files, setFiles] = useState<FileMention[]>([])
  const [loading, setLoading] = useState(false)
  const listRef = useRef<HTMLDivElement>(null)

  // Extract search query after @
  const query = searchTerm.slice(searchTerm.lastIndexOf('@') + 1)

  // Search for files
  useEffect(() => {
    const searchFiles = async () => {
      if (query.length < 1) {
        // List root directory if no query
        setLoading(true)
        try {
          const results = await fileApi.listFiles(sessionId)
          setFiles(results)
        } catch (error) {
          console.error('Failed to list files:', error)
          setFiles([])
        } finally {
          setLoading(false)
        }
      } else {
        // Search for files
        setLoading(true)
        try {
          const results = await fileApi.searchFiles(sessionId, query)
          setFiles(results)
        } catch (error) {
          console.error('Failed to search files:', error)
          setFiles([])
        } finally {
          setLoading(false)
        }
      }
      setSelectedIndex(0)
    }

    searchFiles()
  }, [sessionId, query])

  // Handle keyboard navigation
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      switch (e.key) {
        case 'ArrowDown':
          e.preventDefault()
          setSelectedIndex(prev => 
            prev < files.length - 1 ? prev + 1 : prev
          )
          break
        case 'ArrowUp':
          e.preventDefault()
          setSelectedIndex(prev => prev > 0 ? prev - 1 : prev)
          break
        case 'Enter':
          e.preventDefault()
          if (files[selectedIndex]) {
            onSelect(files[selectedIndex])
          }
          break
        case 'Escape':
          e.preventDefault()
          onClose()
          break
        case 'Tab':
          e.preventDefault()
          if (files[selectedIndex]) {
            onSelect(files[selectedIndex])
          }
          break
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [files, selectedIndex, onSelect, onClose])

  // Scroll selected item into view
  useEffect(() => {
    if (listRef.current) {
      const selectedElement = listRef.current.children[selectedIndex] as HTMLElement
      if (selectedElement) {
        selectedElement.scrollIntoView({ block: 'nearest' })
      }
    }
  }, [selectedIndex])

  if (loading) {
    return (
      <div className="absolute bottom-full left-0 right-0 mb-2 bg-gray-800 border border-gray-700 rounded-lg shadow-lg p-4">
        <p className="text-gray-400 text-sm">Searching files...</p>
      </div>
    )
  }

  if (files.length === 0) {
    return (
      <div className="absolute bottom-full left-0 right-0 mb-2 bg-gray-800 border border-gray-700 rounded-lg shadow-lg p-4">
        <p className="text-gray-400 text-sm">No files found</p>
      </div>
    )
  }

  return (
    <div className="absolute bottom-full left-0 right-0 mb-2 bg-gray-800 border border-gray-700 rounded-lg shadow-lg max-h-80 overflow-y-auto">
      <div ref={listRef} className="py-2">
        {files.map((file, index) => {
          const isSelected = index === selectedIndex
          return (
            <div
              key={file.path}
              className={`px-4 py-2 cursor-pointer transition-colors flex items-center gap-3 ${
                isSelected
                  ? 'bg-blue-600 text-white'
                  : 'hover:bg-gray-700 text-gray-200'
              }`}
              onClick={() => onSelect(file)}
              onMouseEnter={() => setSelectedIndex(index)}
            >
              {file.type === 'directory' ? (
                <Folder size={16} className={isSelected ? 'text-blue-200' : 'text-gray-400'} />
              ) : (
                <File size={16} className={isSelected ? 'text-blue-200' : 'text-gray-400'} />
              )}
              <div className="flex-1 min-w-0">
                <div className="font-medium truncate">{file.name}</div>
                <div className={`text-xs truncate ${
                  isSelected ? 'text-blue-200' : 'text-gray-500'
                }`}>
                  {file.path}
                </div>
              </div>
              {file.size !== undefined && file.type === 'file' && (
                <div className={`text-xs ${
                  isSelected ? 'text-blue-200' : 'text-gray-500'
                }`}>
                  {formatFileSize(file.size)}
                </div>
              )}
            </div>
          )
        })}
      </div>
      <div className="border-t border-gray-700 px-4 py-2 text-xs text-gray-400">
        <span className="font-medium">Tab</span> or <span className="font-medium">Enter</span> to select • 
        <span className="font-medium"> ↑↓</span> to navigate • 
        <span className="font-medium"> Esc</span> to close
      </div>
    </div>
  )
}

function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}