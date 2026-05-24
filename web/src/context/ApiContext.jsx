import { createContext, useContext } from 'react';
import { useAuth } from './AuthContext';

const ApiContext = createContext(null);

const API_BASE = '/api/v1';

export function ApiProvider({ children }) {
  const { token } = useAuth();

  const apiFetch = async (path, options = {}) => {
    const headers = {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options.headers,
    };
    const res = await fetch(`${API_BASE}${path}`, { ...options, headers });
    return res;
  };

  const get = (path) => apiFetch(path);
  const post = (path, body) => apiFetch(path, { method: 'POST', body: JSON.stringify(body) });
  const put = (path, body) => apiFetch(path, { method: 'PUT', body: JSON.stringify(body) });
  const patch = (path, body) => apiFetch(path, { method: 'PATCH', body: JSON.stringify(body) });
  const del = (path) => apiFetch(path, { method: 'DELETE' });

  return (
    <ApiContext.Provider value={{ apiFetch, get, post, put, patch, del }}>
      {children}
    </ApiContext.Provider>
  );
}

export function useApi() {
  const ctx = useContext(ApiContext);
  if (!ctx) throw new Error('useApi must be used within ApiProvider');
  return ctx;
}
