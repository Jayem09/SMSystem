import { useState, useEffect, type FormEvent } from 'react';
import api from '../api/axios';
import DataTable from '../components/DataTable';
import Modal from '../components/Modal';
import FormField from '../components/FormField';
import { useAuth } from '../context/AuthContext';

interface Category { id: number; name: string; }
interface Brand { id: number; name: string; }
interface Product {
  id: number;
  name: string;
  description: string;
  price: number;
  stock: number;
  image_url: string;
  category_id: number;
  brand_id: number;
  category?: Category;
  brand?: Brand;
}

export default function Products() {
  const { user } = useAuth();
  const isAdmin = user?.role === 'admin';
  const [products, setProducts] = useState<Product[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [brands, setBrands] = useState<Brand[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Product | null>(null);
  const [error, setError] = useState('');

  // Form state
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [price, setPrice] = useState('');
  const [stock, setStock] = useState('');
  const [imageUrl, setImageUrl] = useState('');
  const [categoryId, setCategoryId] = useState('');
  const [brandId, setBrandId] = useState('');

  const fetchProducts = async () => {
    try {
      const params: any = {};
      if (search) params.search = search;
      const res = await api.get('/api/products', { params });
      setProducts(res.data.products || []);
    } catch {
      setError('Failed to load products');
    } finally {
      setLoading(false);
    }
  };

  const fetchMeta = async () => {
    const [catRes, brandRes] = await Promise.all([
      api.get('/api/categories'),
      api.get('/api/brands'),
    ]);
    setCategories(catRes.data.categories || []);
    setBrands(brandRes.data.brands || []);
  };

  useEffect(() => { fetchProducts(); fetchMeta(); }, []);
  useEffect(() => { const t = setTimeout(fetchProducts, 300); return () => clearTimeout(t); }, [search]);

  const openCreate = () => {
    setEditing(null);
    setName(''); setDescription(''); setPrice(''); setStock('0');
    setImageUrl(''); setCategoryId(''); setBrandId('');
    setError('');
    setModalOpen(true);
  };

  const openEdit = (p: Product) => {
    setEditing(p);
    setName(p.name);
    setDescription(p.description);
    setPrice(String(p.price));
    setStock(String(p.stock));
    setImageUrl(p.image_url);
    setCategoryId(String(p.category_id));
    setBrandId(String(p.brand_id));
    setError('');
    setModalOpen(true);
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    const payload = {
      name,
      description,
      price: parseFloat(price),
      stock: parseInt(stock),
      image_url: imageUrl,
      category_id: parseInt(categoryId),
      brand_id: parseInt(brandId),
    };
    try {
      if (editing) {
        await api.put(`/api/products/${editing.id}`, payload);
      } else {
        await api.post('/api/products', payload);
      }
      setModalOpen(false);
      fetchProducts();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Operation failed');
    }
  };

  const handleDelete = async (p: Product) => {
    if (!confirm(`Delete product "${p.name}"?`)) return;
    try {
      await api.delete(`/api/products/${p.id}`);
      fetchProducts();
    } catch {
      alert('Failed to delete product');
    }
  };

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-xl font-semibold text-gray-900">Products</h1>
        {isAdmin && (
          <button onClick={openCreate} className="px-4 py-2 text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-500 rounded-md cursor-pointer">
            Add Product
          </button>
        )}
      </div>

      {error && !modalOpen && <p className="text-red-600 text-sm mb-4">{error}</p>}

      <DataTable
        columns={[
          { key: 'name', label: 'Name' },
          { key: 'category', label: 'Category', render: (p) => p.category?.name || '--' },
          { key: 'brand', label: 'Brand', render: (p) => p.brand?.name || '--' },
          { key: 'price', label: 'Price', render: (p) => `P ${p.price.toLocaleString()}` },
          { key: 'stock', label: 'Stock', render: (p) => (
            <span className={p.stock <= 5 ? 'text-red-600 font-medium' : ''}>{p.stock}</span>
          )},
        ]}
        data={products}
        loading={loading}
        searchValue={search}
        onSearchChange={setSearch}
        searchPlaceholder="Search products..."
        onEdit={isAdmin ? openEdit : undefined}
        onDelete={isAdmin ? handleDelete : undefined}
      />

      <Modal open={modalOpen} onClose={() => setModalOpen(false)} title={editing ? 'Edit Product' : 'New Product'}>
        <form onSubmit={handleSubmit}>
          {error && <p className="text-red-600 text-sm mb-3">{error}</p>}
          <FormField label="Name" value={name} onChange={setName} required placeholder="Product name" />
          <FormField label="Description" type="textarea" value={description} onChange={setDescription} />
          <div className="grid grid-cols-2 gap-3">
            <FormField label="Price" type="number" value={price} onChange={setPrice} required min={0} step="0.01" />
            <FormField label="Stock" type="number" value={stock} onChange={setStock} required min={0} />
          </div>
          <FormField
            label="Category"
            type="select"
            value={categoryId}
            onChange={setCategoryId}
            required
            options={categories.map((c) => ({ value: c.id, label: c.name }))}
          />
          <FormField
            label="Brand"
            type="select"
            value={brandId}
            onChange={setBrandId}
            required
            options={brands.map((b) => ({ value: b.id, label: b.name }))}
          />
          <FormField label="Image URL" value={imageUrl} onChange={setImageUrl} placeholder="https://..." />
          <button type="submit" className="w-full mt-2 py-2 text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-500 rounded-md cursor-pointer">
            {editing ? 'Update' : 'Create'}
          </button>
        </form>
      </Modal>
    </div>
  );
}
