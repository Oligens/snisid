import axios from 'axios';
import { useQuery } from '@tanstack/react-query';

export interface IdentityStats {
  total: number;
  by_status: Record<string, number>;
  by_nationality: Record<string, number>;
  verified_count: number;
  biometric_enrolled_count: number;
}

export interface ServiceCheck {
  name: string;
  status: 'UP' | 'DOWN' | 'DEGRADED' | string;
  latency_ms?: number;
}

export interface ReadinessStatus {
  status: 'UP' | 'DOWN' | 'DEGRADED' | string;
  timestamp: string;
  service: string;
  version: string;
  checks: ServiceCheck[];
}

export interface IdentityRecord {
  id: string;
  national_id: string;
  first_name: string;
  last_name: string;
  full_name: string;
  status: string;
  verified: boolean;
  has_biometrics: boolean;
  created_at: string;
}

export interface IdentitySearchResponse {
  items: IdentityRecord[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

const apiClient = axios.create({ timeout: 5000 });

const requestConfig = (accessToken?: string, signal?: AbortSignal) => ({
  signal,
  headers: accessToken ? { Authorization: `Bearer ${accessToken}` } : undefined,
});

export async function fetchIdentityStats(accessToken?: string, signal?: AbortSignal) {
  const response = await apiClient.get<IdentityStats>(
    '/v1/identities/stats',
    requestConfig(accessToken, signal),
  );
  return response.data;
}

export async function fetchIdentities(searchTerm: string, page: number, accessToken?: string, signal?: AbortSignal) {
  const response = await apiClient.get<IdentitySearchResponse>('/v1/identities', {
    ...requestConfig(accessToken, signal),
    params: { search_term: searchTerm || undefined, page, page_size: 20 },
  });
  return response.data;
}

export async function fetchSystemReadiness(signal?: AbortSignal) {
  const response = await apiClient.get<ReadinessStatus>('/ready', {
    ...requestConfig(undefined, signal),
    validateStatus: (status) => status < 600,
  });
  return response.data;
}

export function useIdentityStats(accessToken?: string) {
  return useQuery({
    queryKey: ['identity-stats', accessToken],
    queryFn: ({ signal }) => fetchIdentityStats(accessToken, signal),
    enabled: !!accessToken,
    retry: 1,
    staleTime: 30_000,
  });
}

export function useSystemReadiness() {
  return useQuery({
    queryKey: ['system-readiness'],
    queryFn: ({ signal }) => fetchSystemReadiness(signal),
    retry: 1,
    refetchInterval: 30_000,
    staleTime: 10_000,
  });
}