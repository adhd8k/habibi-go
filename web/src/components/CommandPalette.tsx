import { useState, useEffect, useRef } from 'react'
import { SlashCommand } from '../types'

interface CommandPaletteProps {
  commands: SlashCommand[]
  onSelect: (command: string, args: string) => void
  onClose: () => void
  searchTerm: string
}

export default function CommandPalette({ commands, onSelect, onClose, searchTerm }: CommandPaletteProps) {
  const [selectedIndex, setSelectedIndex] = useState(0)
  const [filteredCommands, setFilteredCommands] = useState<SlashCommand[]>([])
  const [commandArgs] = useState('')
  const listRef = useRef<HTMLDivElement>(null)

  // Filter commands based on search term
  useEffect(() => {
    const search = searchTerm.toLowerCase().slice(1) // Remove leading slash
    const filtered = commands.filter(cmd => 
      cmd.name.toLowerCase().includes(search) ||
      cmd.description.toLowerCase().includes(search)
    )
    setFilteredCommands(filtered)
    setSelectedIndex(0)
  }, [commands, searchTerm])

  // Handle keyboard navigation
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      switch (e.key) {
        case 'ArrowDown':
          e.preventDefault()
          setSelectedIndex(prev => 
            prev < filteredCommands.length - 1 ? prev + 1 : prev
          )
          break
        case 'ArrowUp':
          e.preventDefault()
          setSelectedIndex(prev => prev > 0 ? prev - 1 : prev)
          break
        case 'Enter':
          e.preventDefault()
          if (filteredCommands[selectedIndex]) {
            const cmd = filteredCommands[selectedIndex]
            const fullCommand = searchTerm.split(' ')[0]
            const args = searchTerm.slice(fullCommand.length).trim()
            onSelect(cmd.name, args || commandArgs)
          }
          break
        case 'Escape':
          e.preventDefault()
          onClose()
          break
        case 'Tab':
          e.preventDefault()
          if (filteredCommands[selectedIndex]) {
            // Tab completion - fill in the command name
            const cmd = filteredCommands[selectedIndex]
            const input = document.querySelector('textarea') as HTMLTextAreaElement
            if (input) {
              input.value = cmd.name + ' '
              input.setSelectionRange(input.value.length, input.value.length)
              const event = new Event('input', { bubbles: true })
              input.dispatchEvent(event)
            }
          }
          break
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [filteredCommands, selectedIndex, onSelect, onClose, searchTerm, commandArgs])

  // Scroll selected item into view
  useEffect(() => {
    if (listRef.current) {
      const selectedElement = listRef.current.children[selectedIndex] as HTMLElement
      if (selectedElement) {
        selectedElement.scrollIntoView({ block: 'nearest' })
      }
    }
  }, [selectedIndex])

  if (filteredCommands.length === 0) {
    return (
      <div className="absolute bottom-full left-0 right-0 mb-2 bg-gray-800 border border-gray-700 rounded-lg shadow-lg p-4">
        <p className="text-gray-400 text-sm">No commands found</p>
      </div>
    )
  }

  // Group commands by category
  const groupedCommands = filteredCommands.reduce((acc, cmd) => {
    const category = cmd.category || 'Other'
    if (!acc[category]) acc[category] = []
    acc[category].push(cmd)
    return acc
  }, {} as Record<string, SlashCommand[]>)

  let currentIndex = 0

  return (
    <div className="absolute bottom-full left-0 right-0 mb-2 bg-gray-800 border border-gray-700 rounded-lg shadow-lg max-h-80 overflow-y-auto">
      <div ref={listRef} className="py-2">
        {Object.entries(groupedCommands).map(([category, cmds]) => (
          <div key={category}>
            <div className="px-4 py-1 text-xs text-gray-500 uppercase tracking-wider">
              {category}
            </div>
            {cmds.map((cmd) => {
              const index = currentIndex++
              const isSelected = index === selectedIndex
              return (
                <div
                  key={cmd.name}
                  className={`px-4 py-2 cursor-pointer transition-colors ${
                    isSelected
                      ? 'bg-blue-600 text-white'
                      : 'hover:bg-gray-700 text-gray-200'
                  }`}
                  onClick={() => {
                    const fullCommand = searchTerm.split(' ')[0]
                    const args = searchTerm.slice(fullCommand.length).trim()
                    onSelect(cmd.name, args || commandArgs)
                  }}
                  onMouseEnter={() => setSelectedIndex(index)}
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="font-medium">
                        {cmd.name}
                        {cmd.is_builtin && (
                          <span className="ml-2 text-xs px-2 py-0.5 bg-gray-700 rounded">
                            Built-in
                          </span>
                        )}
                      </div>
                      {cmd.description && (
                        <div className={`text-sm mt-1 ${
                          isSelected ? 'text-blue-200' : 'text-gray-400'
                        }`}>
                          {cmd.description}
                        </div>
                      )}
                    </div>
                    {isSelected && cmd.arguments && cmd.arguments.length > 0 && (
                      <div className="ml-4 text-xs text-blue-200">
                        Args: {cmd.arguments.join(', ')}
                      </div>
                    )}
                  </div>
                </div>
              )
            })}
          </div>
        ))}
      </div>
      <div className="border-t border-gray-700 px-4 py-2 text-xs text-gray-400">
        <span className="font-medium">Tab</span> to complete • 
        <span className="font-medium"> ↑↓</span> to navigate • 
        <span className="font-medium"> Enter</span> to select • 
        <span className="font-medium"> Esc</span> to close
      </div>
    </div>
  )
}