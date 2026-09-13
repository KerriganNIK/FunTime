import { useState } from 'react';
import { ArrowDown, ArrowRight, ArrowUpRight, Asterisk, CircleHelp, Gamepad2, Globe2, Monitor, Plus, Search, Smartphone, Sparkles, Users, X } from 'lucide-react';
import { AppHeader } from '@/widgets/app-header';
import { GameCard, type Game } from '@/entities/game';
import { Button } from '@/shared/ui/button';
import { Dialog } from '@/shared/ui/dialog';
import { EnterRoom } from '@/features/enter-room';
import { useCatalog } from '../model/use-catalog';
import { PlanetArtwork } from './PlanetArtwork';
import { HeroArtwork } from './HeroArtwork';
import styles from './HomePage.module.css';

export function HomePage() {
  const { state, retry } = useCatalog();
  const [search, setSearch] = useState('');
  const [filter, setFilter] = useState<'all' | 'coming_soon'>('all');
  const [selected, setSelected] = useState<Game | null>(null);
  const [joinOpen, setJoinOpen] = useState(false);
  const games = state.status === 'ready' ? state.games : [];
  const visible = games.filter(game => (filter === 'all' || game.status === filter) && game.title.toLocaleLowerCase('ru').includes(search.trim().toLocaleLowerCase('ru')));
  const resetFilters = () => { setSearch(''); setFilter('all'); };

  return <>
    <AppHeader onJoin={() => setJoinOpen(true)} />
    <main id="main" className={styles.main}>
      <section className={styles.hero} aria-labelledby="hero-title">
        <div className={styles.heroCopy}>
          <div className={styles.eyebrow}><span /> ВРЕМЯ БЫТЬ ВМЕСТЕ</div>
          <h1 id="hero-title">Твоя компания.<br/>Ваши <span className={styles.titleAccent}>правила<Asterisk aria-hidden="true" /></span>.</h1>
          <p>Собирайте друзей. Выбирайте игру.<br/>Превращайте обычный вечер в вашу историю.</p>
          <div className={styles.heroActions}><a className={styles.primaryLink} href="#games">Выбрать игру <ArrowDown size={18}/></a><a className={styles.secondaryLink} href="#how-it-works">Как это работает <ArrowUpRight size={17}/></a></div>
          <div className={styles.heroNote}><Globe2 size={15}/> Прямо в браузере <span>·</span> Вместе, даже на расстоянии</div>
        </div>
        <HeroArtwork />
      </section>

      <section id="games" className={styles.catalog} aria-labelledby="games-title">
        <div className={styles.sectionTop}>
          <div><div className={styles.overline}>КОЛЛЕКЦИЯ FUNTIME</div><h2 id="games-title">Во что сыграем?<span className={styles.count}>{state.status === 'ready' ? String(games.length).padStart(2, '0') : '—'}</span></h2></div>
          <label className={styles.search}><Search size={17}/><input type="search" placeholder="Найти игру" aria-label="Найти игру" value={search} onChange={event => setSearch(event.target.value)}/>{search && <button onClick={() => setSearch('')} aria-label="Очистить поиск"><X size={15}/></button>}</label>
        </div>
        <div className={styles.filterRow}><div className={styles.filters} aria-label="Фильтр игр"><button aria-pressed={filter === 'all'} onClick={() => setFilter('all')}>Все игры</button><button aria-pressed={filter === 'coming_soon'} onClick={() => setFilter('coming_soon')}><Sparkles size={14}/> Скоро</button></div><span className={styles.collectionNote}>Начинаем с большой идеи</span></div>

        {state.status === 'loading' && <div className={styles.loading} role="status"><div className={styles.skeleton}/><span>Собираем игры для вашей компании…</span></div>}
        {state.status === 'error' && <div className={styles.empty} role="alert"><CircleHelp size={32}/><h3>Каталог немного задерживается</h3><p>Не удалось загрузить игры. Попробуйте ещё раз.</p><Button variant="secondary" onClick={retry}>Попробовать снова <ArrowRight size={17}/></Button></div>}
        {state.status === 'ready' && visible.length === 0 && <div className={styles.empty}><Search size={30}/><h3>{games.length === 0 ? 'Коллекция скоро появится' : 'Такой игры пока нет'}</h3><p>{games.length === 0 ? 'Загляните чуть позже — мы готовим первые игры.' : 'Попробуйте другое название или посмотрите все игры.'}</p>{games.length > 0 && <Button variant="secondary" onClick={resetFilters}>Показать все игры</Button>}</div>}
        {state.status === 'ready' && visible.length > 0 && <div className={styles.gameGrid}>
          {visible.map(game => <GameCard key={game.id} game={game} onDetails={() => setSelected(game)} artwork={game.id === 'world-domination' ? <PlanetArtwork/> : <Gamepad2 aria-hidden="true"/>}/>)}
          {!search.trim() && filter === 'all' && <article className={styles.nextCard}>
            <div className={styles.nextArt} aria-hidden="true"><div className={styles.nextShape}><Asterisk size={92} strokeWidth={1.1}/></div><span className={styles.nextPlus}><Plus size={26}/></span><span className={styles.dottedOrbit}/></div>
            <div className={styles.nextCopy}><span className={styles.smallBadge}>ПРОДОЛЖЕНИЕ СЛЕДУЕТ</span><h3>У веселья<br/>большие планы.</h3><p>Новые игры, новые поводы собраться.<br/>Наша коллекция только начинается.</p><span className={styles.nextFooter}>Дальше — больше <Sparkles size={16}/></span></div>
          </article>}
        </div>}
      </section>

      <section id="how-it-works" className={styles.how} aria-labelledby="how-title">
        <div className={styles.howIntro}><div className={styles.overline}>ВСЁ ПРОСТО</div><h2 id="how-title">Один вечер.<br/>Три простых шага.</h2><p>От выбора игры<br/>до первого раунда.</p></div>
        <div className={styles.steps}>
          <div className={styles.step}><span className={styles.stepIcon}><Monitor size={22}/></span><span className={styles.stepNumber}>01</span><h3>Выберите игру</h3><p>Откройте FunTime на большом экране и создайте комнату.</p></div>
          <div className={styles.step}><span className={styles.stepIcon}><Smartphone size={22}/></span><span className={styles.stepNumber}>02</span><h3>Соберите своих</h3><p>Поделитесь кодом. Друзья смогут присоединиться со своих устройств.</p></div>
          <div className={styles.step}><span className={styles.stepIcon}><Users size={22}/></span><span className={styles.stepNumber}>03</span><h3>Пусть будет весело</h3><p>Вы вместе, игра на экране. Всё остальное — уже ваша история.</p></div>
        </div>
      </section>

      <section className={styles.faq} aria-label="Вопросы о FunTime"><span><CircleHelp size={18}/> На всякий случай</span><details><summary>Нужно что-нибудь устанавливать?<Plus size={17}/></summary><p>FunTime работает в браузере. Для игры понадобятся интернет и устройства, с которых удобно присоединиться к комнате.</p></details><details><summary>Когда можно будет сыграть?<Plus size={17}/></summary><p>«Мировое господство» уже доступно. Создайте комнату, пригласите друзей по коду и распределите страны. Для начала нужны две команды хотя бы по одному игроку.</p></details></section>
    </main>
    <footer className={styles.footer}><span className={styles.footerBrand}><Asterisk size={20}/> funtime <span>Хорошее время — общее.</span></span><span>Сделано для ваших вечеров <span className={styles.footerStar}>✳</span></span></footer>

    <Dialog open={joinOpen} onClose={() => setJoinOpen(false)} title="Присоединиться к друзьям">
      <div className={styles.dialogIcon}><Users size={28}/></div><p className={styles.dialogText}>Введите код от ведущего и займите место в своей команде.</p><EnterRoom />
    </Dialog>
    <Dialog open={selected !== null} onClose={() => setSelected(null)} title={selected?.title || 'Об игре'}>
      {selected && <><span className={styles.modalBadge}>{selected.status === 'available' ? 'Готова к игре' : 'Скоро в FunTime'}</span><p className={styles.dialogText}>{selected.description}</p>{selected.status === 'available' ? <><p className={styles.dialogText}>2–10 стран, по 1–5 участников. Шесть раундов с двумя фазами по 12 минут. Ведущий управляет темпом, игроки принимают решения со своих устройств.</p><EnterRoom gameId={selected.id} /></> : <p className={styles.dialogText}>Эта игра пока готовится к запуску.</p>}</>}
    </Dialog>
  </>;
}
