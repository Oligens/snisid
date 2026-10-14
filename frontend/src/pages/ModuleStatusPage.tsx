import React from 'react';
import { CircleHelp } from 'lucide-react';

export const ModuleStatusPage: React.FC<{ title: string; description?: string; apiHint?: string }> = ({ title, description, apiHint }) => (
  <div className="mx-auto max-w-4xl space-y-6">
    <header>
      <p className="text-xs font-semibold uppercase text-cyan-200/80">SNISID · Module</p>
      <h1 className="mt-2 text-2xl font-semibold text-white sm:text-3xl">{title}</h1>
      <p className="mt-1 text-sm text-slate-400">{description || 'Cette vue est en attente de son intégration frontend.'}</p>
    </header>
    <section className="flex gap-4 rounded-lg border border-white/10 bg-white/[0.035] p-5 sm:p-6" role="status">
      <CircleHelp size={20} className="mt-0.5 shrink-0 text-cyan-200" aria-hidden="true" />
      <div>
        <h2 className="font-medium text-white">Service non connecté à cette vue</h2>
        <p className="mt-1 text-sm leading-6 text-slate-400">Aucune donnée ne sera affichée tant qu’une API de lecture et son contrat d’accès n’auront pas été validés.</p>
        {apiHint && <p className="mt-3 text-xs text-slate-500">Endpoint repéré dans le dépôt : <code className="text-slate-300">{apiHint}</code>. Intégration à vérifier.</p>}
      </div>
    </section>
  </div>
);