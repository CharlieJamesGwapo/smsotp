import { useState, useEffect } from 'react'
import api from '../api/client'
import { useAuth } from '../context/AuthContext'
import toast from 'react-hot-toast'

export default function Settings() {
  const { refreshUser } = useAuth()
  const [settings, setSettings] = useState({
    phone_gateway_url: '',
    semaphore_api_key: '',
    semaphore_enabled: 'false',
    semaphore_sender_name: '',
    sms_delay_ms: '1000',
  })
  const [passwords, setPasswords] = useState({ current_password: '', new_password: '', confirm: '' })
  const [_gatewayStatus, setGatewayStatus] = useState<'checking' | 'online' | 'offline'>('checking')

  useEffect(() => {
    api.get('/settings').then((r) => setSettings(r.data))
  }, [])

  useEffect(() => {
    if (!settings.phone_gateway_url) return
    setGatewayStatus('checking')
    // We'll just check if the setting exists; actual gateway test would need a backend endpoint
    setGatewayStatus('online')
  }, [settings.phone_gateway_url])

  const saveSettings = async () => {
    try {
      await api.put('/settings', settings)
      toast.success('Settings saved!')
    } catch {
      toast.error('Failed to save settings')
    }
  }

  const changePassword = async () => {
    if (passwords.new_password !== passwords.confirm) {
      toast.error('Passwords do not match')
      return
    }
    if (passwords.new_password.length < 6) {
      toast.error('Password must be at least 6 characters')
      return
    }
    try {
      await api.put('/auth/password', {
        current_password: passwords.current_password,
        new_password: passwords.new_password,
      })
      toast.success('Password changed!')
      setPasswords({ current_password: '', new_password: '', confirm: '' })
      refreshUser()
    } catch (e: any) {
      toast.error(e.response?.data?.error || 'Failed to change password')
    }
  }

  return (
    <div className="max-w-2xl">
      <h2 className="text-2xl font-bold mb-6">Settings</h2>

      {/* Gateway Settings */}
      <div className="bg-white rounded-xl p-6 shadow mb-6">
        <h3 className="font-bold text-lg mb-4">SMS Gateway</h3>

        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1">Phone Gateway URL</label>
            <input
              value={settings.phone_gateway_url}
              onChange={(e) => setSettings({ ...settings, phone_gateway_url: e.target.value })}
              className="w-full border rounded-lg px-3 py-2"
              placeholder="http://192.168.100.165:8080"
            />
            <p className="text-xs text-gray-400 mt-1">
              Your Android phone gateway address (primary, free)
            </p>
          </div>

          <div className="border-t pt-4">
            <label className="flex items-center gap-2 cursor-pointer mb-3">
              <input
                type="checkbox"
                checked={settings.semaphore_enabled === 'true'}
                onChange={(e) => setSettings({ ...settings, semaphore_enabled: e.target.checked ? 'true' : 'false' })}
              />
              <span className="font-medium">Enable Semaphore Fallback</span>
            </label>

            {settings.semaphore_enabled === 'true' && (
              <div className="space-y-3 ml-6">
                <div>
                  <label className="block text-sm font-medium mb-1">Semaphore API Key</label>
                  <input
                    value={settings.semaphore_api_key}
                    onChange={(e) => setSettings({ ...settings, semaphore_api_key: e.target.value })}
                    className="w-full border rounded-lg px-3 py-2"
                    type="password"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">Sender Name</label>
                  <input
                    value={settings.semaphore_sender_name}
                    onChange={(e) => setSettings({ ...settings, semaphore_sender_name: e.target.value })}
                    className="w-full border rounded-lg px-3 py-2"
                  />
                </div>
              </div>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">SMS Delay (ms)</label>
            <input
              type="number"
              value={settings.sms_delay_ms}
              onChange={(e) => setSettings({ ...settings, sms_delay_ms: e.target.value })}
              className="w-full border rounded-lg px-3 py-2"
            />
            <p className="text-xs text-gray-400 mt-1">Delay between each SMS (default: 1000ms)</p>
          </div>
        </div>

        <button
          onClick={saveSettings}
          className="mt-4 px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
        >
          Save Settings
        </button>
      </div>

      {/* Change Password */}
      <div className="bg-white rounded-xl p-6 shadow">
        <h3 className="font-bold text-lg mb-4">Change Password</h3>
        <div className="space-y-3">
          <div>
            <label className="block text-sm font-medium mb-1">Current Password</label>
            <input
              type="password"
              value={passwords.current_password}
              onChange={(e) => setPasswords({ ...passwords, current_password: e.target.value })}
              className="w-full border rounded-lg px-3 py-2"
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">New Password</label>
            <input
              type="password"
              value={passwords.new_password}
              onChange={(e) => setPasswords({ ...passwords, new_password: e.target.value })}
              className="w-full border rounded-lg px-3 py-2"
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Confirm New Password</label>
            <input
              type="password"
              value={passwords.confirm}
              onChange={(e) => setPasswords({ ...passwords, confirm: e.target.value })}
              className="w-full border rounded-lg px-3 py-2"
            />
          </div>
        </div>
        <button
          onClick={changePassword}
          className="mt-4 px-6 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700"
        >
          Change Password
        </button>
      </div>
    </div>
  )
}
