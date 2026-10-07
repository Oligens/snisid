import React from 'react';
import { Shield, CheckCircle, AlertTriangle, FileText, Lock, RefreshCw, Clock } from 'lucide-react';
import { motion } from 'framer-motion';

const KpiCard: React.FC<{ title: string; value: string; icon: React.ReactNode; trend: string }> = ({ title, value, icon, trend }) => (
  <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} className="bg-[#0f1218] border border-white/5 rounded-2xl p-6 shadow-xl flex flex-col gap-4">
    <div className="flex items-start justify-between">
      <h3 className="text-gray-400 font-medium text-sm uppercase tracking-wider">{title}</h3>
      <div className="p-2 bg-white/5 rounded-xl">{icon}</div>
    </div>
    <div>
      <div className="text-3xl font-bold text-white mb-1">{value}</div>
      <div className="text-xs text-gray-500">{trend}</div>
    </div>
  </motion.div>
);

const FipsDashboard = () => {
  return (
    <div className="min-h-screen bg-[#0a0c10] p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-black tracking-tighter text-white">FIPS 140 Compliance</h1>
          <p className="text-gray-500 uppercase tracking-[0.2em] text-[10px] font-bold mt-1">Cryptographic Module Validation</p>
        </div>
        <div className="flex items-center gap-2 px-3 py-1.5 bg-emerald-500/10 border border-emerald-500/20 rounded-full">
          <span className="w-2 h-2 rounded-full bg-emerald-500" />
          <span className="text-emerald-500 text-xs font-bold uppercase tracking-wider">Level 3 Compliant</span>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <KpiCard title="FIPS Modules" value="12" icon={<Shield className="w-5 h-5 text-blue-500" />} trend="All validated" />
        <KpiCard title="Algorithms Passed" value="48/48" icon={<CheckCircle className="w-5 h-5 text-emerald-500" />} trend="CAVP verified" />
        <KpiCard title="Pending CMVP" value="2" icon={<Clock className="w-5 h-5 text-amber-500" />} trend="In review" />
        <KpiCard title="Security Policies" value="24" icon={<FileText className="w-5 h-5 text-violet-500" />} trend="100% documented" />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-[#0f1218] border border-white/5 rounded-2xl p-6 shadow-xl">
          <h3 className="text-lg font-bold text-white mb-4 flex items-center gap-2"><Lock className="w-4 h-4 text-blue-500" /> Cryptographic Inventory</h3>
          <div className="space-y-2">
            {[
              { module: 'AES-256-GCM', mode: 'Encryption', cert: 'C2056', status: 'Active' },
              { module: 'RSA-4096 PKCS#1', mode: 'Signing', cert: 'C1892', status: 'Active' },
              { module: 'ECDSA P-384', mode: 'Signing', cert: 'C2123', status: 'Active' },
              { module: 'SHA-3-512', mode: 'Hashing', cert: 'C1789', status: 'Active' },
              { module: 'HMAC-SHA3-512', mode: 'MAC', cert: 'C1901', status: 'Active' },
            ].map((c) => (
              <div key={c.module} className="flex items-center justify-between p-3 bg-white/5 rounded-xl">
                <div>
                  <div className="text-sm font-medium text-white">{c.module}</div>
                  <div className="text-xs text-gray-500">{c.mode} - {c.cert}</div>
                </div>
                <span className={`px-2 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider bg-emerald-500/10 text-emerald-500`}>{c.status}</span>
              </div>
            ))}
          </div>
        </div>

        <div className="bg-[#0f1218] border border-white/5 rounded-2xl p-6 shadow-xl">
          <h3 className="text-lg font-bold text-white mb-4">Validation Status</h3>
          <div className="space-y-4">
            {[
              { label: 'FIPS 140-3 Level 3', val: 'Certified', color: 'bg-emerald-500' },
              { label: 'FIPS 140-3 Level 2', val: 'In Progress', color: 'bg-amber-500' },
              { label: 'CAVP Algorithm Validation', val: '100%', color: 'bg-emerald-500' },
              { label: 'CMVP Module Validation', val: '83%', color: 'bg-blue-500' },
            ].map((m) => (
              <div key={m.label}>
                <div className="flex justify-between text-sm mb-1">
                  <span className="text-gray-400">{m.label}</span>
                  <span className="text-white font-bold">{m.val}</span>
                </div>
                <div className="h-1.5 bg-white/5 rounded-full overflow-hidden">
                  <div className={`h-full ${m.color} rounded-full`} style={{ width: m.val === 'Certified' ? '100%' : m.val === 'In Progress' ? '45%' : '90%' }} />
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};

export default FipsDashboard;
