import { Moon, Sun } from 'lucide-react';
import { useTheme } from '@/shared/lib/theme';
import styles from './SwitchTheme.module.css';

export function SwitchTheme() {
  const { theme, setTheme } = useTheme();
  const label = theme === 'light' ? 'Включить тёмную тему' : 'Включить светлую тему';
  return <button className={styles.toggle} aria-label={label} title={label} onClick={() => setTheme(theme === 'light' ? 'dark' : 'light')}>
    {theme === 'light' ? <Moon size={19} /> : <Sun size={19} />}
  </button>;
}
