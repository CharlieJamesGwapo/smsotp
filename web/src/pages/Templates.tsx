import { useState, useEffect } from 'react'
import api from '../api/client'
import toast from 'react-hot-toast'

interface Template {
  id: number
  name: string
  body: string
}

function charInfo(text: string) {
  const len = text.length
  if (len === 0) return ''
  const limit = 160
  const parts = Math.ceil(len / limit)
  return `${len} chars (${parts} SMS part${parts > 1 ? 's' : ''})`
}

export default function Templates() {
  const [templates, setTemplates] = useState<Template[]>([])
  const [showModal, setShowModal] = useState(false)
  const [editing, setEditing] = useState<Template | null>(null)
  const [form, setForm] = useState({ name: '', body: '' })

  const load = () => {
    api.get('/templates?per_page=100').then((r) => setTemplates(r.data.data || []))
  }

  useEffect(() => { load() }, [])

  const openCreate = () => {
    setEditing(null)
    setForm({ name: '', body: '' })
    setShowModal(true)
  }

  const openEdit = (t: Template) => {
    setEditing(t)
    setForm({ name: t.name, body: t.body })
    setShowModal(true)
  }

  const save = async () => {
    try {
      if (editing) {
        await api.put(`/templates/${editing.id}`, form)
        toast.success('Template updated')
      } else {
        await api.post('/templates', form)
        toast.success('Template created')
      }
      setShowModal(false)
      load()
    } catch (e: any) {
      toast.error(e.response?.data?.error || 'Error')
    }
  }

  const remove = async (id: number) => {
    if (!confirm('Delete this template?')) return
    await api.delete(`/templates/${id}`)
    toast.success('Template deleted')
    load()
  }

  // Preview with sample data
  const preview = (body: string) =>
    body
      .replace(/\{name\}/g, 'Juan Dela Cruz')
      .replace(/\{phone\}/g, '09171234567')
      .replace(/\{email\}/g, 'juan@email.com')
      .replace(/\{code\}/g, '123456')

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold">Templates</h2>
        <button onClick={openCreate} className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
          + Add Template
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {templates.map((t) => (
          <div key={t.id} className="bg-white rounded-xl p-5 shadow">
            <h3 className="font-bold">{t.name}</h3>
            <p className="text-gray-600 mt-2 text-sm whitespace-pre-wrap">{t.body}</p>
            <div className="mt-2 text-xs text-gray-400">
              Preview: <span className="text-gray-600">{preview(t.body)}</span>
            </div>
            <div className="flex gap-2 mt-3">
              <button onClick={() => openEdit(t)} className="text-sm text-blue-600 hover:underline">Edit</button>
              <button onClick={() => remove(t.id)} className="text-sm text-red-600 hover:underline">Delete</button>
            </div>
          </div>
        ))}
        {templates.length === 0 && (
          <p className="text-gray-400 col-span-2 text-center py-8">No templates yet</p>
        )}
      </div>

      {showModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-[520px]">
            <h3 className="text-lg font-bold mb-4">{editing ? 'Edit Template' : 'Add Template'}</h3>
            <div className="space-y-3">
              <input
                placeholder="Template Name *"
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                className="w-full border rounded-lg px-3 py-2"
              />
              <div>
                <textarea
                  placeholder="Message body. Use {name}, {phone}, {email}, {code} as variables"
                  value={form.body}
                  onChange={(e) => setForm({ ...form, body: e.target.value })}
                  className="w-full border rounded-lg px-3 py-2"
                  rows={4}
                />
                <div className="flex justify-between text-xs text-gray-400 mt-1">
                  <span>Variables: {'{name}'}, {'{phone}'}, {'{email}'}, {'{code}'}</span>
                  <span className={form.body.length > 160 ? 'text-orange-500' : ''}>{charInfo(form.body)}</span>
                </div>
              </div>
              {form.body && (
                <div className="bg-gray-50 p-3 rounded-lg">
                  <p className="text-xs text-gray-400 mb-1">Preview:</p>
                  <p className="text-sm">{preview(form.body)}</p>
                </div>
              )}
            </div>
            <div className="flex justify-end gap-2 mt-4">
              <button onClick={() => setShowModal(false)} className="px-4 py-2 bg-gray-200 rounded-lg">Cancel</button>
              <button onClick={save} className="px-4 py-2 bg-blue-600 text-white rounded-lg">Save</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
