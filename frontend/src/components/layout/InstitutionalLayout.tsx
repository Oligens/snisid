import React from 'react';
import { NavLink } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  Activity, Bell, BookOpen, Building2, ChevronDown, FileText, Fingerprint,
  KeyRound, LayoutDashboard, LogOut, Menu, Network, Search, Settings,
  Shield, ShieldAlert, Users, X,
} from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { useLocation } from 'react-router-dom';
import { useApplicationAuth } from '../../auth/AuthProvider';
import { useSystemReadiness } from '../../services/dashboardApi';
import { useSecurityStore } from '../../store/securityStore';

type Item = { label: string; to: string; icon: LucideIcon; role?: string };

const groups: { label: string; items: Item[] }[] = [
  { label: 'Identité', items: [
    { label: 'Registre national', to: '/identities', icon: Users },
    { label: 'Recherche', to: '/identity/search', icon: Search },
    { label: 'Vérification', to: '/identity/verify', icon: Shield },
    { label: 'Historique', to: '/identity/history', icon: Activity },
  ] },
  { label: 'Biométrie', items: [
    { label: 'Enrôlement', to: '/biometrics', icon: Fingerprint, role: 'OPERATOR' },
    { label: 'Vérification', to: '/biometrics/verification', icon: Shield, role: 'OPERATOR' },
    { label: 'Identification', to: '/biometrics/identification', icon: Search, role: 'OPERATOR' },
    { label: 'Déduplication', to: '/biometrics/deduplication', icon: Users, role: 'OPERATOR' },
  ] },
  { label: 'Documents', items: [
    { label: 'Cartes nationales', to: '/documents', icon: FileText },
    { label: 'Émission', to: '/documents/issuance', icon: FileText },
    { label: 'Renouvellement', to: '/documents/renewal', icon: Activity },
    { label: 'Révocation', to: '/documents/revocation', icon: ShieldAlert },
  ] },
  { label: 'Sécurité', items: [
    { label: 'IAM', to: '/iam', icon: KeyRound },
    { label: 'PKI', to: '/pki', icon: Shield },
    { label: 'Fraude & risque', to: '/fraud', icon: ShieldAlert },
    { label: 'SOC / Audit', to: '/audit', icon: Activity, role: 'AUDITOR' },
  ] },
  { label: 'Interopérabilité', items: [
    { label: 'API & partenaires', to: '/interoperability', icon: Network },
    { label: 'Connecteurs', to: '/interoperability/connectors', icon: Building2 },
    { label: 'Transactions', to: '/interoperability/transactions', icon: Activity },
  ] },
  { label: 'Données', items: [
    { label: 'Statistiques', to: '/reports', icon: Activity },
    { label: 'Rapports', to: '/reports/national', icon: FileText },
    { label: 'Exports', to: '/reports/exports', icon: FileText },
  ] },
  { label: 'Administration', items: [
    { label: 'Utilisateurs & rôles', to: '/admin', icon: Users, role: 'SUPER_ADMIN' },
    { label: 'Organisations', to: '/admin/organizations', icon: Building2, role: 'SUPER_ADMIN' },
    { label: 'Configuration', to: '/admin/configuration', icon: Settings, role: 'SUPER_ADMIN' },
    { label: 'Audit système', to: '/admin/audit', icon: BookOpen, role: 'SUPER_ADMIN' },
  ] },
];

const legacyPreviewRoutes = [
  '/audit',
  '/dashboard/sigint', '/dashboard/humint', '/dashboard/air-defense',
  '/dashboard/mil-c2', '/dashboard/bio-surveillance', '/dashboard/exec-protection',
  '/dashboard/transport-security', '/dashboard/radiation-safety',
  '/dashboard/fusion-center', '/dashboard/counterintel',
  '/dashboard/critical-infrastructure', '/dashboard/fisa-court', '/dashboard/classification',
];

function getSidebarDefault() {
  try {
    return window.localStorage.getItem('snisid.sidebar.collapsed') !== 'true';
  } catch {
    return true;
  }
}

export const InstitutionalLayout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { t, i18n } = useTranslation();
  const location = useLocation();
  const auth = useApplicationAuth();
  const { hasRole } = useSecurityStore();
  const readiness = useSystemReadiness();
  const [expanded, setExpanded] = React.useState(getSidebarDefault);
  const [mobileOpen, setMobileOpen] = React.useState(false);
  const [openGroups, setOpenGroups] = React.useState<Record<string, boolean>>({ Identité: true });
  const [notificationsOpen, setNotificationsOpen] = React.useState(false);
  const [profileOpen, setProfileOpen] = React.useState(false);
  const labelsVisible = expanded || mobileOpen;
  const isLegacyPreview = legacyPreviewRoutes.some((route) => location.pathname === route || location.pathname.startsWith(`${route}/`));
  const systemStatus = readiness.isLoading ? 'Vérification…' : readiness.isError ? 'Statut indisponible'
    : readiness.data?.status === 'UP' ? 'Système opérationnel'
      : readiness.data?.status === 'DEGRADED' ? 'Service dégradé' : 'Service indisponible';

  const toggleExpanded = () => {
    setExpanded((current) => {
      try {
        window.localStorage.setItem('snisid.sidebar.collapsed', String(current));
      } catch {
        // Keep the control usable if storage is unavailable.
      }
      return !current;
    });
  };

  return (
    <div className="min-h-screen bg-snisid-bg text-snisid-text lg:flex">
      {mobileOpen && <button className="fixed inset-0 z-30 bg-black/60 lg:hidden" aria-label="Fermer la navigation" onClick={() => setMobileOpen(false)} />}
      <aside className={`fixed inset-y-0 left-0 z-40 flex w-[min(18rem,calc(100vw-3rem))] flex-col border-r border-white/10 bg-[#0a1021] transition-[transform,width] duration-200 lg:sticky lg:top-0 lg:h-screen lg:translate-x-0 ${mobileOpen ? 'translate-x-0' : '-translate-x-full'} ${expanded ? 'lg:w-64' : 'lg:w-20'}`}>
        <div className="flex h-16 shrink-0 items-center justify-between border-b border-white/10 px-4">
          <div className={`min-w-0 ${labelsVisible ? 'flex items-center gap-3' : 'hidden lg:block'}`}>
            <div className="flex h-9 w-9 items-center justify-center rounded-md border border-cyan-200/20 bg-[#0F3D91]/50 text-xs font-bold text-white">SN</div>
            {labelsVisible && <div><p className="text-sm font-semibold text-white">SNISID</p><p className="text-[10px] leading-tight text-slate-400">Système National<br />d’Identité Numérique</p></div>}
          </div>
          <button type="button" onClick={() => mobileOpen ? setMobileOpen(false) : toggleExpanded()} aria-label={mobileOpen ? 'Fermer le menu' : expanded ? 'Réduire la barre latérale' : 'Développer la barre latérale'} aria-expanded={mobileOpen || expanded} className="rounded-md p-2 text-slate-300 hover:bg-white/10 focus-visible:outline focus-visible:outline-2 focus-visible:outline-cyan-300">
            {mobileOpen ? <X size={18} /> : <Menu size={18} />}
          </button>
        </div>
        <nav className="custom-scrollbar flex-1 overflow-y-auto px-3 py-4" aria-label="Navigation principale">
          <SidebarLink to="/dashboard" label="Tableau de bord" icon={LayoutDashboard} showLabel={labelsVisible} onNavigate={() => setMobileOpen(false)} />
          <div className="my-3 border-t border-white/[0.08]" />
          {groups.map((group) => {
            const items = group.items.filter((item) => !item.role || hasRole(item.role));
            if (!items.length) return null;
            const isOpen = openGroups[group.label] ?? false;
            return <section key={group.label} className="mb-2">
              <button type="button" onClick={() => setOpenGroups((current) => ({ ...current, [group.label]: !isOpen }))} aria-expanded={isOpen} title={labelsVisible ? undefined : group.label} className={`flex min-h-9 w-full items-center justify-between rounded-md px-3 text-left text-[11px] font-semibold uppercase text-slate-500 hover:text-slate-300 focus-visible:outline focus-visible:outline-2 focus-visible:outline-cyan-300 ${labelsVisible ? '' : 'lg:justify-center'}`}>
                {labelsVisible && <span>{group.label}</span>}{labelsVisible && <ChevronDown size={14} className={`transition-transform ${isOpen ? '' : '-rotate-90'}`} aria-hidden="true" />}
              </button>
              {(isOpen || !labelsVisible) && <div className="mt-1 space-y-1">{items.map((item) => <SidebarLink key={item.to} {...item} showLabel={labelsVisible} onNavigate={() => setMobileOpen(false)} />)}</div>}
            </section>;
          })}
          <div className="my-3 border-t border-white/[0.08]" />
          <SidebarLink to="/glossary" label={t('nav.glossary')} icon={BookOpen} showLabel={labelsVisible} onNavigate={() => setMobileOpen(false)} />
        </nav>
        <div className="shrink-0 space-y-3 border-t border-white/10 p-3">
          <label className="sr-only" htmlFor="language-select">Langue</label>
          <select id="language-select" value={i18n.language} onChange={(event) => i18n.changeLanguage(event.target.value)} aria-label="Choisir la langue" className={`min-h-10 rounded-md border border-white/10 bg-white/[0.04] px-2 text-xs text-slate-200 focus-visible:outline focus-visible:outline-2 focus-visible:outline-cyan-300 ${labelsVisible ? 'w-full' : 'hidden'}`}>
            <option value="fr">Français</option><option value="ht">Kreyòl Ayisyen</option><option value="en">English</option>
          </select>
          {auth.isDemo && labelsVisible && <p className="px-2 text-xs text-amber-200" role="status">Mode de démonstration</p>}
        </div>
      </aside>

      <div className="flex min-h-screen min-w-0 flex-1 flex-col">
        <header className="sticky top-0 z-20 flex h-16 shrink-0 items-center gap-3 border-b border-white/10 bg-[#070B19]/90 px-3 backdrop-blur-md sm:px-5 lg:px-7">
          <button type="button" onClick={() => setMobileOpen(true)} aria-label="Ouvrir la navigation" className="rounded-md p-2 text-slate-300 hover:bg-white/10 focus-visible:outline focus-visible:outline-2 focus-visible:outline-cyan-300 lg:hidden"><Menu size={19} /></button>
          <div className="hidden min-w-0 flex-1 sm:block"><label className="sr-only" htmlFor="global-search">Recherche globale</label><div className="relative max-w-xl"><Search size={16} className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" aria-hidden="true" /><input id="global-search" disabled placeholder="Recherche globale non connectée" className="h-10 w-full cursor-not-allowed rounded-md border border-white/[0.08] bg-white/[0.03] pl-9 pr-3 text-sm text-slate-400 placeholder:text-slate-500" /></div></div>
          <div className="ml-auto flex items-center gap-2 sm:gap-3">
            <div className="hidden items-center gap-2 text-xs text-slate-300 md:flex" role="status" aria-live="polite"><span className={`h-2 w-2 rounded-full ${readiness.data?.status === 'UP' ? 'bg-emerald-400' : 'bg-amber-400'}`} />{systemStatus}</div>
            <div className="relative">
              <button type="button" onClick={() => { setNotificationsOpen(!notificationsOpen); setProfileOpen(false); }} aria-label="Notifications" aria-expanded={notificationsOpen} className="rounded-md p-2 text-slate-300 hover:bg-white/10 focus-visible:outline focus-visible:outline-2 focus-visible:outline-cyan-300"><Bell size={18} /></button>
              {notificationsOpen && <div className="absolute right-0 top-12 z-50 w-72 rounded-lg border border-white/10 bg-[#10182a] p-4 shadow-2xl" role="dialog" aria-label="Notifications système"><h2 className="text-sm font-semibold text-white">Notifications</h2><p className="mt-2 text-sm text-slate-400">Aucun flux de notification connecté.</p></div>}
            </div>
            <div className="relative">
              <button type="button" onClick={() => { setProfileOpen(!profileOpen); setNotificationsOpen(false); }} aria-label="Ouvrir le menu du profil" aria-expanded={profileOpen} className="flex min-h-10 items-center gap-2 rounded-md border border-white/10 px-2 text-sm text-slate-200 hover:bg-white/[0.05] focus-visible:outline focus-visible:outline-2 focus-visible:outline-cyan-300 sm:px-3"><span className="flex h-7 w-7 items-center justify-center rounded-full bg-[#0F3D91]/60 text-xs font-semibold text-white" aria-hidden="true">{(auth.user?.profile.preferred_username?.[0] || 'U').toUpperCase()}</span><span className="hidden max-w-36 truncate sm:block">{auth.user?.profile.preferred_username || 'Utilisateur'}</span></button>
              {profileOpen && <div className="absolute right-0 top-12 z-50 min-w-48 rounded-lg border border-white/10 bg-[#10182a] p-2 shadow-2xl" role="menu"><p className="truncate px-3 py-2 text-xs text-slate-400">{auth.user?.profile.email || auth.user?.profile.preferred_username || 'Utilisateur'}</p><button type="button" role="menuitem" onClick={() => void auth.signoutRedirect()} className="flex min-h-10 w-full items-center gap-2 rounded-md px-3 text-left text-sm text-red-200 hover:bg-white/[0.06] focus-visible:outline focus-visible:outline-2 focus-visible:outline-cyan-300"><LogOut size={16} />Déconnexion</button></div>}
            </div>
          </div>
        </header>
        <main id="main-content" className="flex-1 p-4 sm:p-6 lg:p-8" tabIndex={-1}>
          {isLegacyPreview && <div className="mb-5 border-l-2 border-amber-400 bg-amber-400/[0.06] px-4 py-3 text-sm text-amber-100" role="status"><strong>Demo / Preview :</strong> cette vue historique peut contenir des données d’exemple non vérifiées. Ne pas les utiliser pour une décision opérationnelle.</div>}
          {children}
        </main>
      </div>
    </div>
  );
};

const SidebarLink: React.FC<Item & { showLabel: boolean; onNavigate: () => void }> = ({ to, label, icon: Icon, showLabel, onNavigate }) => (
  <NavLink to={to} onClick={onNavigate} title={showLabel ? undefined : label} aria-label={label} className={({ isActive }) => `flex min-h-10 items-center gap-3 rounded-md border px-3 text-sm transition-colors focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-cyan-300 ${showLabel ? '' : 'justify-center'} ${isActive ? 'border-[#1558C0]/40 bg-[#0F3D91]/30 text-white' : 'border-transparent text-slate-400 hover:border-white/[0.07] hover:bg-white/[0.05] hover:text-slate-100'}`}>
    <Icon size={17} className="shrink-0" aria-hidden="true" />{showLabel && <span className="truncate">{label}</span>}
  </NavLink>
);