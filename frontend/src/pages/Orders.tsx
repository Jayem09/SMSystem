import { useState, useEffect } from 'react';
import api from '../api/axios';
import DataTable from '../components/DataTable';
import Modal from '../components/Modal';
import { useAuth } from '../hooks/useAuth';
import { Printer, Eye, Trash2 } from 'lucide-react';

interface Customer { id: number; name: string; }
interface Product { id: number; name: string; price: number; stock: number; }
interface OrderItem {
  id: number;
  product_id: number;
  quantity: number;
  unit_price: number;
  subtotal: number;
  product?: Product;
}
interface Order {
  id: number;
  customer_id?: number | null;
  guest_name?: string;
  guest_phone?: string;
  total_amount: number;
  status: string;
  payment_method: string;
  created_at: string;
  customer?: Customer;
  items?: OrderItem[];
}

const statusColors: Record<string, string> = {
  pending: 'bg-yellow-100 text-yellow-800',
  confirmed: 'bg-blue-100 text-blue-800',
  completed: 'bg-green-100 text-green-800',
  cancelled: 'bg-red-100 text-red-800',
};

export default function Orders() {
  const { user } = useAuth();
  const isAdmin = user?.role === 'admin';
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [invoiceModalOpen, setInvoiceModalOpen] = useState(false);
  const [itemsModalOpen, setItemsModalOpen] = useState(false);
  const [selectedOrder, setSelectedOrder] = useState<Order | null>(null);
  const [error, setError] = useState('');



  const fetchOrders = async () => {
    try {
      const res = await api.get('/api/orders');
      setOrders(res.data.orders);
    } catch (err: unknown) {
      console.error('Failed to fetch orders', err);
      setError('Failed to load orders');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchOrders(); }, []);



  const handleDelete = async (order: Order) => {
    if (!confirm(`Delete order #${order.id}?`)) return;
    try {
      await api.delete(`/api/orders/${order.id}`);
      fetchOrders();
    } catch {
      alert('Failed to delete order');
    }
  };

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 tracking-tight">Orders</h1>
          <p className="text-gray-500 mt-1">Manage sales, payments, and generate customer receipts.</p>
        </div>
      </div>

      {error && <p className="text-red-600 text-sm mb-4">{error}</p>}

      <DataTable
        columns={[
          { key: 'id', label: 'Order #', render: (o) => `#${o.id}` },
          { key: 'customer', label: 'Customer', render: (o) => o.customer?.name || o.guest_name || 'Walk-In' },
          { key: 'total_amount', label: 'Total', render: (o) => `P ${o.total_amount.toLocaleString()}` },
          { key: 'status', label: 'Status', render: (o) => (
            <span className={`px-2 py-0.5 rounded text-xs font-medium ${statusColors[o.status] || 'bg-gray-100 text-gray-800'}`}>
              {o.status}
            </span>
          )},
          { key: 'payment_method', label: 'Payment' },
          { key: 'created_at', label: 'Date', render: (o) => new Date(o.created_at).toLocaleDateString() },
        ]}
        data={orders}
        loading={loading}
        actions={(order) => (
          <div className="flex items-center gap-3 justify-end">
            <button
              onClick={() => { setSelectedOrder(order); setItemsModalOpen(true); }}
              className="p-2 text-gray-400 hover:text-gray-900 hover:bg-gray-100 rounded-xl transition-colors cursor-pointer"
              title="View Details"
            >
              <Eye className="w-4 h-4" />
            </button>
            <button
              onClick={() => { setSelectedOrder(order); setInvoiceModalOpen(true); }}
              className="p-2 text-indigo-500 hover:text-indigo-700 hover:bg-indigo-50 rounded-xl transition-colors cursor-pointer"
              title="Print Invoice"
            >
              <Printer className="w-4 h-4" />
            </button>

            {isAdmin && (
              <button 
                onClick={() => handleDelete(order)} 
                className="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-xl transition-colors cursor-pointer"
                title="Delete Order"
              >
                <Trash2 className="w-4 h-4" />
              </button>
            )}
          </div>
        )}
      />

      {/* Items Modal */}
      <Modal open={itemsModalOpen} onClose={() => setItemsModalOpen(false)} title={selectedOrder ? `Order #${selectedOrder.id} - Items` : 'Order Items'}>
        {selectedOrder && (
          <div className="max-h-[60vh] overflow-y-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-xs text-gray-500 uppercase">
                  <th className="pb-3 border-b border-gray-100">Product</th>
                  <th className="pb-3 border-b border-gray-100 text-center">Qty</th>
                  <th className="pb-3 border-b border-gray-100 text-right">Unit Price</th>
                  <th className="pb-3 border-b border-gray-100 text-right">Subtotal</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {selectedOrder.items?.map((item) => (
                  <tr key={item.id} className="hover:bg-gray-50">
                    <td className="py-3 text-gray-900 font-medium">{item.product?.name || `Product #${item.product_id}`}</td>
                    <td className="py-3 text-gray-600 text-center">{item.quantity}</td>
                    <td className="py-3 text-gray-600 text-right">₱{item.unit_price.toLocaleString()}</td>
                    <td className="py-3 text-gray-900 font-bold text-right">₱{item.subtotal.toLocaleString()}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Modal>



      {/* Invoice Modal */}
      <Modal open={invoiceModalOpen} onClose={() => setInvoiceModalOpen(false)} title="Print Preview">
        {selectedOrder && (
          <div className="p-4 overflow-y-auto max-h-[80vh]">
            <div id="printable-invoice" className="bg-white p-8 border border-gray-100 rounded-xl shadow-sm font-sans text-gray-900">
              <div className="flex justify-between items-start mb-8 pb-8 border-b border-gray-100">
                <div>
                  <h1 className="text-3xl font-black text-gray-900 tracking-tighter mb-1">SM TYRE DEPOT</h1>
                  <p className="text-xs font-bold text-gray-500 uppercase tracking-widest">Premium Auto Care & Tyres</p>
                </div>
                <div className="text-right">
                  <h2 className="text-xl font-bold text-gray-900 mb-1">INVOICE</h2>
                  <p className="text-sm font-medium text-gray-500">#{selectedOrder.id}</p>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-8 mb-12">
                <div>
                  <h3 className="text-xs font-bold text-gray-400 uppercase tracking-widest mb-2">Billed To</h3>
                  <p className="font-bold text-gray-900">{selectedOrder.customer?.name || selectedOrder.guest_name || 'Walk-In Customer'}</p>
                  <p className="text-sm text-gray-600">
                    {selectedOrder.customer_id ? `Customer ID: ${selectedOrder.customer_id}` : (selectedOrder.guest_phone ? `Contact: ${selectedOrder.guest_phone}` : 'No Contact Info')}
                  </p>
                </div>
                <div className="text-right">
                  <h3 className="text-xs font-bold text-gray-400 uppercase tracking-widest mb-2">Issued On</h3>
                  <p className="font-bold text-gray-900">{new Date(selectedOrder.created_at).toLocaleDateString()}</p>
                  <p className="text-sm text-gray-600">Payment: {selectedOrder.payment_method.toUpperCase()}</p>
                </div>
              </div>

              <table className="w-full mb-8">
                <thead>
                  <tr className="text-left border-b-2 border-gray-900">
                    <th className="py-4 text-xs font-black uppercase tracking-widest">Description</th>
                    <th className="py-4 text-xs font-black uppercase tracking-widest text-center">Qty</th>
                    <th className="py-4 text-xs font-black uppercase tracking-widest text-right">Price</th>
                    <th className="py-4 text-xs font-black uppercase tracking-widest text-right">Amount</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100">
                  {selectedOrder.items?.map((item) => (
                    <tr key={item.id}>
                      <td className="py-4 font-bold text-gray-900">{item.product?.name}</td>
                      <td className="py-4 text-center font-medium">{item.quantity}</td>
                      <td className="py-4 text-right text-gray-600">₱{item.unit_price.toLocaleString()}</td>
                      <td className="py-4 text-right font-bold">₱{item.subtotal.toLocaleString()}</td>
                    </tr>
                  ))}
                </tbody>
              </table>

              <div className="flex justify-end pt-8 border-t-2 border-gray-900">
                <div className="w-64 space-y-3">
                  <div className="flex justify-between text-sm text-gray-500 font-medium">
                    <span>Subtotal</span>
                    <span>₱{selectedOrder.items?.reduce((a,c) => a + c.subtotal, 0).toLocaleString()}</span>
                  </div>
                  <div className="flex justify-between text-lg font-black text-gray-900 pt-3 border-t border-gray-100">
                    <span className="tracking-tighter uppercase">Total Amount</span>
                    <span>₱{selectedOrder.total_amount.toLocaleString()}</span>
                  </div>
                </div>
              </div>

              <div className="mt-20 pt-8 border-t border-gray-100 text-center">
                <p className="text-xs font-bold text-gray-400 uppercase tracking-widest leading-relaxed">
                  Thank you for your business!<br/>
                  Please keep this receipt for your records.
                </p>
              </div>
            </div>

            <style dangerouslySetInnerHTML={{ __html: `
              @media print {
                body * { visibility: hidden; }
                #printable-invoice, #printable-invoice * { visibility: visible; }
                #printable-invoice { 
                  position: absolute; 
                  left: 0; 
                  top: 0; 
                  width: 100%; 
                  border: none !important;
                  box-shadow: none !important;
                }
              }
            `}} />

            <button 
              onClick={() => window.print()} 
              className="mt-6 w-full py-4 text-sm font-black text-white bg-gray-900 hover:bg-gray-800 rounded-2xl transition-all shadow-lg hover:shadow-xl active:scale-95 flex items-center justify-center gap-2"
            >
              <Printer className="w-4 h-4" />
              PRINT RECEIPT
            </button>
          </div>
        )}
      </Modal>
    </div>
  );
}
