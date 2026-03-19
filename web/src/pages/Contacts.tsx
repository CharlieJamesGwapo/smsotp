import { useState, useEffect, useRef } from 'react'
import api from '../api/client'
import toast from 'react-hot-toast'

interface Contact {
  id: number
  name: string
  phone: string
  email: string
  notes: string
}

export default function Contacts() {
  const [contacts, setContacts] = useState<Contact[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [showModal, setShowModal] = useState(false)
  const [editing, setEditing] = useState<Contact | null>(null)
  const [form, setForm] = useState({ name: '', phone: '', email: '', notes: '' })
  const fileRef = useRef<HTMLInputElement>(null)

  const load = () => {
    api.get(`/contacts?page=${page}&per_page=20&search=${search}`).then((r) => {
      setContacts(r.data.data || [])
      setTotal(r.data.total)
    })
  }

  useEffect(() => { load() }, [page, search])

  const openCreate = () => {
    setEditing(null)
    setForm({ name: '', phone: '', email: '', notes: '' })
    setShowModal(true)
  }

  const openEdit = (c: Contact) => {
    setEditing(c)
    setForm({ name: c.name, phone: c.phone, email: c.email || '', notes: c.notes || '' })
    setShowModal(true)
  }

  const save = async () => {
    try {
      if (editing) {
        await api.put(`/contacts/${editing.id}`, form)
        toast.success('Contact updated')
      } else {
        await api.post('/contacts', form)
        toast.success('Contact created')
      }
      setShowModal(false)
      load()
    } catch (e: any) {
      toast.error(e.response?.data?.error || 'Error saving contact')
    }
  }

  const remove = async (id: number) => {
    if (!confirm('Delete this contact?')) return
    await api.delete(`/contacts/${id}`)
    toast.success('Contact deleted')
    load()
  }

  const importCSV = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    const fd = new FormData()
    fd.append('file', file)
    try {
      const res = await api.post('/contacts/import', fd, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      toast.success(`Imported ${res.data.imported}, skipped ${res.data.skipped}`)
      load()
    } catch {
      toast.error('Import failed')
    }
    if (fileRef.current) fileRef.current.value = ''
  }

  const totalPages = Math.ceil(total / 20)

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold">Contacts ({total})</h2>
        <div className="flex gap-2">
          <input
            ref={fileRef}
            type="file"
            accept=".csv"
            onChange={importCSV}
            className="hidden"
          />
          <button
            onClick={() => fileRef.current?.click()}
            className="px-4 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700"
          >
            Import CSV
          </button>
          <button
            onClick={openCreate}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            + Add Contact
          </button>
        </div>
      </div>

      <input
        type="text"
        placeholder="Search contacts..."
        value={search}
        onChange={(e) => { setSearch(e.target.value); setPage(1) }}
        className="w-full border rounded-lg px-3 py-2 mb-4 focus:outline-none focus:ring-2 focus:ring-blue-500"
      />

      <div className="bg-white rounded-xl shadow overflow-hidden">
        <table className="w-full">
          <thead className="bg-gray-50">
            <tr>
              <th className="text-left px-4 py-3 text-sm font-medium text-gray-500">Name</th>
              <th className="text-left px-4 py-3 text-sm font-medium text-gray-500">Phone</th>
              <th className="text-left px-4 py-3 text-sm font-medium text-gray-500">Email</th>
              <th className="text-left px-4 py-3 text-sm font-medium text-gray-500">Notes</th>
              <th className="text-right px-4 py-3 text-sm font-medium text-gray-500">Actions</th>
            </tr>
          </thead>
          <tbody>
            {contacts.map((c) => (
              <tr key={c.id} className="border-t hover:bg-gray-50">
                <td className="px-4 py-3 font-medium">{c.name}</td>
                <td className="px-4 py-3 text-gray-600">{c.phone}</td>
                <td className="px-4 py-3 text-gray-600">{c.email || '-'}</td>
                <td className="px-4 py-3 text-gray-600 truncate max-w-48">{c.notes || '-'}</td>
                <td className="px-4 py-3 text-right space-x-2">
                  <button onClick={() => openEdit(c)} className="text-blue-600 hover:underline text-sm">Edit</button>
                  <button onClick={() => remove(c.id)} className="text-red-600 hover:underline text-sm">Delete</button>
                </td>
              </tr>
            ))}
            {contacts.length === 0 && (
              <tr><td colSpan={5} className="text-center py-8 text-gray-400">No contacts found</td></tr>
            )}
          </tbody>
        </table>
      </div>

      {totalPages > 1 && (
        <div className="flex justify-center gap-2 mt-4">
          {Array.from({ length: totalPages }, (_, i) => (
            <button
              key={i}
              onClick={() => setPage(i + 1)}
              className={`px-3 py-1 rounded ${page === i + 1 ? 'bg-blue-600 text-white' : 'bg-gray-200'}`}
            >
              {i + 1}
            </button>
          ))}
        </div>
      )}

      {showModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-[480px]">
            <h3 className="text-lg font-bold mb-4">{editing ? 'Edit Contact' : 'Add Contact'}</h3>
            <div className="space-y-3">
              <input
                placeholder="Name *"
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                className="w-full border rounded-lg px-3 py-2"
              />
              <input
                placeholder="Phone * (e.g., 09171234567)"
                value={form.phone}
                onChange={(e) => setForm({ ...form, phone: e.target.value })}
                className="w-full border rounded-lg px-3 py-2"
              />
              <input
                placeholder="Email"
                value={form.email}
                onChange={(e) => setForm({ ...form, email: e.target.value })}
                className="w-full border rounded-lg px-3 py-2"
              />
              <textarea
                placeholder="Notes"
                value={form.notes}
                onChange={(e) => setForm({ ...form, notes: e.target.value })}
                className="w-full border rounded-lg px-3 py-2"
                rows={2}
              />
            </div>
            <div className="flex justify-end gap-2 mt-4">
              <button onClick={() => setShowModal(false)} className="px-4 py-2 bg-gray-200 rounded-lg">Cancel</button>
              <button onClick={save} className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">Save</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
