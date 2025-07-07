class NotificationManager {
  private audioContext: AudioContext | null = null
  private isEnabled = true

  constructor() {
    // Initialize audio context on first user interaction
    this.initializeAudio()
  }

  private async initializeAudio() {
    try {
      // Create audio context when needed
      if (!this.audioContext) {
        this.audioContext = new (window.AudioContext || (window as any).webkitAudioContext)()
      }
      
      // Resume if suspended (required by some browsers)
      if (this.audioContext.state === 'suspended') {
        await this.audioContext.resume()
      }
    } catch (error) {
      console.warn('Failed to initialize audio context:', error)
    }
  }

  // Generate a pleasant ding sound using Web Audio API
  private async createDingSound(): Promise<AudioBuffer | null> {
    if (!this.audioContext) return null

    try {
      const sampleRate = this.audioContext.sampleRate
      const duration = 0.3 // 300ms
      const length = sampleRate * duration
      const buffer = this.audioContext.createBuffer(1, length, sampleRate)
      const data = buffer.getChannelData(0)

      // Generate a classic "ding" sound (bright, clean bell tone)
      for (let i = 0; i < length; i++) {
        const t = i / sampleRate
        // Primary ding tone at 1000Hz with subtle harmonic at 2000Hz
        const tone1 = Math.sin(2 * Math.PI * 1000 * t) * 0.7
        const tone2 = Math.sin(2 * Math.PI * 2000 * t) * 0.2
        
        // Sharp attack, quick decay envelope for crisp "ding"
        const envelope = Math.exp(-t * 5) * (1 - Math.exp(-t * 50))
        
        data[i] = (tone1 + tone2) * envelope
      }

      return buffer
    } catch (error) {
      console.warn('Failed to create ding sound:', error)
      return null
    }
  }

  async playDing() {
    if (!this.isEnabled) return

    try {
      await this.initializeAudio()
      if (!this.audioContext) return

      const buffer = await this.createDingSound()
      if (!buffer) return

      const source = this.audioContext.createBufferSource()
      const gainNode = this.audioContext.createGain()
      
      source.buffer = buffer
      source.connect(gainNode)
      gainNode.connect(this.audioContext.destination)
      
      // Set volume
      gainNode.gain.setValueAtTime(0.1, this.audioContext.currentTime)
      
      source.start()
    } catch (error) {
      console.warn('Failed to play ding sound:', error)
    }
  }

  setEnabled(enabled: boolean) {
    this.isEnabled = enabled
    
    // Store preference in localStorage
    localStorage.setItem('notifications-enabled', enabled.toString())
  }

  getEnabled(): boolean {
    // Get preference from localStorage
    const stored = localStorage.getItem('notifications-enabled')
    if (stored !== null) {
      this.isEnabled = stored === 'true'
    }
    return this.isEnabled
  }

  // Request permission for notifications (for future browser notifications)
  async requestPermission(): Promise<boolean> {
    if ('Notification' in window) {
      const permission = await Notification.requestPermission()
      return permission === 'granted'
    }
    return false
  }
}

// Create singleton instance
export const notificationManager = new NotificationManager()

// Utility functions
export const playNotificationSound = () => notificationManager.playDing()
export const setNotificationsEnabled = (enabled: boolean) => notificationManager.setEnabled(enabled)
export const getNotificationsEnabled = () => notificationManager.getEnabled()