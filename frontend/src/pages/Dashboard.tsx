import { useAuth } from '../context/AuthContext';
import { useNavigate } from 'react-router-dom';

export default function Dashboard() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="min-h-screen w-full bg-white">
      {/* Top Nav */}
      <nav className="w-full border-b border-gray-200 px-6 py-3 flex items-center justify-between bg-white">
        <span className="text-lg font-semibold text-gray-900">SMSystem</span>
        <div className="flex items-center gap-4">
          <div className="text-right">
            <p className="text-sm font-medium text-gray-900">{user?.name}</p>
            <p className="text-xs text-gray-400 capitalize">{user?.role}</p>
          </div>
          <button
            onClick={handleLogout}
            className="px-3 py-1.5 text-sm border border-gray-200 rounded-md text-gray-500 hover:bg-gray-50 transition-colors cursor-pointer"
          >
            Logout
          </button>
        </div>
      </nav>

      {/* Main */}
      <main className="w-full px-6 py-8 max-w-6xl mx-auto">
        <h1 className="text-2xl font-semibold text-gray-900 mb-1">Dashboard</h1>
        <p className="text-sm text-gray-500 mb-8">Welcome back, {user?.name}.</p>

        {/* Stats */}
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
          {[
            { label: 'Users', value: '--' },
            { label: 'Products', value: '--' },
            { label: 'Customers', value: '--' },
            { label: 'Orders', value: '--' },
          ].map((stat) => (
            <div key={stat.label} className="border border-gray-200 rounded-lg p-4 bg-white">
              <p className="text-xs text-gray-400 uppercase tracking-wide mb-1">{stat.label}</p>
              <p className="text-2xl font-semibold text-gray-900">{stat.value}</p>
            </div>
          ))}
        </div>

        {/* Profile */}
        <div className="border border-gray-200 rounded-lg bg-white">
          <div className="px-4 py-3 border-b border-gray-200">
            <h2 className="text-sm font-semibold text-gray-900">Your Profile</h2>
          </div>
          <div className="divide-y divide-gray-200">
            {[
              { label: 'Name', value: user?.name },
              { label: 'Email', value: user?.email },
              { label: 'Role', value: user?.role },
              { label: 'Joined', value: user?.created_at ? new Date(user.created_at).toLocaleDateString() : '--' },
            ].map((row) => (
              <div key={row.label} className="px-4 py-3 flex justify-between items-center">
                <span className="text-sm text-gray-500">{row.label}</span>
                <span className="text-sm font-medium text-gray-900 capitalize">{row.value}</span>
              </div>
            ))}
          </div>
        </div>
      </main>
    </div>
  );
}
