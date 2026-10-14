import React from 'react';
import { Search } from 'lucide-react';
import { useApplicationAuth } from '../auth/AuthProvider';
import { fetchIdentities } from '../services/dashboardApi';
import { useQuery } from '@tanstack/react-query';

export const IdentityRegistryPage: React.FC = () => {
  const auth = useApplicationAuth();
  const [input, setInput] = React.useState('');
  const [searchTerm, setSearchTerm] = React.useState('');
  const [page, setPage] = React.useState(1);
  const identities = useQuery({
    queryKey: ['identities', searchTerm, page, auth.user?.access_token],
    queryFn: ({ signal }) => fetchIdentities(searchTerm, page, auth.user?.access_token, signal),
    enabled: import.meta.env.DEV || !!auth.user?.access_token,
    retry: 1,
    staleTime: 15_000,
  });

  const submitSearch = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setPage(1);
    setSearchTerm(input.trim());
  };

  return (
    <div className="space-y-6">
      <header>
        <p className="text-xs font-semibold uppercase text-cyan-200/80">SNISID · Identité</p>
        <h1 className="mt-2 text-2xl font-semibold text-white sm:text-3xl">Registre national</h1>
        <p className="mt-1 text-sm text-slate-400">Recherche des enregistrements disponibles dans le service d’identité.</p>
      </header>

      <form onSubmit={submitSearch} className="flex flex-col gap-2 sm:flex-row" role="search">
        <label className="sr-only" htmlFor="identity-search">NIF ou nom</label>
        <div className="relative min-w-0 flex-1">
          <Search size={17} className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" aria-hidden="true" />
          <input id="identity-search" value={input} onChange={(event) => setInput(event.target.value)} placeholder="NIF, nom ou prénom" className="min-h-11 w-full rounded-md border border-white/10 bg-white/[0.04] pl-10 pr-3 text-sm text-white placeholder:text-slate-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-cyan-300" />
        </div>
        <button type="submit" className="min-h-11 rounded-md bg-[#0F3D91] px-5 text-sm font-medium text-white hover:bg-[#1558C0] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-cyan-300">Rechercher</button>
      </form>

      <section className="overflow-hidden rounded-lg border border-white/10 bg-slate-900/70" aria-label="Résultats du registre">
        {identities.isLoading ? (
          <div className="p-6 text-sm text-slate-300" role="status">Chargement des identités…</div>
        ) : identities.isError ? (
          <div className="flex flex-col gap-3 p-6 sm:flex-row sm:items-center sm:justify-between" role="alert">
            <p className="text-sm text-slate-300">Impossible de charger les identités. Service non connecté.</p>
            <button type="button" onClick={() => void identities.refetch()} className="w-fit text-sm font-medium text-cyan-200 underline underline-offset-4 focus-visible:outline focus-visible:outline-2 focus-visible:outline-cyan-300">Réessayer</button>
          </div>
        ) : !identities.data?.items.length ? (
          <div className="p-10 text-center text-sm text-slate-400" role="status">Aucune donnée disponible</div>
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="w-full min-w-[680px] border-collapse text-left text-sm">
                <thead className="bg-white/[0.035] text-xs text-slate-400">
                  <tr><th scope="col" className="px-4 py-3 font-medium">Identifiant national</th><th scope="col" className="px-4 py-3 font-medium">Nom complet</th><th scope="col" className="px-4 py-3 font-medium">Statut</th><th scope="col" className="px-4 py-3 font-medium">Vérification</th><th scope="col" className="px-4 py-3 font-medium">Biométrie</th></tr>
                </thead>
                <tbody className="divide-y divide-white/[0.07]">
                  {identities.data.items.map((identity) => (
                    <tr key={identity.id} className="hover:bg-white/[0.025]">
                      <td className="whitespace-nowrap px-4 py-3 font-mono text-xs text-cyan-100">{identity.national_id}</td>
                      <td className="px-4 py-3 text-slate-200">{identity.full_name || `${identity.first_name} ${identity.last_name}`}</td>
                      <td className="px-4 py-3 text-slate-300">{identity.status || 'Statut indisponible'}</td>
                      <td className="px-4 py-3 text-slate-300">{identity.verified ? 'Vérifiée' : 'Non vérifiée'}</td>
                      <td className="px-4 py-3 text-slate-300">{identity.has_biometrics ? 'Enrôlée' : 'Non enrôlée'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <div className="flex flex-col gap-3 border-t border-white/[0.07] px-4 py-3 text-sm text-slate-400 sm:flex-row sm:items-center sm:justify-between">
              <span>{identities.data.total.toLocaleString('fr-HT')} résultats · page {identities.data.page} sur {Math.max(identities.data.total_pages, 1)}</span>
              <div className="flex gap-2">
                <button type="button" disabled={page <= 1} onClick={() => setPage((current) => current - 1)} className="min-h-9 rounded-md border border-white/10 px-3 text-slate-200 hover:bg-white/[0.05] disabled:cursor-not-allowed disabled:opacity-40">Précédent</button>
                <button type="button" disabled={page >= identities.data.total_pages} onClick={() => setPage((current) => current + 1)} className="min-h-9 rounded-md border border-white/10 px-3 text-slate-200 hover:bg-white/[0.05] disabled:cursor-not-allowed disabled:opacity-40">Suivant</button>
              </div>
            </div>
          </>
        )}
      </section>
    </div>
  );
};