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
    <div className="min-h-screen w-full bg-surface">
      {/* Top Nav */}
      <nav className="w-full border-b border-edge px-6 py-3 flex items-center justify-between bg-surface">
        <span className="text-lg font-semibold text-ink">SMSystem</span>
        <div className="flex items-center gap-4">
          <div className="text-right">
            <p className="text-sm font-medium text-ink">{user?.name}</p>
            <p className="text-xs text-ink-muted capitalize">{user?.role}</p>
          </div>
          <button
            onClick={handleLogout}
            className="px-3 py-1.5 text-sm border border-edge rounded-md text-ink-secondary hover:bg-surface-hover transition-colors cursor-pointer"
          >
            Logout
          </button>
        </div>
      </nav>

      {/* Main */}
      <main className="w-full px-6 py-8 max-w-6xl mx-auto">
        <h1 className="text-2xl font-semibold text-ink mb-1">Dashboard</h1>
        <p className="text-sm text-ink-secondary mb-8">Welcome back, {user?.name}.</p>

        {/* Stats */}
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
          {[
            { label: 'Users', value: '--' },
            { label: 'Products', value: '--' },
            { label: 'Customers', value: '--' },
            { label: 'Orders', value: '--' },
          ].map((stat) => (
            <div key={stat.label} className="border border-edge rounded-lg p-4 bg-surface">
              <p className="text-xs text-ink-muted uppercase tracking-wide mb-1">{stat.label}</p>
              <p className="text-2xl font-semibold text-ink">{stat.value}</p>
            </div>
          ))}
        </div>

        {/* Profile */}
        <div className="border border-edge rounded-lg bg-surface">
          <div className="px-4 py-3 border-b border-edge">
            <h2 className="text-sm font-semibold text-ink">Your Profile</h2>
          </div>
          <div className="divide-y divide-edge">
            {[
              { label: 'Name', value: user?.name },
              { label: 'Email', value: user?.email },
              { label: 'Role', value: user?.role },
              { label: 'Joined', value: user?.created_at ? new Date(user.created_at).toLocaleDateString() : '--' },
            ].map((row) => (
              <div key={row.label} className="px-4 py-3 flex justify-between items-center">
                <span className="text-sm text-ink-secondary">{row.label}</span>
                <span className="text-sm font-medium text-ink capitalize">{row.value}</span>
              </div>
            ))}
          </div>
        </div>
      </main>
    </div>
  );
}
