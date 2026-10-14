import React, { useEffect, useState, createContext, useContext } from 'react';
import { AuthProvider as OidcProvider, useAuth as useOidcAuth } from 'react-oidc-context';
import { WebStorageStateStore } from 'oidc-client-ts';

// Configure Keycloak OIDC Settings
const oidcConfig = {
  authority: import.meta.env.VITE_KEYCLOAK_URL || 'http://localhost:8080/realms/snisid',
  client_id: import.meta.env.VITE_KEYCLOAK_CLIENT_ID || 'snisid-frontend',
  redirect_uri: window.location.origin + '/callback',
  post_logout_redirect_uri: window.location.origin + '/',
  response_type: 'code',
  scope: 'openid profile email roles',
  // Store OIDC state in session storage, not local storage (better for PII isolation)
  userStore: new WebStorageStateStore({ store: window.sessionStorage }),
  automaticSilentRenew: true,
  revokeAccessTokenOnSignout: true,
};

const AuthTimeoutContext = createContext<() => void>(() => {});

export const useAuthTimeout = () => useContext(AuthTimeoutContext);

export const useApplicationAuth = () => {
  const auth = useOidcAuth();
  const isDemo = import.meta.env.DEV && !!auth.error;

  return {
    ...auth,
    isDemo,
  };
};

const SessionTimeoutWrapper: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const auth = useOidcAuth();
  const [lastActivity, setLastActivity] = useState<number>(Date.now());
  const TIMEOUT_MS = 15 * 60 * 1000; // 15 minutes

  useEffect(() => {
    const updateActivity = () => setLastActivity(Date.now());
    const events = ['mousemove', 'keydown', 'click', 'scroll', 'touchstart'];
    
    events.forEach(e => window.addEventListener(e, updateActivity));
    
    const interval = setInterval(() => {
      if (auth.isAuthenticated && Date.now() - lastActivity > TIMEOUT_MS) {
        console.warn('Session expired due to inactivity. Logging out...');
        auth.signoutRedirect();
      }
    }, 60000); // Check every minute

    return () => {
      events.forEach(e => window.removeEventListener(e, updateActivity));
      clearInterval(interval);
    };
  }, [auth, lastActivity]);

  const resetTimeout = () => setLastActivity(Date.now());

  return (
    <AuthTimeoutContext.Provider value={resetTimeout}>
      {children}
    </AuthTimeoutContext.Provider>
  );
};

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  return (
    <OidcProvider {...oidcConfig}>
      <SessionTimeoutWrapper>
        {children}
      </SessionTimeoutWrapper>
    </OidcProvider>
  );
};
