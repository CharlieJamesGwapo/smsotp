import { useState, useEffect } from 'react'
import api from '../api/client'
import toast from 'react-hot-toast'

interface Contact { id: number; name: string; phone: string }
interface Group { id: number; name: string; member_count: number }
interface Template { id: number; name: string; body: string }

function charInfo(text: string) {
  const len = text.length
  if (len === 0) return ''
  const parts = Math.ceil(len / 160)
  return `${len} chars (${parts} SMS part${parts > 1 ? 's' : ''})`
}

export default function SendSMS() {
  const [contacts, setContacts] = useState<Contact[]>([])
  const [groups, setGroups] = useState<Group[]>([])
  const [templates, setTemplates] = useState<Template[]>([])
  const [selectedContacts, setSelectedContacts] = useState<number[]>([])
  const [selectedGroups, setSelectedGroups] = useState<number[]>([])
  const [templateId, setTemplateId] = useState<number | null>(null)
  const [body, setBody] = useState('')
  const [variables, setVariables] = useState<Record<string, string>>({})
  const [schedule, setSchedule] = useState(false)
  const [scheduledAt, setScheduledAt] = useState('')
  const [sending, setSending] = useState(false)
  const [step, setStep] = useState<'compose' | 'preview'>('compose')
  const [contactSearch, setContactSearch] = useState('')

  useEffect(() => {
    Promise.all([
      api.get('/contacts?per_page=100'),
      api.get('/groups?per_page=100'),
      api.get('/templates?per_page=100'),
    ]).then(([c, g, t]) => {
      setContacts(c.data.data || [])
      setGroups(g.data.data || [])
      setTemplates(t.data.data || [])
    })
  }, [])

  const toggleContact = (id: number) => {
    setSelectedContacts((prev) =>
      prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]
    )
  }

  const toggleGroup = (id: number) => {
    setSelectedGroups((prev) =>
      prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]
    )
  }

  const selectTemplate = (id: number) => {
    const t = templates.find((x) => x.id === id)
    if (t) {
      setTemplateId(t.id)
      setBody(t.body)
    }
  }

  const messageText = templateId ? templates.find((t) => t.id === templateId)?.body || body : body
  const recipientCount = selectedContacts.length + selectedGroups.reduce((sum, gid) => {
    const g = groups.find((x) => x.id === gid)
    return sum + (g?.member_count || 0)
  }, 0)

  const send = async () => {
    setSending(true)
    try {
      const payload: any = {
        contact_ids: selectedContacts,
        group_ids: selectedGroups,
        variables,
      }
      if (templateId) payload.template_id = templateId
      else payload.body = body

      if (schedule && scheduledAt) {
        payload.scheduled_at = new Date(scheduledAt).toISOString()
        const res = await api.post('/messages/schedule', payload)
        toast.success(`${res.data.scheduled} messages scheduled!`)
      } else {
        const res = await api.post('/messages/send', payload)
        toast.success(`Sent: ${res.data.sent}, Failed: ${res.data.failed}`)
      }

      // Reset
      setSelectedContacts([])
      setSelectedGroups([])
      setTemplateId(null)
      setBody('')
      setVariables({})
      setSchedule(false)
      setScheduledAt('')
      setStep('compose')
    } catch (e: any) {
      toast.error(e.response?.data?.error || 'Send failed')
    } finally {
      setSending(false)
    }
  }

  const filteredContacts = contacts.filter((c) =>
    c.name.toLowerCase().includes(contactSearch.toLowerCase()) ||
    c.phone.includes(contactSearch)
  )

  if (step === 'preview') {
    return (
      <div>
        <h2 className="text-2xl font-bold mb-6">Preview & Send</h2>
        <div className="bg-white rounded-xl p-6 shadow max-w-2xl">
          <div className="mb-4">
            <h4 className="font-medium text-gray-500 mb-1">Recipients</h4>
            <p>{recipientCount} contact(s)</p>
            <div className="text-sm text-gray-500 mt-1">
              {selectedContacts.map((id) => contacts.find((c) => c.id === id)?.name).filter(Boolean).join(', ')}
              {selectedGroups.length > 0 && (
                <span> + Groups: {selectedGroups.map((id) => groups.find((g) => g.id === id)?.name).filter(Boolean).join(', ')}</span>
              )}
            </div>
          </div>
          <div className="mb-4">
            <h4 className="font-medium text-gray-500 mb-1">Message</h4>
            <div className="bg-gray-50 p-3 rounded-lg whitespace-pre-wrap">{messageText}</div>
            <p className="text-xs text-gray-400 mt-1">{charInfo(messageText)}</p>
          </div>
          {schedule && (
            <div className="mb-4">
              <h4 className="font-medium text-gray-500 mb-1">Scheduled For</h4>
              <p>{new Date(scheduledAt).toLocaleString()}</p>
            </div>
          )}
          {Object.keys(variables).length > 0 && (
            <div className="mb-4">
              <h4 className="font-medium text-gray-500 mb-1">Custom Variables</h4>
              {Object.entries(variables).map(([k, v]) => (
                <p key={k} className="text-sm">{`{${k}}`} = {v}</p>
              ))}
            </div>
          )}
          <div className="flex gap-2 mt-6">
            <button onClick={() => setStep('compose')} className="px-4 py-2 bg-gray-200 rounded-lg">Back</button>
            <button
              onClick={send}
              disabled={sending}
              className="px-6 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50"
            >
              {sending ? 'Sending...' : schedule ? 'Schedule Messages' : 'Send Now'}
            </button>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div>
      <h2 className="text-2xl font-bold mb-6">Send SMS</h2>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Recipients */}
        <div className="bg-white rounded-xl p-5 shadow">
          <h3 className="font-bold mb-3">Select Recipients</h3>

          {/* Groups */}
          {groups.length > 0 && (
            <div className="mb-4">
              <h4 className="text-sm font-medium text-gray-500 mb-2">Groups</h4>
              <div className="space-y-1">
                {groups.map((g) => (
                  <label key={g.id} className="flex items-center gap-2 px-3 py-2 hover:bg-gray-50 rounded cursor-pointer">
                    <input
                      type="checkbox"
                      checked={selectedGroups.includes(g.id)}
                      onChange={() => toggleGroup(g.id)}
                    />
                    <span>{g.name}</span>
                    <span className="text-xs text-gray-400">({g.member_count})</span>
                  </label>
                ))}
              </div>
            </div>
          )}

          {/* Individual contacts */}
          <h4 className="text-sm font-medium text-gray-500 mb-2">Individual Contacts</h4>
          <input
            type="text"
            placeholder="Search contacts..."
            value={contactSearch}
            onChange={(e) => setContactSearch(e.target.value)}
            className="w-full border rounded-lg px-3 py-2 mb-2 text-sm"
          />
          <div className="space-y-1 max-h-60 overflow-y-auto">
            {filteredContacts.map((c) => (
              <label key={c.id} className="flex items-center gap-2 px-3 py-1 hover:bg-gray-50 rounded cursor-pointer">
                <input
                  type="checkbox"
                  checked={selectedContacts.includes(c.id)}
                  onChange={() => toggleContact(c.id)}
                />
                <span className="text-sm">{c.name}</span>
                <span className="text-xs text-gray-400">{c.phone}</span>
              </label>
            ))}
          </div>

          <p className="text-sm text-gray-500 mt-3">
            Selected: {selectedContacts.length} contacts, {selectedGroups.length} groups (~{recipientCount} recipients)
          </p>
        </div>

        {/* Message */}
        <div className="bg-white rounded-xl p-5 shadow">
          <h3 className="font-bold mb-3">Compose Message</h3>

          {templates.length > 0 && (
            <div className="mb-4">
              <label className="text-sm font-medium text-gray-500">Use Template</label>
              <select
                value={templateId || ''}
                onChange={(e) => {
                  const val = e.target.value
                  if (val) selectTemplate(Number(val))
                  else { setTemplateId(null); setBody('') }
                }}
                className="w-full border rounded-lg px-3 py-2 mt-1"
              >
                <option value="">-- Custom Message --</option>
                {templates.map((t) => (
                  <option key={t.id} value={t.id}>{t.name}</option>
                ))}
              </select>
            </div>
          )}

          <div className="mb-4">
            <label className="text-sm font-medium text-gray-500">Message Body</label>
            <textarea
              value={body}
              onChange={(e) => { setBody(e.target.value); setTemplateId(null) }}
              placeholder="Type your message... Use {name}, {phone}, {email}, {code}"
              className="w-full border rounded-lg px-3 py-2 mt-1"
              rows={5}
            />
            <div className="flex justify-between text-xs text-gray-400 mt-1">
              <span>Variables: {'{name}'}, {'{phone}'}, {'{email}'}, {'{code}'}</span>
              <span className={body.length > 160 ? 'text-orange-500' : ''}>{charInfo(body)}</span>
            </div>
          </div>

          {/* Custom variable input for {code} etc */}
          <div className="mb-4">
            <label className="text-sm font-medium text-gray-500">Custom Variable: code</label>
            <input
              type="text"
              placeholder="Value for {code} (optional)"
              value={variables.code || ''}
              onChange={(e) => setVariables({ ...variables, code: e.target.value })}
              className="w-full border rounded-lg px-3 py-2 mt-1"
            />
          </div>

          {/* Schedule */}
          <div className="mb-4">
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={schedule}
                onChange={(e) => setSchedule(e.target.checked)}
              />
              <span className="text-sm font-medium">Schedule for later</span>
            </label>
            {schedule && (
              <input
                type="datetime-local"
                value={scheduledAt}
                onChange={(e) => setScheduledAt(e.target.value)}
                className="w-full border rounded-lg px-3 py-2 mt-2"
              />
            )}
          </div>

          <button
            onClick={() => setStep('preview')}
            disabled={
              (selectedContacts.length === 0 && selectedGroups.length === 0) ||
              (!templateId && !body) ||
              (schedule && !scheduledAt)
            }
            className="w-full py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Preview Message
          </button>
        </div>
      </div>
    </div>
  )
}
