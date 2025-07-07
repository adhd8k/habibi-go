import { useState, useEffect } from 'react'
import { getNotificationsEnabled, setNotificationsEnabled, playNotificationSound } from '../utils/notifications'

export function NotificationSettings() {
  const [isEnabled, setIsEnabled] = useState(getNotificationsEnabled())
  const [isOpen, setIsOpen] = useState(false)

  useEffect(() => {
    setIsEnabled(getNotificationsEnabled())
  }, [])

  const handleToggle = (enabled: boolean) => {
    setIsEnabled(enabled)
    setNotificationsEnabled(enabled)
    
    // Play a test sound when enabling
    if (enabled) {
      // Small delay to ensure the setting is applied
      setTimeout(() => {
        playNotificationSound()
      }, 100)
    }
  }

  return (
    <div className="relative">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="p-2 text-gray-600 hover:text-gray-800 rounded-lg hover:bg-gray-100 transition-colors"
        title="Notification Settings"
      >
        <span className="text-lg">{isEnabled ? '🔔' : '🔕'}</span>
      </button>

      {isOpen && (
        <>
          {/* Overlay to close dropdown when clicking outside */}
          <div 
            className="fixed inset-0 z-10" 
            onClick={() => setIsOpen(false)}
          />
          
          {/* Dropdown panel */}
          <div className="absolute top-full right-0 mt-2 w-64 bg-white rounded-lg shadow-lg border border-gray-200 z-20">
            <div className="p-4">
              <h3 className="text-sm font-medium text-gray-900 mb-3">
                Notification Settings
              </h3>
              
              <div className="space-y-3">
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={isEnabled}
                    onChange={(e) => handleToggle(e.target.checked)}
                    className="rounded border-gray-300 text-blue-600 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50"
                  />
                  <span className="ml-2 text-sm text-gray-700">
                    Play sound when Claude responds
                  </span>
                </label>
                
                <div className="text-xs text-gray-500">
                  Get audio notifications when Claude completes a response in any session.
                </div>
                
                {isEnabled && (
                  <button
                    onClick={() => playNotificationSound()}
                    className="text-xs px-2 py-1 bg-blue-100 text-blue-700 rounded hover:bg-blue-200 transition-colors"
                  >
                    Test Sound
                  </button>
                )}
              </div>
            </div>
          </div>
        </>
      )}
    </div>
  )
}