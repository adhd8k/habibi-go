import { useAppStore } from '../../../store'

export function DarkModeToggle() {
  const { darkMode, setDarkMode } = useAppStore()

  return (
    <button
      onClick={() => setDarkMode(!darkMode)}
      className="p-2 text-gray-600 hover:text-gray-800 dark:text-gray-400 dark:hover:text-gray-200 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
      title={darkMode ? "Switch to light mode" : "Switch to dark mode"}
    >
      <span className="text-lg">
        {darkMode ? '☀️' : '🌙'}
      </span>
    </button>
  )
}