import { ArrowUpRight, Sparkles } from 'lucide-react';
import type { ReactNode } from 'react';
import type { Game } from '../model/game';
import styles from './GameCard.module.css';

const statusLabels: Record<Game['status'], string> = { coming_soon: 'Скоро в FunTime', available: 'Доступна', maintenance: 'На обновлении' };

export function GameCard({ game, artwork, onDetails }: { game: Game; artwork: ReactNode; onDetails: () => void }) {
  return <article className={`${styles.card} ${styles[game.accent]}`}>
    <button className={styles.art} onClick={onDetails} aria-label={`Подробнее об игре «${game.title}»`}>
      <span className={styles.badge}><Sparkles size={13} />{statusLabels[game.status]}</span>
      {artwork}
      <span className={styles.artLabel}>ПЛАНЕТА ОДНА. АМБИЦИЙ МНОГО.</span>
      <span className={styles.roundArrow}><ArrowUpRight size={21} /></span>
    </button>
    <div className={styles.content}>
      <span className={styles.category}>{game.category}</span>
      <h3>{game.title}</h3>
      <p>{game.description}</p>
      <div className={styles.bottom}><span className={styles.status}><span />{game.status === 'coming_soon' ? 'Готовим к запуску' : statusLabels[game.status]}</span><button onClick={onDetails}>Об игре <ArrowUpRight size={17} /></button></div>
    </div>
  </article>;
}
