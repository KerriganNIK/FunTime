import { Link } from 'react-router';
import { ArrowLeft } from 'lucide-react';
import styles from './NotFoundPage.module.css';

export function NotFoundPage() {
  return <main className={styles.page}><span>404</span><h1>Кажется, мы вышли за карту.</h1><p>Такой страницы нет. Вернёмся туда, где собирается компания?</p><Link to="/"><ArrowLeft size={18}/> На главную FunTime</Link></main>;
}
