import { useState, useEffect, type FormEvent } from 'react';
import api from '../api/axios';
import DataTable from '../components/DataTable';
import Modal from '../components/Modal';
import FormField from '../components/FormField';
import { useAuth } from '../context/AuthContext';

interface Customer {
  id: number;
  name: string;
  email: string;
  phone: string;
  address: string;
}

export default function Customers() {
  const { user } = useAuth();
  const isAdmin = user?.role === 'admin';
  const [customers, setCustomers] = useState<Customer[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Customer | null>(null);
  const [error, setError] = useState('');

  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [phone, setPhone] = useState('');
  const [address, setAddress] = useState('');

  const fetchCustomers = async () => {
    try {
      const params: any = {};
      if (search) params.search = search;
      const res = await api.get('/api/customers', { params });
      setCustomers(res.data.customers || []);
    } catch {
      setError('Failed to load customers');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchCustomers(); }, []);
  useEffect(() => { const t = setTimeout(fetchCustomers, 300); return () => clearTimeout(t); }, [search]);

  const openCreate = () => {
    setEditing(null);
    setName(''); setEmail(''); setPhone(''); setAddress('');
    setError('');
    setModalOpen(true);
  };

  const openEdit = (c: Customer) => {
    setEditing(c);
    setName(c.name); setEmail(c.email); setPhone(c.phone); setAddress(c.address);
    setError('');
    setModalOpen(true);
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    const payload = { name, email, phone, address };
    try {
      if (editing) {
        await api.put(`/api/customers/${editing.id}`, payload);
      } else {
        await api.post('/api/customers', payload);
      }
      setModalOpen(false);
      fetchCustomers();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Operation failed');
    }
  };

  const handleDelete = async (c: Customer) => {
    if (!confirm(`Delete customer "${c.name}"?`)) return;
    try {
      await api.delete(`/api/customers/${c.id}`);
      fetchCustomers();
    } catch {
      alert('Failed to delete customer');
    }
  };

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-xl font-semibold text-gray-900">Customers</h1>
        <button onClick={openCreate} className="px-4 py-2 text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-500 rounded-md cursor-pointer">
          Add Customer
        </button>
      </div>

      {error && !modalOpen && <p className="text-red-600 text-sm mb-4">{error}</p>}

      <DataTable
        columns={[
          { key: 'name', label: 'Name' },
          { key: 'email', label: 'Email', render: (c) => c.email || '--' },
          { key: 'phone', label: 'Phone', render: (c) => c.phone || '--' },
        ]}
        data={customers}
        loading={loading}
        searchValue={search}
        onSearchChange={setSearch}
        searchPlaceholder="Search customers..."
        onEdit={openEdit}
        onDelete={isAdmin ? handleDelete : undefined}
      />

      <Modal open={modalOpen} onClose={() => setModalOpen(false)} title={editing ? 'Edit Customer' : 'New Customer'}>
        <form onSubmit={handleSubmit}>
          {error && <p className="text-red-600 text-sm mb-3">{error}</p>}
          <FormField label="Name" value={name} onChange={setName} required placeholder="Full name" />
          <FormField label="Email" type="email" value={email} onChange={setEmail} placeholder="email@example.com" />
          <FormField label="Phone" value={phone} onChange={setPhone} placeholder="09XX XXX XXXX" />
          <FormField label="Address" type="textarea" value={address} onChange={setAddress} placeholder="Address" />
          <button type="submit" className="w-full mt-2 py-2 text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-500 rounded-md cursor-pointer">
            {editing ? 'Update' : 'Create'}
          </button>
        </form>
      </Modal>
    </div>
  );
}
