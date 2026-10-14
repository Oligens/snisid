import React from 'react';
import { NavLink } from 'react-router-dom';
import { useApplicationAuth } from '../../auth/AuthProvider';
import { useTranslation } from 'react-i18next';
import { LayoutDashboard, Users, Fingerprint, ShieldAlert, LogOut, Menu, Wifi, WifiOff, RefreshCw, BookOpen } from 'lucide-react';
import { useSecurityStore } from '../../store/securityStore';

// ─── Network Status Hook ─────────────────────────────────────────────────────

type NetStatus = 'online' | 'offline' | 'syncing';

function useNetworkStatus(): NetStatus {
  const [status, setStatus] = React.useState<NetStatus>(
    navigator.onLine ? 'online' : 'offline'
  );

  React.useEffect(() => {
    const handleOnline  = () => setStatus('online');
    const handleOffline = () => setStatus('offline');
    window.addEventListener('online',  handleOnline);
    window.addEventListener('offline', handleOffline);
    return () => {
      window.removeEventListener('online',  handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }, []);

  return status;
}

// ─── NetworkBadge ─────────────────────────────────────────────────────────────

const NetworkBadge: React.FC = () => {
  const { t } = useTranslation();
  const status = useNetworkStatus();

  const config = {
    online:  { label: t('network.online'),   color: 'text-emerald-400 border-emerald-500/30 bg-emerald-500/5', dot: 'bg-emerald-400 animate-pulse', icon: <Wifi size={13} />        },
    offline: { label: t('network.offline'),  color: 'text-red-400 border-red-500/30 bg-red-500/5',             dot: 'bg-red-400',                   icon: <WifiOff size={13} />     },
    syncing: { label: t('network.syncing'),  color: 'text-yellow-400 border-yellow-500/30 bg-yellow-500/5',    dot: 'bg-yellow-400 animate-pulse',  icon: <RefreshCw size={13} className="animate-spin" /> },
  }[status];

  return (
    <div className={`flex items-center gap-2 text-xs px-3 py-1.5 rounded-full border font-medium transition-all ${config.color}`}
         role="status" aria-live="polite" aria-label={`Réseau: ${config.label}`}>
      {config.icon}
      <span className={`w-1.5 h-1.5 rounded-full flex-shrink-0 ${config.dot}`} />
      <span className="hidden sm:inline">{config.label}</span>
    </div>
  );
};

// ─── AppLayout ────────────────────────────────────────────────────────────────

export const AppLayout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { t, i18n } = useTranslation();
  const auth = useApplicationAuth();
  const { hasRole } = useSecurityStore();
  const [isSidebarOpen, setSidebarOpen] = React.useState(true);

  return (
    <div className="flex h-screen overflow-hidden bg-slate-950 text-slate-100">
      {/* Sidebar / Navigation */}
      <nav
        className={`bg-slate-900 border-r border-slate-800 transition-all duration-300 ${isSidebarOpen ? 'w-64' : 'w-20'} flex flex-col`}
        aria-label="Main Navigation"
      >
        <div className="h-16 flex items-center justify-between px-4 border-b border-slate-800">
          {isSidebarOpen && (
            <div className="flex items-center gap-2">
              <div className="w-7 h-7 rounded-md bg-[#1565C0] flex items-center justify-center text-white font-bold text-xs">SN</div>
              <span className="font-bold text-lg tracking-tight text-[#ECEFF1]">SNISID</span>
            </div>
          )}
          <button
            onClick={() => setSidebarOpen(!isSidebarOpen)}
            className="p-2 rounded-lg hover:bg-slate-800 focus:outline-none focus:ring-2 focus:ring-[#1565C0]"
            aria-label="Toggle Sidebar"
            aria-expanded={isSidebarOpen}
          >
            <Menu size={20} />
          </button>
        </div>

        <div className="flex-1 py-6 flex flex-col gap-1.5 px-3 overflow-y-auto">
          <NavItem to="/"           icon={<LayoutDashboard size={20} />} label={t('nav.dashboard')}  isOpen={isSidebarOpen} />
          <NavItem to="/identities" icon={<Users size={20} />}           label={t('nav.identities')} isOpen={isSidebarOpen} />
          <NavItem to="/glossary"   icon={<BookOpen size={20} />}        label={t('nav.glossary')}   isOpen={isSidebarOpen} />

          {hasRole('OPERATOR') && (
            <NavItem to="/biometrics" icon={<Fingerprint size={20} />} label={t('nav.biometrics')} isOpen={isSidebarOpen} />
          )}

          {hasRole('AUDITOR') && (
            <NavItem to="/audit" icon={<ShieldAlert size={20} />} label={t('nav.audit')} isOpen={isSidebarOpen} />
          )}
        </div>

        <div className="p-4 border-t border-slate-800 space-y-3">
          {isSidebarOpen && (
            <select
              value={i18n.language}
              onChange={(e) => i18n.changeLanguage(e.target.value)}
              className="w-full bg-slate-800 border border-slate-700 rounded-lg p-2 text-xs focus:outline-none focus:ring-2 focus:ring-[#1565C0] text-slate-200"
              aria-label="Select Language"
            >
              <option value="fr">🇫🇷 Français</option>
              <option value="ht">🇭🇹 Kreyòl Ayisyen</option>
              <option value="en">🇺🇸 English</option>
            </select>
          )}

          <button
            onClick={() => void auth.signoutRedirect()}
            className={`flex items-center gap-3 w-full p-2 text-red-400 hover:bg-slate-800 rounded-lg transition-colors focus:outline-none focus:ring-2 focus:ring-red-500 ${!isSidebarOpen && 'justify-center'}`}
            aria-label={t('auth.logout')}
          >
            <LogOut size={20} />
            {isSidebarOpen && <span className="text-sm">{t('auth.logout')}</span>}
          </button>
        </div>
      </nav>

      {/* Main Content Area */}
      <main className="flex-1 flex flex-col relative overflow-hidden" id="main-content">
        <header className="h-16 bg-slate-900 border-b border-slate-800 flex items-center justify-between px-6 z-10">
          <h1 className="text-base font-semibold text-slate-200">{t('app.title')}</h1>

          <div className="flex items-center gap-3">
            {/* Dynamic network status indicator — always visible (UX rule) */}
            <NetworkBadge />

            {auth.isDemo && (
              <span className="rounded border border-amber-500/40 bg-amber-500/10 px-2 py-1 text-xs font-semibold text-amber-300" role="status">
                DÉMO — authentification simulée
              </span>
            )}

            <div className="text-sm font-medium bg-slate-800 px-3 py-1.5 rounded-full border border-slate-700 text-slate-200">
              {auth.user?.profile.preferred_username || 'Operateur'}
            </div>
          </div>
        </header>

        <div className="flex-1 overflow-auto p-6 bg-slate-950">
          {children}
        </div>
      </main>
    </div>
  );
};


const NavItem: React.FC<{ to: string; icon: React.ReactNode; label: string; isOpen: boolean }> = ({ to, icon, label, isOpen }) => (
  <NavLink 
    to={to} 
    className={({ isActive }) => 
      `flex items-center gap-3 p-3 rounded-lg transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500
      ${isActive ? 'bg-blue-600 text-white shadow-lg shadow-blue-500/20' : 'text-slate-300 hover:bg-slate-800 hover:text-white'}
      ${!isOpen && 'justify-center'}`
    }
    title={!isOpen ? label : undefined}
    aria-label={label}
  >
    {icon}
    {isOpen && <span className="font-medium">{label}</span>}
  </NavLink>
);
