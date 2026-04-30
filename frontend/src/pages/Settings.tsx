import { useEffect, useState } from 'react';
import { Settings as SettingsIcon } from 'lucide-react';
import api from '../api/axios';

export default function Settings() {
  const developerEmail = 'johndinglasan12@gmail.com';
  const [loading, setLoading] = useState(false);
  const [storeName, setStoreName] = useState('SMSystem');

  useEffect(() => {
    const fetchSettings = async () => {
      try {
        setLoading(true);
        const res = await api.get('/api/settings') as { data: { store_name?: string } };
        if (res.data?.store_name) {
          setStoreName(res.data.store_name);
        }
      } catch (err) {
        console.error('Failed to load settings', err);
      } finally {
        setLoading(false);
      }
    };

    fetchSettings();
  }, []);

  if (loading) {
    return (
      <div className="p-6 h-full flex items-center justify-center">
        <div className="text-gray-400">Loading Configuration...</div>
      </div>
    );
  }

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <div className="flex items-center gap-3 mb-8">
        <div className="p-3 bg-gray-900 text-white rounded-xl">
          <SettingsIcon className="w-6 h-6" />
        </div>
        <div>
          <h1 className="text-2xl font-bold text-gray-900 tracking-tight">System Settings</h1>
          <p className="text-sm text-gray-500">Manage application configuration and preferences.</p>
        </div>
      </div>

      <div className="bg-white rounded-2xl border border-gray-200 shadow-sm overflow-hidden">
        <div className="border-b border-gray-100 p-6">
          <h2 className="text-lg font-bold text-gray-900">General Configuration</h2>
          <p className="text-sm text-gray-500 mt-1">Basic settings for the application interface and behavior.</p>
        </div>

        <div className="p-6">
          <div className="space-y-4 max-w-md">
            <div>
              <label className="block text-sm font-bold text-gray-900 mb-2">Store Name</label>
              <input
                type="text"
                value={storeName}
                readOnly
                className="w-full px-4 py-2 bg-gray-100 border border-gray-200 rounded-xl text-sm text-gray-500 cursor-not-allowed outline-none"
              />
            </div>

            <div>
              <label className="block text-sm font-bold text-gray-900 mb-2">Developer Email</label>
              <input
                type="email"
                value={developerEmail}
                readOnly
                className="w-full px-4 py-2 bg-gray-100 border border-gray-200 rounded-xl text-sm text-gray-500 cursor-not-allowed outline-none"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
