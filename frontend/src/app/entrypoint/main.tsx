import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { Router } from '../router/Router';
import { StoreProvider } from '../providers/StoreProvider';
import '../styles/global.css';
import '@/shared/styles/game.css';

createRoot(document.getElementById('root')!).render(<StrictMode><StoreProvider><Router/></StoreProvider></StrictMode>);
