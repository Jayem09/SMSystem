import { useState, useEffect, type FormEvent } from 'react';
import api from '../api/axios';
import DataTable from '../components/DataTable';
import Modal from '../components/Modal';
import FormField from '../components/FormField';
import { useAuth } from '../context/AuthContext';

interface Brand {
  id: number;
  name: string;
  logo_url: string;
}

export default function Brands() {
  const { user } = useAuth();
  const isAdmin = user?.role === 'admin';
  const [brands, setBrands] = useState<Brand[]>([]);
  const [loading, setLoading] = useState(true);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Brand | null>(null);
  const [name, setName] = useState('');
  const [logoUrl, setLogoUrl] = useState('');
  const [error, setError] = useState('');

  const fetchBrands = async () => {
    try {
      const res = await api.get('/api/brands');
      setBrands(res.data.brands || []);
    } catch {
      setError('Failed to load brands');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchBrands(); }, []);

  const openCreate = () => {
    setEditing(null);
    setName('');
    setLogoUrl('');
    setError('');
    setModalOpen(true);
  };

  const openEdit = (brand: Brand) => {
    setEditing(brand);
    setName(brand.name);
    setLogoUrl(brand.logo_url);
    setError('');
    setModalOpen(true);
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    try {
      if (editing) {
        await api.put(`/api/brands/${editing.id}`, { name, logo_url: logoUrl });
      } else {
        await api.post('/api/brands', { name, logo_url: logoUrl });
      }
      setModalOpen(false);
      fetchBrands();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Operation failed');
    }
  };

  const handleDelete = async (brand: Brand) => {
    if (!confirm(`Delete brand "${brand.name}"?`)) return;
    try {
      await api.delete(`/api/brands/${brand.id}`);
      fetchBrands();
    } catch {
      alert('Failed to delete brand');
    }
  };

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-xl font-semibold text-gray-900">Brands</h1>
        {isAdmin && (
          <button onClick={openCreate} className="px-4 py-2 text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-500 rounded-md cursor-pointer">
            Add Brand
          </button>
        )}
      </div>

      {error && !modalOpen && <p className="text-red-600 text-sm mb-4">{error}</p>}

      <DataTable
        columns={[
          { key: 'name', label: 'Name' },
          { key: 'logo_url', label: 'Logo URL', render: (b) => b.logo_url ? <span className="text-gray-400 text-xs truncate max-w-48 inline-block">{b.logo_url}</span> : <span className="text-gray-300">--</span> },
        ]}
        data={brands}
        loading={loading}
        onEdit={isAdmin ? openEdit : undefined}
        onDelete={isAdmin ? handleDelete : undefined}
      />

      <Modal open={modalOpen} onClose={() => setModalOpen(false)} title={editing ? 'Edit Brand' : 'New Brand'}>
        <form onSubmit={handleSubmit}>
          {error && <p className="text-red-600 text-sm mb-3">{error}</p>}
          <FormField label="Name" value={name} onChange={setName} required placeholder="Brand name" />
          <FormField label="Logo URL" value={logoUrl} onChange={setLogoUrl} placeholder="https://..." />
          <button type="submit" className="w-full mt-2 py-2 text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-500 rounded-md cursor-pointer">
            {editing ? 'Update' : 'Create'}
          </button>
        </form>
      </Modal>
    </div>
  );
}
