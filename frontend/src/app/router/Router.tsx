import { BrowserRouter, Route, Routes } from 'react-router';
import { HomePage } from '@/pages/home';
import { NotFoundPage } from '@/pages/not-found';

export function Router() {
  return <BrowserRouter><Routes><Route path="/" element={<HomePage/>}/><Route path="*" element={<NotFoundPage/>}/></Routes></BrowserRouter>;
}
