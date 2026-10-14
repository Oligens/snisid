import { BrowserRouter, Routes, Route, Navigate, Outlet } from 'react-router-dom';
import { InstitutionalLayout } from './components/layout/InstitutionalLayout';
import { ProtectedRoute } from './components/auth/ProtectedRoute';
import { RealTimeMetrics } from './pages/Dashboard/RealTimeMetrics';
import { CaptureInterface } from './components/biometrics/CaptureInterface';
import { IdentityRegistryPage } from './pages/IdentityRegistryPage';
import { ModuleStatusPage } from './pages/ModuleStatusPage';
import { AuditPage } from './pages/AuditPage';
import { GlossaryPage } from './pages/GlossaryPage';
import SigintDashboard from './pages/SigintDashboard';
import HumintDashboard from './pages/HumintDashboard';
import AirDefenseCOP from './pages/AirDefenseCOP';
import MilC2Dashboard from './pages/MilC2Dashboard';
import BioSurveillanceDashboard from './pages/BioSurveillanceDashboard';
import ExecProtectionDashboard from './pages/ExecProtectionDashboard';
import TransportSecurityDashboard from './pages/TransportSecurityDashboard';
import RadiationSafetyDashboard from './pages/RadiationSafetyDashboard';
import FusionCenterDashboard from './pages/FusionCenterDashboard';
import CounterintelDashboard from './pages/CounterintelDashboard';
import CriticalInfraDashboard from './pages/CriticalInfraDashboard';
import FisaCourtDashboard from './pages/FisaCourtDashboard';
import ClassificationDashboard from './pages/ClassificationDashboard';

// Biometric page wrapper
const BiometricPage = () => (
  <div className="flex flex-col items-center justify-center h-full">
    <CaptureInterface onCapture={(blob) => console.log('Captured', blob)} />
  </div>
);

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/preview" element={<InstitutionalLayout><RealTimeMetrics /></InstitutionalLayout>} />
        {/* Protected Routes Wrapper */}
        <Route element={<ProtectedRoute />}>
          <Route element={<InstitutionalLayout><Outlet /></InstitutionalLayout>}>
            {/* Dashboard — accessible à tous les agents connectés */}
            <Route path="/" element={<RealTimeMetrics />} />
            <Route path="/dashboard" element={<RealTimeMetrics />} />

            {/* Identity reads use the verified FastAPI list/search contract. */}
            <Route path="/identities" element={<IdentityRegistryPage />} />
            <Route path="/identity/search" element={<IdentityRegistryPage />} />
            <Route path="/identity/verify" element={<ModuleStatusPage title="Vérification d’identité" apiHint="POST /v1/identities/{identity_id}/verify" />} />
            <Route path="/identity/history" element={<ModuleStatusPage title="Historique d’identité" apiHint="GET /v1/identities/{identity_id}/history" />} />
            <Route path="/identity" element={<ModuleStatusPage title="Module identité" />} />

            {/* Biométrie — opérateur uniquement */}
            <Route element={<ProtectedRoute requiredRoles={['OPERATOR']} />}>
              <Route path="/biometrics" element={<BiometricPage />} />
              <Route path="/biometrics/*" element={<ModuleStatusPage title="Module biométrie" />} />
            </Route>

            {/* Audit & Investigation forensique — auditeur uniquement */}
            <Route element={<ProtectedRoute requiredRoles={['AUDITOR']} />}>
              <Route path="/audit" element={<AuditPage />} />
            </Route>

            <Route path="/documents/*" element={<ModuleStatusPage title="Documents nationaux" apiHint="POST /v1/identities/{identity_id}/documents" />} />
            <Route path="/iam/*" element={<ModuleStatusPage title="Gestion des identités et des accès" />} />
            <Route path="/pki/*" element={<ModuleStatusPage title="Infrastructure à clés publiques" apiHint="/v1/pki" />} />
            <Route path="/fraud/*" element={<ModuleStatusPage title="Fraude et risque" />} />
            <Route path="/security/*" element={<ModuleStatusPage title="Centre de sécurité" />} />
            <Route path="/interoperability/*" element={<ModuleStatusPage title="Interopérabilité" />} />
            <Route path="/reports/*" element={<ModuleStatusPage title="Rapports et statistiques" />} />
            <Route element={<ProtectedRoute requiredRoles={['SUPER_ADMIN']} />}>
              <Route path="/admin/*" element={<ModuleStatusPage title="Administration" />} />
            </Route>

            {/* National Security Dashboards */}
            <Route path="/dashboard/sigint" element={<SigintDashboard />} />
            <Route path="/dashboard/humint" element={<HumintDashboard />} />
            <Route path="/dashboard/air-defense" element={<AirDefenseCOP />} />
            <Route path="/dashboard/mil-c2" element={<MilC2Dashboard />} />
            <Route path="/dashboard/bio-surveillance" element={<BioSurveillanceDashboard />} />
            <Route path="/dashboard/exec-protection" element={<ExecProtectionDashboard />} />
            <Route path="/dashboard/transport-security" element={<TransportSecurityDashboard />} />
            <Route path="/dashboard/radiation-safety" element={<RadiationSafetyDashboard />} />
            <Route path="/dashboard/fusion-center" element={<FusionCenterDashboard />} />
            <Route path="/dashboard/counterintel" element={<CounterintelDashboard />} />
            <Route path="/dashboard/critical-infrastructure" element={<CriticalInfraDashboard />} />
            <Route path="/dashboard/fisa-court" element={<FisaCourtDashboard />} />
            <Route path="/dashboard/classification" element={<ClassificationDashboard />} />

            {/* Glossaire Technique — accessible à tous */}
            <Route path="/glossary" element={<GlossaryPage />} />
          </Route>
        </Route>

        {/* Fallback */}
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;

