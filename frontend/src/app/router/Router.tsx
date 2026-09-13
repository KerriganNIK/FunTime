import { BrowserRouter, Route, Routes } from 'react-router';
import { HomePage } from '@/pages/home';
import { NotFoundPage } from '@/pages/not-found';
import { RoomPage } from '@/pages/room';

export function Router() {
  return <BrowserRouter><Routes><Route path="/" element={<HomePage/>}/><Route path="/rooms/:code" element={<RoomPage/>}/><Route path="/rooms/:code/screen" element={<RoomPage screen/>}/><Route path="*" element={<NotFoundPage/>}/></Routes></BrowserRouter>;
}
