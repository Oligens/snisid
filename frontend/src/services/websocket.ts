import { useEffect } from 'react';
import { create } from 'zustand';
import { useAuth } from 'react-oidc-context';
import { useApplicationAuth } from '../auth/AuthProvider';


type WebSocketState = {
  socket: WebSocket | null;
  isConnected: boolean;
  metrics: {
    totalIdentities: number;
    verifiedIdentities: number;
    suspiciousActivities: number;
    activeAlerts: number;
    cpuLoad: number;
  };
  connect: (token: string) => void;
  disconnect: () => void;
};

export const useWebSocketStore = create<WebSocketState>((set, get) => ({
  socket: null,
  isConnected: false,
  metrics: {
    totalIdentities: 0,
    verifiedIdentities: 0,
    suspiciousActivities: 0,
    activeAlerts: 0,
    cpuLoad: 0
  },
  
  connect: (token: string) => {
    const { socket } = get();
    if (socket?.readyState === WebSocket.OPEN) return;

    // Use wss:// in production, ws:// in dev
    const wsUrl = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/events.risk';
    
    // Pass token as subprotocol or query param depending on backend support
    // (Query param is generally less secure due to access logs, subprotocol is preferred)
    const newSocket = new WebSocket(wsUrl, ['access_token', token]);

    newSocket.onopen = () => {
      console.log('WebSocket Connected');
      set({ isConnected: true });
    };

    newSocket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        // Assuming backend sends a 'metrics_update' event
        if (data.type === 'metrics_update') {
          set(state => ({
            metrics: { ...state.metrics, ...data.payload }
          }));
        }
      } catch (e) {
        console.error('Failed to parse WebSocket message', e);
      }
    };

    newSocket.onclose = (event) => {
      console.log('WebSocket Disconnected', event.reason);
      set({ isConnected: false, socket: null });
      
      // Exponential backoff reconnection could be implemented here
      if (!event.wasClean) {
        setTimeout(() => get().connect(token), 5000);
      }
    };

    newSocket.onerror = (error) => {
      console.error('WebSocket Error:', error);
      // Close will be fired automatically
    };

    set({ socket: newSocket });
  },

  disconnect: () => {
    const { socket } = get();
    if (socket) {
      socket.close(1000, 'User logged out');
      set({ socket: null, isConnected: false });
    }
  }
}));

// Hook to automatically manage WebSocket connection lifecycle
export const useWebSocketConnection = () => {
  const auth = useApplicationAuth();
  const { connect, disconnect, isConnected, metrics } = useWebSocketStore();

  useEffect(() => {
    if (!auth.isDemo && auth.isAuthenticated && auth.user?.access_token) {
      connect(auth.user.access_token);
    } else {
      disconnect();
    }

    return () => {
      // Don't disconnect on unmount, let the store manage the singleton connection
      // unless we specifically want to kill it when the dashboard unmounts.
    };
  }, [auth.isAuthenticated, auth.user?.access_token, connect, disconnect]);

  return { isConnected, metrics };
};
