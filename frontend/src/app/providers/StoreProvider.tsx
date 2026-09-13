import { useEffect, type ReactNode } from 'react';
import { Provider } from 'react-redux';
import { setupListeners } from '@reduxjs/toolkit/query';
import { store } from '../store/store';

export function StoreProvider({ children }: { children: ReactNode }) {
  useEffect(() => setupListeners(store.dispatch), []);
  return <Provider store={store}>{children}</Provider>;
}
