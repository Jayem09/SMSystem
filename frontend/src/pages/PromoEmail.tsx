import { useState, useEffect, useMemo } from 'react';
import api from '../api/axios';
import { Mail, Send, Users, Truck, CheckCircle, XCircle, Loader2, Calendar, Tag, MessageSquare, Layout, Eye, Settings, Briefcase } from 'lucide-react';

interface Customer {
  id: number;
  name: string;
  email: string;
}

interface Supplier {
  id: number;
  name: string;
  email: string;
}

interface Recipient {
  email: string;
  name: string;
}

const TEMPLATES = [
  { id: 'buy4get1', name: 'BUY X GET Y FREE', defaultDiscount: 'Buy 4 Get 1 Free', accent: '#d97706' },
  { id: 'discount', name: 'DISCOUNT SALE', defaultDiscount: '20% OFF', accent: '#4f46e5' },
  { id: 'seasonal', name: 'SEASONAL PROMO', defaultDiscount: 'Special Offer', accent: '#059669' },
];

export default function PromoEmail() {
  const [customers, setCustomers] = useState<Customer[]>([]);
  const [suppliers, setSuppliers] = useState<Supplier[]>([]);
  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);

  // Form state
  const [campaignTitle, setCampaignTitle] = useState('');
  const [subjectLine, setSubjectLine] = useState('');
  const [selectedTemplate, setSelectedTemplate] = useState(TEMPLATES[0].id);
  const [promoCode, setPromoCode] = useState('');
  const [discount, setDiscount] = useState(TEMPLATES[0].defaultDiscount);
  const [validUntil, setValidUntil] = useState('');
  const [details, setDetails] = useState('');

  // Selection state
  const [recipientType, setRecipientType] = useState<'customers' | 'suppliers'>('customers');
  const [selectedRecipients, setSelectedRecipients] = useState<Recipient[]>([]);
  const [selectAll, setSelectAll] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');

  // Result state
  const [result, setResult] = useState<{ success: number; failed: number; failed_emails: string[] } | null>(null);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      const [customersRes, suppliersRes] = await Promise.all([
        api.get('/api/customers'),
        api.get('/api/suppliers'),
      ]);

      const customersData = ((customersRes.data as any).customers || customersRes.data || []) as Customer[];
      const suppliersData = ((suppliersRes.data as any).suppliers || suppliersRes.data || []) as Supplier[];

      setCustomers(customersData.filter(c => c.email?.trim()));
      setSuppliers(suppliersData.filter(s => s.email?.trim()));
    } catch (err) {
      console.error('Failed to fetch data:', err);
    } finally {
      setLoading(false);
    }
  };

  const currentRecipients = useMemo(() => {
    const list = recipientType === 'customers' ? customers : suppliers;
    if (!searchQuery) return list;
    return list.filter(r => 
      r.name.toLowerCase().includes(searchQuery.toLowerCase()) || 
      r.email.toLowerCase().includes(searchQuery.toLowerCase())
    );
  }, [recipientType, customers, suppliers, searchQuery]);

  const handleSelectAll = () => {
    if (selectAll) {
      setSelectedRecipients([]);
    } else {
      setSelectedRecipients(
        currentRecipients.map((r) => ({
          email: r.email,
          name: r.name,
        }))
      );
    }
    setSelectAll(!selectAll);
  };

  const handleToggleRecipient = (r: Customer | Supplier) => {
    const exists = selectedRecipients.find((rec) => rec.email === r.email);
    if (exists) {
      setSelectedRecipients(selectedRecipients.filter((rec) => rec.email !== r.email));
    } else {
      setSelectedRecipients([
        ...selectedRecipients,
        { email: r.email, name: r.name },
      ]);
    }
  };

  const handleTemplateChange = (templateId: string) => {
    setSelectedTemplate(templateId);
    const template = TEMPLATES.find((t) => t.id === templateId);
    if (template) {
      setDiscount(template.defaultDiscount);
    }
  };

  const handleSend = async () => {
    if (selectedRecipients.length === 0) {
      alert('Please select at least one recipient');
      return;
    }
    if (!promoCode.trim()) {
      alert('Please enter a promo code');
      return;
    }

    setSending(true);
    setResult(null);

    try {
      const res = await api.post('/api/promo/send', {
        recipients: selectedRecipients,
        template: selectedTemplate,
        subject: subjectLine,
        promo_code: promoCode,
        discount,
        valid_until: validUntil,
        details,
      });

      setResult(res.data as any);
    } catch (err: unknown) {
      console.error('Send error:', err);
      const axiosErr = err as { response?: { data?: { error?: string } }; message?: string };
      const msg = axiosErr.response?.data?.error || axiosErr.message || 'Failed to send emails';
      alert(msg);
    } finally {
      setSending(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-screen bg-gray-950">
        <Loader2 className="w-12 h-12 animate-spin text-gray-700" />
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#030712] text-gray-100 p-8 font-sans">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-6 mb-16 px-4">
        <div>
          <h1 className="text-4xl font-black tracking-[0.1em] text-white mb-2" style={{ letterSpacing: '0.15em' }}>PROMO</h1>
          <p className="text-gray-500 font-bold uppercase tracking-[0.3em] text-[10px]">Premium Marketing Suite</p>
        </div>

        {result && (
          <div className={`px-6 py-3 rounded-full border backdrop-blur-xl animate-in fade-in slide-in-from-right-4 ${
            result.failed === 0 
            ? 'bg-emerald-500/10 border-emerald-500/20 text-emerald-400' 
            : 'bg-amber-500/10 border-amber-500/20 text-amber-400'
          }`}>
            <div className="flex items-center gap-3">
              <CheckCircle className="w-4 h-4" />
              <div className="text-[10px] font-black uppercase tracking-widest">
                Sent {result.success} campaigns
              </div>
            </div>
          </div>
        )}
      </div>

      <div className="grid grid-cols-1 xl:grid-cols-12 gap-10 items-start max-w-7xl mx-auto">
        
        {/* Column 1: Configuration (4 cols) */}
        <div className="xl:col-span-4 space-y-8">
          <div className="bg-white/[0.03] backdrop-blur-2xl rounded-2xl p-8 border border-white/5">
            <h3 className="text-[10px] font-black tracking-[0.3em] text-gray-500 uppercase mb-8">Campaign Setup</h3>
            
            <div className="space-y-8">
              <div>
                <label className="block text-[8px] font-black text-gray-600 uppercase tracking-widest mb-3">Internal Identifier</label>
                <input
                  type="text"
                  value={campaignTitle}
                  onChange={(e) => setCampaignTitle(e.target.value)}
                  placeholder="Reference"
                  className="w-full bg-black/40 border-b border-white/10 rounded-none px-0 py-3 focus:border-white outline-none transition-all placeholder:text-gray-700 font-medium text-sm"
                />
              </div>

              <div>
                <label className="block text-[8px] font-black text-gray-600 uppercase tracking-widest mb-3">Subject Line</label>
                <input
                  type="text"
                  value={subjectLine}
                  onChange={(e) => setSubjectLine(e.target.value)}
                  placeholder="Inbox Title"
                  className="w-full bg-black/40 border-b border-white/10 rounded-none px-0 py-3 focus:border-white outline-none transition-all placeholder:text-gray-700 font-medium text-sm"
                />
              </div>

              <div>
                <label className="block text-[8px] font-black text-gray-600 uppercase tracking-widest mb-4">Template Selection</label>
                <div className="space-y-2">
                  {TEMPLATES.map((t) => (
                    <button
                      key={t.id}
                      onClick={() => handleTemplateChange(t.id)}
                      className={`w-full flex items-center justify-between p-4 rounded-xl border transition-all ${
                        selectedTemplate === t.id 
                        ? 'bg-white/5 border-white/20' 
                        : 'bg-transparent border-white/5 opacity-40 hover:opacity-80'
                      }`}
                    >
                      <span className="text-[10px] font-bold tracking-widest">{t.name}</span>
                      <div className="w-1.5 h-1.5 rounded-full" style={{ backgroundColor: t.accent }}></div>
                    </button>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Column 2: Content (4 cols) */}
        <div className="xl:col-span-4 space-y-8">
          <div className="bg-white/[0.03] backdrop-blur-2xl rounded-2xl p-8 border border-white/5">
            <h3 className="text-[10px] font-black tracking-[0.3em] text-gray-500 uppercase mb-8">Offer Details</h3>

            <div className="space-y-8">
              <div className="grid grid-cols-2 gap-6">
                <div>
                  <label className="block text-[8px] font-black text-gray-600 uppercase tracking-widest mb-3">Code</label>
                  <input
                    type="text"
                    value={promoCode}
                    onChange={(e) => setPromoCode(e.target.value.toUpperCase())}
                    placeholder="CODE"
                    className="w-full bg-black/40 border-b border-white/10 rounded-none px-0 py-3 focus:border-white outline-none transition-all font-mono font-bold text-center tracking-[0.2em] text-sm"
                  />
                </div>
                <div>
                  <label className="block text-[8px] font-black text-gray-600 uppercase tracking-widest mb-3">Value</label>
                  <input
                    type="text"
                    value={discount}
                    onChange={(e) => setDiscount(e.target.value)}
                    placeholder="20% OFF"
                    className="w-full bg-black/40 border-b border-white/10 rounded-none px-0 py-3 focus:border-white outline-none transition-all font-bold text-sm"
                  />
                </div>
              </div>

              <div>
                <label className="block text-[8px] font-black text-gray-600 uppercase tracking-widest mb-3">Expiration</label>
                <input
                  type="date"
                  value={validUntil}
                  onChange={(e) => setValidUntil(e.target.value)}
                  className="w-full bg-black/40 border-b border-white/10 rounded-none px-0 py-3 focus:border-white outline-none transition-all text-sm font-medium"
                />
              </div>

              <div>
                <label className="block text-[8px] font-black text-gray-600 uppercase tracking-widest mb-3">More Information</label>
                <textarea
                  value={details}
                  onChange={(e) => setDetails(e.target.value)}
                  placeholder="Terms or personal notes..."
                  rows={4}
                  className="w-full bg-black/20 border border-white/5 rounded-xl px-4 py-4 focus:border-white/20 outline-none transition-all resize-none text-xs leading-relaxed"
                />
              </div>
            </div>
          </div>
          
          <div className="px-8 flex items-center gap-2 opacity-30">
             <Eye className="w-3 h-3" />
             <span className="text-[8px] font-black tracking-widest uppercase">Live Preview Active</span>
          </div>
        </div>

        {/* Column 3: Audience (4 cols) */}
        <div className="xl:col-span-4 space-y-8">
          <div className="bg-white/[0.03] backdrop-blur-2xl rounded-2xl p-8 border border-white/5 flex flex-col h-[540px]">
            <div className="flex items-center justify-between mb-8">
              <h3 className="text-[10px] font-black tracking-[0.3em] text-gray-500 uppercase">Audience</h3>
              <span className="text-[8px] font-black tracking-widest text-white px-2 py-0.5 bg-white/10 rounded-full">{selectedRecipients.length}</span>
            </div>

            <div className="flex p-1 bg-black/40 rounded-xl mb-6 ring-1 ring-white/5">
              <button
                onClick={() => { setRecipientType('customers'); setSelectAll(false); setSelectedRecipients([]); }}
                className={`flex-1 flex items-center justify-center py-2.5 rounded-lg text-[9px] font-black tracking-widest transition-all ${recipientType === 'customers' ? 'bg-white/10 text-white' : 'text-gray-600 hover:text-gray-400'}`}
              >
                CUSTOMERS
              </button>
              <button
                onClick={() => { setRecipientType('suppliers'); setSelectAll(false); setSelectedRecipients([]); }}
                className={`flex-1 flex items-center justify-center py-2.5 rounded-lg text-[9px] font-black tracking-widest transition-all ${recipientType === 'suppliers' ? 'bg-white/10 text-white' : 'text-gray-600 hover:text-gray-400'}`}
              >
                SUPPLIERS
              </button>
            </div>

            <div className="space-y-4 mb-6">
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Search..."
                className="w-full bg-transparent border-b border-white/10 px-0 py-2 focus:border-white outline-none text-[10px] placeholder:text-gray-700"
              />
              <button 
                onClick={handleSelectAll}
                className="w-full text-[8px] font-black tracking-[0.2em] uppercase text-gray-600 hover:text-white transition-colors"
              >
                {selectAll ? 'Deselect all' : `Select all (${currentRecipients.length})`}
              </button>
            </div>

            <div className="flex-1 overflow-y-auto space-y-2 pr-2 custom-scrollbar">
              {currentRecipients.map((r) => {
                const isSelected = selectedRecipients.some((rec) => rec.email === r.email);
                return (
                  <div
                    key={r.id}
                    onClick={() => handleToggleRecipient(r)}
                    className={`group flex items-center gap-4 p-3 rounded-xl border cursor-pointer transition-all ${isSelected ? 'bg-white/5 border-white/10' : 'bg-transparent border-transparent hover:bg-white/5'}`}
                  >
                    <div className={`w-3.5 h-3.5 rounded-sm border transition-all ${isSelected ? 'bg-white border-white' : 'border-white/10'}`}>
                      {isSelected && <CheckCircle className="w-2.5 h-2.5 text-black mx-auto mt-0.5" />}
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-[10px] font-bold text-white truncate uppercase tracking-tight">{r.name}</p>
                      <p className="text-[8px] text-gray-600 truncate">{r.email}</p>
                    </div>
                  </div>
                );
              })}
            </div>

            <div className="mt-8 pt-4 border-t border-white/5">
              <button
                onClick={handleSend}
                disabled={sending || selectedRecipients.length === 0 || !promoCode.trim()}
                className={`w-full flex items-center justify-center gap-4 py-5 rounded-2xl font-black text-[10px] tracking-[0.3em] transition-all ${
                  sending || selectedRecipients.length === 0 || !promoCode.trim()
                  ? 'bg-gray-900 text-gray-700 cursor-not-allowed'
                  : 'bg-white text-black hover:opacity-90 active:scale-95'
                }`}
              >
                {sending ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  'APPROVE CAMPAIGN'
                )}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}