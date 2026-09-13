import { Asterisk, MousePointer2, Smile, Sparkles } from 'lucide-react';
import styles from './HomePage.module.css';

export function HeroArtwork() {
  return <div className={styles.heroArt} aria-hidden="true">
    <div className={styles.orbitOne} /><div className={styles.orbitTwo} />
    <span className={styles.artCaption}>ХОРОШО, КОГДА ВСЕ В СБОРЕ</span>
    <div className={`${styles.sticker} ${styles.limeSticker}`}><Smile size={75} strokeWidth={1.4}/></div>
    <div className={`${styles.sticker} ${styles.purpleSticker}`}><Asterisk size={86} strokeWidth={1.5}/></div>
    <div className={`${styles.sticker} ${styles.peachSticker}`}><span>ft.</span></div>
    <Sparkles className={styles.artSpark} size={32} strokeWidth={1.2}/>
    <div className={styles.cursor}><MousePointer2 size={23} fill="#303729"/><span>твоя компания</span></div>
    <span className={styles.artFootnote}>МЕНЬШЕ СКРОЛЛА. БОЛЬШЕ ИГР.</span>
  </div>;
}
