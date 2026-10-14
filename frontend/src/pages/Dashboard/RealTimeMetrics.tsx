import React from 'react';
import { Activity, CheckCircle2, Fingerprint, ShieldAlert, Users } from 'lucide-react';
import { useApplicationAuth } from '../../auth/AuthProvider';
import { useIdentityStats, useSystemReadiness } from '../../services/dashboardApi';

const formatValue = (value: number | undefined, loading: boolean) => {
  if (loading) return 'Chargement…';
  return typeof value === 'number' ? value.toLocaleString('fr-HT') : 'Données indisponibles';
};

export const RealTimeMetrics: React.FC = () => {
  const auth = useApplicationAuth();
  const stats = useIdentityStats(auth.user?.access_token);
  const readiness = useSystemReadiness();
  const statsLoading = stats.isLoading;
  const systemStatus = readiness.isLoading
    ? 'Vérification en cours'
    : readiness.isError
      ? 'Statut indisponible'
      : readiness.data?.status === 'UP'
        ? 'Système opérationnel'
        : readiness.data?.status === 'DEGRADED'
          ? 'Service dégradé'
          : 'Service indisponible';

  const metrics = [
    { label: 'Identités enregistrées', value: formatValue(stats.data?.total, statsLoading), icon: Users },
    { label: 'Identités vérifiées', value: formatValue(stats.data?.verified_count, statsLoading), icon: CheckCircle2 },
    { label: 'Enrôlements biométriques', value: formatValue(stats.data?.biometric_enrolled_count, statsLoading), icon: Fingerprint },
    { label: 'Alertes de sécurité', value: 'Données indisponibles', icon: ShieldAlert },
  ];

  return (
    <div className="space-y-6 md:space-y-8">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="text-xs font-semibold uppercase text-cyan-200/80">SNISID · Supervision nationale</p>
          <h1 className="mt-2 text-2xl font-semibold text-white sm:text-3xl">Tableau de bord</h1>
          <p className="mt-1 text-sm text-slate-400">Vue globale du Système National d’Identité Numérique</p>
        </div>
        <div className="inline-flex w-fit items-center gap-2 rounded-md border border-white/10 bg-white/[0.04] px-3 py-2 text-sm text-slate-300" role="status" aria-live="polite">
          <span className={`h-2 w-2 rounded-full ${readiness.data?.status === 'UP' ? 'bg-emerald-400' : 'bg-amber-400'}`} />
          {systemStatus}
        </div>
      </header>

      {stats.isError && (
        <div className="flex flex-col gap-3 border-l-2 border-amber-400 bg-amber-400/[0.06] px-4 py-3 sm:flex-row sm:items-center sm:justify-between" role="alert">
          <p className="text-sm text-slate-300">Impossible de charger les statistiques. Vérifiez la connexion au service d’identité.</p>
          <button type="button" onClick={() => void stats.refetch()} className="w-fit text-sm font-medium text-cyan-200 underline decoration-cyan-200/40 underline-offset-4 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-cyan-300">
            Réessayer
          </button>
        </div>
      )}

      <section aria-label="Indicateurs d’identité" className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4">
        {metrics.map(({ label, value, icon: Icon }) => (
          <article key={label} className="rounded-lg border border-white/10 bg-white/[0.045] p-4 shadow-lg shadow-black/10 backdrop-blur-sm sm:p-5">
            <div className="flex items-start justify-between gap-3">
              <h2 className="text-sm font-medium text-slate-300">{label}</h2>
              <Icon size={18} className="shrink-0 text-cyan-200" aria-hidden="true" />
            </div>
            <p className={`mt-5 break-words text-xl font-semibold ${value === 'Données indisponibles' ? 'text-slate-500' : 'text-white'}`}>
              {value}
            </p>
          </article>
        ))}
      </section>

      <section className="grid grid-cols-1 gap-4 xl:grid-cols-[1.4fr_1fr]">
        <article className="rounded-lg border border-white/10 bg-slate-900/70 p-4 sm:p-5">
          <div className="flex items-start gap-3">
            <Activity size={18} className="mt-0.5 text-cyan-200" aria-hidden="true" />
            <div>
              <h2 className="font-semibold text-white">Activité nationale</h2>
              <p className="mt-1 text-sm text-slate-400">Aucun endpoint d’historique d’activité n’a été identifié.</p>
            </div>
          </div>
          <div className="mt-6 flex min-h-40 items-center justify-center border-t border-white/[0.07] text-sm text-slate-500" role="status">
            Données indisponibles
          </div>
        </article>

        <article className="rounded-lg border border-white/10 bg-slate-900/70 p-4 sm:p-5">
          <h2 className="font-semibold text-white">État des services</h2>
          {readiness.isLoading ? (
            <p className="mt-5 text-sm text-slate-400" role="status">Vérification des services…</p>
          ) : readiness.isError ? (
            <p className="mt-5 text-sm text-amber-200" role="status">Statut indisponible</p>
          ) : readiness.data?.checks?.length ? (
            <ul className="mt-4 divide-y divide-white/[0.07]">
              {readiness.data.checks?.map((check) => (
                <li key={check.name} className="flex items-center justify-between gap-3 py-3 text-sm">
                  <span className="capitalize text-slate-300">{check.name}</span>
                  <span className={`inline-flex items-center gap-2 ${check.status === 'UP' ? 'text-emerald-300' : 'text-amber-200'}`}>
                    <span className={`h-1.5 w-1.5 rounded-full ${check.status === 'UP' ? 'bg-emerald-400' : 'bg-amber-400'}`} />
                    {check.status === 'UP' ? 'Opérationnel' : check.status === 'DEGRADED' ? 'Dégradé' : 'Indisponible'}
                  </span>
                </li>
              ))}
            </ul>
          ) : (
            <p className="mt-5 text-sm text-slate-400" role="status">Aucun détail de service fourni par le backend.</p>
          )}
        </article>
      </section>

      <section className="rounded-lg border border-white/10 bg-slate-900/70 p-4 sm:p-5">
        <h2 className="font-semibold text-white">Activité récente</h2>
        <p className="mt-1 text-sm text-slate-400">Aucun endpoint de lecture des événements n’a été identifié.</p>
        <div className="mt-4 border-t border-white/[0.07] py-8 text-center text-sm text-slate-500" role="status">
          Aucune donnée disponible
        </div>
      </section>
    </div>
  );
};
