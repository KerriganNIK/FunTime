import { ArrowUpRight, Asterisk } from 'lucide-react';
import { Link } from 'react-router';
import { SwitchTheme } from '@/features/switch-theme';
import styles from './AppHeader.module.css';

export function AppHeader({ onJoin }: { onJoin: () => void }) {
  return <header className={styles.header}>
    <a className={styles.skip} href="#main">Перейти к содержимому</a>
    <Link to="/" aria-label="FunTime — главная" className={styles.logo}><span className={styles.mark}><Asterisk size={30} strokeWidth={2.8} /></span>fun<span className={styles.time}>time</span><span className={styles.dot}>✳</span></Link>
    <nav className={styles.nav} aria-label="Основная навигация">
      <a className={styles.active} href="/#games">Игры<span /></a>
      <a href="/#how-it-works">Как это работает</a>
    </nav>
    <div className={styles.actions}><SwitchTheme /><button className={styles.join} onClick={onJoin}>Войти по коду <ArrowUpRight size={17} /></button></div>
  </header>;
}
