import { useId, useState } from 'react';
import { Shield, Sprout, Radiation, Plus, X, Save, ScanEye, LockKeyhole, Play, ArrowRight } from 'lucide-react';
import { coins, planCost, cityRoles, CountryFlag, CityArtwork, type Country, type Plan, type WorldMatch } from '@/entities/world-domination';
import { Button } from '@/shared/ui/button';

export function WorldOrders({ country, match, command, busy, host = false }: { country: Country; match: WorldMatch; command: (kind: string, payload: unknown) => Promise<boolean>; busy: boolean; host?: boolean }) {
  const [draft, setDraft] = useState<Plan | null>(null);
  const [target, setTarget] = useState('');
  const [saved, setSaved] = useState(false);
  const [section, setSection] = useState('cities');
  const sectionId = useId();
  const plan = draft ?? country.plan!;
  const edit = (change: Partial<Plan>) => { setDraft({ ...plan, ...change }); setSaved(false); };
  const toggle = (key: 'upgrades' | 'shields' | 'sanctions' | 'spies', id: string) => { const list = plan[key] ?? []; edit({ [key]: list.includes(id) ? list.filter(v => v !== id) : [...list, id] }); };
  const locked = match.phase !== 'headquarters' || match.paused || country.eliminated || busy;
  const retaliation = country.retaliationRound > 0;
  const conflict = draft !== null && draft.version !== country.plan!.version;
  const prices = country.prices!;
  const modern = match.rules.version >= 2;
  const cost = planCost(plan, prices);
  const left = (country.balanceCents ?? 0) - cost;
  const expectedPhase = `${match.round}:${match.phase}`;
  const intelligenceCity = country.cities.find(c => c.role === 'intelligence' && !c.destroyed);
  const spyCapacity = (intelligenceCity?.level ?? 0) + Number(!!intelligenceCity && plan.upgrades.includes(intelligenceCity.id));
  const others = match.countries.filter(c => c.id !== country.id && !c.eliminated && c.retaliationRound === 0);
  return <section className="game-panel" aria-labelledby="orders-title">
    <div className="game-section-heading"><div><span className="game-kicker">СЕКРЕТНО · ШТАБ СТРАНЫ</span><h2 id="orders-title"><CountryFlag country={country.id} /> {country.name}</h2></div><div className="game-balance"><small>ДОСТУПНЫЙ БЮДЖЕТ</small><strong>{coins(country.balanceCents ?? 0)} <span>монет</span></strong></div></div>
    <div className="game-economy"><span>Развитие <b>{country.score}%</b></span><span>Последний доход <b>+{coins(country.incomeCents ?? 0)}</b></span><span>Арсенал <b>{country.bombs ?? 0} бомб</b></span></div>
    {!country.eliminated && (match.phase === 'council' || match.paused) && <div className="game-action-notice" role="status">
      <LockKeyhole size={23} aria-hidden="true" />
      <div><h3>{match.paused ? 'Игра на паузе' : 'Покупки откроются в штабе'}</h3><p>{match.paused
        ? host ? 'Возобновите игру, чтобы участники могли принимать решения.' : 'Ведущий приостановил игру. После возобновления можно продолжить работу.'
        : host ? 'Сейчас заседание ООН: команды обсуждают стратегию. Когда обсуждение закончится, откройте штабы — станут доступны покупки и выборы.' : 'Сейчас заседание ООН. Покупки и переговоры откроются, когда ведущий перейдёт к фазе штаба или закончится таймер.'}</p></div>
      {host && <Button disabled={busy} onClick={() => void command(match.paused ? 'resume' : 'advance', { expectedPhase })}>{match.paused ? <Play size={17} /> : <ArrowRight size={17} />}{match.paused ? 'Снять паузу' : 'Перейти к покупкам'}</Button>}
    </div>}
    {match.phase === 'headquarters' && !country.eliminated && <p className="game-muted">Выберите покупки, затем нажмите «Сохранить план». Сумма расходов меняется сразу; деньги спишутся и улучшения применятся при расчёте раунда ведущим.{modern && ' Доход следующего раунда начислится после расчёта. Скидки учитывают уровни городов до улучшений.'}</p>}
    {country.sanctionedBy?.length ? <p className="game-notice">Санкции: {country.sanctionedBy.map(id => match.countries.find(c => c.id === id)?.name).join(', ')}. По −15% дохода за страну.</p> : null}
    {retaliation && <p className="game-notice">{country.eliminated ? 'Страна выбыла из игры.' : `Акт возмездия в раунде ${country.retaliationRound}: можно запустить оставшиеся бомбы.`}</p>}
    {match.phase === 'council' && <div className="game-city-overview" aria-label="Города вашей страны">{country.cities.map(city => <div key={city.id}><span>{city.name}</span><strong>{city.development}%</strong><small>{city.destroyed ? 'Уничтожен' : cityRoles[city.role ?? '']?.name ?? 'Город'}</small></div>)}</div>}
    {match.phase === 'headquarters' && !country.eliminated && <>
    <div className="game-purchase-preview" aria-live="polite" aria-atomic="true"><div><small>В плане · к оплате</small><strong>{coins(cost)} монет</strong></div><div><small>Останется после покупок</small><strong className={left < 0 ? 'game-error' : ''}>{coins(left)} монет</strong></div></div>
    <nav className="game-order-tabs" aria-label="Разделы плана">{[['cities','Города'], ['defense','Оборона'], ['support','Экономика'], ...(modern ? [['intelligence','Разведка']] : [])].map(([id,label]) => <button key={id} type="button" aria-pressed={section === id} aria-controls={`${sectionId}-${id}`} onClick={() => setSection(id)}>{label}</button>)}</nav>
    <fieldset disabled={locked} className="game-fieldset">
      <div id={`${sectionId}-cities`} data-order-section="cities" data-active={section === 'cities'} className="game-city-grid">{country.cities.map((city, i) => <article key={city.id} className={`game-city game-city-${cityRoles[city.role ?? '']?.color ?? 'violet'} ${city.destroyed ? 'game-destroyed' : ''}`}>
        <span className="game-kicker">{cityRoles[city.role ?? '']?.name ?? (i === 0 ? 'СТОЛИЦА' : `ГОРОД ${i + 1}`)}</span><CityArtwork role={city.role} /><h3>{city.name}</h3>
        <div className="game-city-metrics"><strong className="game-city-value">{city.development}<span>%</span></strong>{modern && <span className="game-level">Ур. {city.level ?? 0}{plan.upgrades.includes(city.id) ? ' → +1' : ''}</span>}</div><div className="game-meter"><span style={{ width: `${Math.min(city.development / 3, 100)}%` }} /></div>
        {modern && <p className="game-city-benefit">{cityRoles[city.role ?? '']?.benefit}</p>}
        {city.destroyed ? <p>Уничтожен · восстановление невозможно</p> : <><label className="game-check"><input type="checkbox" checked={plan.upgrades.includes(city.id)} disabled={retaliation} onChange={() => toggle('upgrades', city.id)} /><span>Улучшить <small>+{city.development < 100 ? 20 : city.development < 200 ? 25 : 100} п.п.{modern ? ' · +1 уровень' : ''}</small></span><b>{coins(prices.upgrade)}</b></label><label className="game-check"><input type="checkbox" checked={city.shield || plan.shields.includes(city.id)} disabled={retaliation || city.shield} onChange={() => toggle('shields', city.id)} /><span>Щит{modern ? ' · купол' : ''}<small>{city.shield ? 'Уже установлен' : 'Поглощает один удар'}</small></span><b>{coins(prices.shield)}</b></label></>}
      </article>)}</div>
      <div className="game-two-col">
        <section id={`${sectionId}-defense`} data-order-section="defense" data-active={section === 'defense'} className="game-subpanel"><h3><Radiation size={20} /> Ядерная программа</h3><p className="game-muted">В арсенале: <b>{country.bombs ?? 0}</b>. Программа: {country.nuclearRound ? country.nuclearRound > match.round ? `готова в раунде ${country.nuclearRound}` : 'активна' : 'не запущена'}.</p>
          <label className="game-check"><input type="checkbox" checked={plan.nuclear} disabled={retaliation || !!country.nuclearRound} onChange={e => edit({ nuclear: e.target.checked })} /><span>Запустить программу</span><b>500</b></label>
          <label className="game-inline-label">Создать бомбы · {coins(prices.bomb)} за штуку<input type="number" min={0} max={1000} value={plan.bombs} disabled={retaliation || !country.nuclearRound || country.nuclearRound > match.round} onChange={e => edit({ bombs: Math.max(0, Math.min(1000, Math.trunc(Number(e.target.value)))) })} /></label>
          <p className="game-muted">Производство ухудшает экологию. Каждый запуск снижает развитие всех городов мира, даже если сработал щит.</p>
          <label className="game-label">Цель удара<select value={target} onChange={e => setTarget(e.target.value)}><option value="">Выберите город</option>{match.countries.map(c => <optgroup key={c.id} label={c.name}>{c.cities.filter(city => !city.destroyed).map(city => <option key={city.id} value={city.id}>{city.name} · {city.development}%</option>)}</optgroup>)}</select></label>
          <Button variant="secondary" disabled={!target || plan.launches.length >= (country.bombs ?? 0) + plan.bombs} onClick={() => edit({ launches: [...plan.launches, target] })}><Plus size={16} /> Добавить запуск</Button>
          <ul className="game-order-list">{plan.launches.map((id, i) => <li key={`${id}-${i}`}>{match.countries.flatMap(c => c.cities).find(c => c.id === id)?.name}<button aria-label={`Отменить запуск ${i + 1}`} onClick={() => edit({ launches: plan.launches.filter((_, n) => n !== i) })}><X size={16} /></button></li>)}</ul>
        </section>
        <section id={`${sectionId}-support`} data-order-section="support" data-active={section === 'support'} className="game-subpanel"><h3><Sprout size={20} /> Экология и поддержка</h3><p className="game-muted">{modern ? <>Загрязнение мира: <b>{match.pollution}%</b>. Доход: {coins(match.baseIncomeCents)} × {Math.max(0, 100-match.pollution)}% + туризм. Затем −15% за каждую страну с санкциями.</> : <>Экология мира: <b>{match.ecology}%</b>.</>} Улучшение помогает всем странам.</p><label className="game-check"><input type="checkbox" checked={plan.ecology} disabled={retaliation} onChange={e => edit({ ecology: e.target.checked })} /><span>Улучшить экологию</span><b>{coins(prices.ecology)}</b></label>
          <h4>Анонимные пожертвования</h4><p className="game-muted">Получатель увидит перевод после расчёта, без имени отправителя.</p>{others.map(c => <label key={c.id} className="game-inline-label">{c.name}<input aria-label={`Пожертвовать: ${c.name}`} type="number" min={0} max={1000000000} step="0.01" disabled={retaliation} value={(plan.donations.find(d => d.countryId === c.id)?.amountCents ?? 0) / 100} onChange={e => { const amountCents = Math.max(0, Math.round(Number(e.target.value) * 100)); edit({ donations: [...plan.donations.filter(d => d.countryId !== c.id), ...(amountCents ? [{ countryId: c.id, amountCents }] : [])] }); }} /></label>)}
          <h4><Shield size={17} /> Санкции · бесплатно</h4>{others.map(c => <label key={c.id} className="game-check"><input type="checkbox" checked={plan.sanctions.includes(c.id)} disabled={retaliation} onChange={() => toggle('sanctions', c.id)} /><span>{c.name}</span></label>)}
        </section>
      </div>
      {modern && <section id={`${sectionId}-intelligence`} data-order-section="intelligence" data-active={section === 'intelligence'} className="game-subpanel game-intelligence"><div className="game-section-heading"><h3><ScanEye size={21} /> Разведка</h3><span className="game-badge">Доступно: {spyCapacity} уровней</span></div><p className="game-muted">Один уровень за страну: после расчёта получите снимок её бюджета, бомб, программы, уровней и куполов. Можно использовать уровень от запланированного улучшения. При отмене покупки разведка тоже отменится.</p><div className="game-spy-options">{others.map(c => <label key={c.id} className="game-check"><input type="checkbox" checked={(plan.spies ?? []).includes(c.id)} disabled={retaliation || !(plan.spies ?? []).includes(c.id) && (plan.spies ?? []).length >= spyCapacity} onChange={() => toggle('spies', c.id)} /><span>{c.name}</span><small>1 уровень</small></label>)}</div></section>}
    </fieldset>
    <div className="game-save-bar"><div><small>Остаток после покупок</small><strong className={left < 0 ? 'game-error' : ''}>{coins(left)} монет</strong></div><Button disabled={locked || conflict || !draft} onClick={async () => { if (await command('plan', plan)) { setDraft(null); setSaved(true); } }}><Save size={17} /> Сохранить план</Button></div>
    <div className="game-save-feedback" role="status"><p>{left < 0 ? 'Перерасход: при расчёте случайные покупки будут отменены.' : draft ? 'Есть несохранённые изменения' : saved ? 'План сохранён для всей команды' : 'Показан общий сохранённый план'}</p>{!draft && cost > 0 && <p>Ожидаем расчёта раунда. Покупки ещё не списаны.</p>}</div>
    {conflict && <p className="game-notice" role="alert">Другой участник изменил план. <button onClick={() => setDraft(null)}>Загрузить общий план</button> — ваш черновик будет заменён.</p>}
    </>}
    {!!country.intelligence?.length && <div className="game-intelligence"><h3><ScanEye size={20} /> Закрытые отчёты</h3><p className="game-muted">Данные на момент окончания указанного раунда. Доступны только вашей команде и ведущему.</p>{country.intelligence.slice().reverse().map((report, i) => <details key={`${report.round}:${report.countryId}:${i}`} className="game-subpanel"><summary>{report.countryName} · после раунда {report.round}</summary><div className="game-economy"><span>Бюджет <b>{coins(report.balanceCents)}</b></span><span>Бомбы <b>{report.bombs}</b></span><span>Программа <b>{report.nuclearRound ? `с раунда ${report.nuclearRound}` : 'нет'}</b></span></div>{report.cities.map(c => <div className="game-city-row" key={c.id}><span>{c.name} · ур. {c.level ?? 0}</span><b>{c.destroyed ? 'Уничтожен' : c.shield ? 'Купол установлен' : 'Без купола'}</b></div>)}</details>)}</div>}
  </section>;
}
