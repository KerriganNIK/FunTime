import { useState } from 'react';
import { Shield, Sprout, Radiation, Plus, X, Save } from 'lucide-react';
import { coins, planCost, type Country, type Plan, type WorldMatch } from '@/entities/world-domination';
import { Button } from '@/shared/ui/button';

export function WorldOrders({ country, match, command, busy }: { country: Country; match: WorldMatch; command: (kind: string, payload: unknown) => Promise<boolean>; busy: boolean }) {
  const [draft, setDraft] = useState<Plan | null>(null);
  const [target, setTarget] = useState('');
  const [saved, setSaved] = useState(false);
  const plan = draft ?? country.plan!;
  const edit = (change: Partial<Plan>) => { setDraft({ ...plan, ...change }); setSaved(false); };
  const toggle = (key: 'upgrades' | 'shields' | 'sanctions', id: string) => edit({ [key]: plan[key].includes(id) ? plan[key].filter(v => v !== id) : [...plan[key], id] });
  const locked = match.phase !== 'headquarters' || match.paused || country.eliminated || busy;
  const retaliation = country.retaliationRound > 0;
  const conflict = draft !== null && draft.version !== country.plan!.version;
  const left = (country.balanceCents ?? 0) - planCost(plan);
  const others = match.countries.filter(c => c.id !== country.id && !c.eliminated && c.retaliationRound === 0);
  return <section className="game-panel" aria-labelledby="orders-title">
    <div className="game-section-heading"><div><span className="game-kicker">СЕКРЕТНО · ШТАБ СТРАНЫ</span><h2 id="orders-title">{country.name}</h2></div><div className="game-balance"><small>Бюджет</small><strong>{coins(country.balanceCents ?? 0)} <span>монет</span></strong></div></div>
    <p className="game-muted">Доход этой фазы: +{coins(country.incomeCents ?? 0)}. Все участники команды редактируют общий план. Сохранённые решения исполняются в конце фазы штаба.</p>
    {country.sanctionedBy?.length ? <p className="game-notice">Санкции: {country.sanctionedBy.map(id => match.countries.find(c => c.id === id)?.name).join(', ')}. По −15% дохода за страну.</p> : null}
    {retaliation && <p className="game-notice">{country.eliminated ? 'Страна выбыла из игры.' : `Акт возмездия в раунде ${country.retaliationRound}: можно запустить оставшиеся бомбы.`}</p>}
    <fieldset disabled={locked} className="game-fieldset">
      <div className="game-city-grid">{country.cities.map((city, i) => <article key={city.id} className={`game-city ${city.destroyed ? 'game-destroyed' : ''}`}>
        <span className="game-kicker">{i === 0 ? 'СТОЛИЦА' : `ГОРОД ${i + 1}`}{city.shield ? ' · ЩИТ' : ''}</span><h3>{city.name}</h3><strong className="game-city-value">{city.development}<span>%</span></strong><div className="game-meter"><span style={{ width: `${Math.min(city.development / 3, 100)}%` }} /></div>
        {city.destroyed ? <p>Уничтожен · восстановление невозможно</p> : <><label className="game-check"><input type="checkbox" checked={plan.upgrades.includes(city.id)} disabled={retaliation} onChange={() => toggle('upgrades', city.id)} /><span>Улучшить <small>+{city.development < 100 ? 20 : city.development < 200 ? 25 : 100} п.п.</small></span><b>150</b></label><label className="game-check"><input type="checkbox" checked={city.shield || plan.shields.includes(city.id)} disabled={retaliation || city.shield} onChange={() => toggle('shields', city.id)} /><span>Щит</span><b>300</b></label></>}
      </article>)}</div>
      <div className="game-two-col">
        <section className="game-subpanel"><h3><Radiation size={20} /> Ядерная программа</h3><p className="game-muted">В арсенале: <b>{country.bombs ?? 0}</b>. Программа: {country.nuclearRound ? country.nuclearRound > match.round ? `готова в раунде ${country.nuclearRound}` : 'активна' : 'не запущена'}.</p>
          <label className="game-check"><input type="checkbox" checked={plan.nuclear} disabled={retaliation || !!country.nuclearRound} onChange={e => edit({ nuclear: e.target.checked })} /><span>Запустить программу</span><b>500</b></label>
          <label className="game-inline-label">Создать бомбы · 300 за штуку<input type="number" min={0} max={1000} value={plan.bombs} disabled={retaliation || !country.nuclearRound || country.nuclearRound > match.round} onChange={e => edit({ bombs: Math.max(0, Math.min(1000, Math.trunc(Number(e.target.value)))) })} /></label>
          <p className="game-muted">Производство ухудшает экологию. Каждый запуск снижает развитие всех городов мира, даже если сработал щит.</p>
          <label className="game-label">Цель удара<select value={target} onChange={e => setTarget(e.target.value)}><option value="">Выберите город</option>{match.countries.map(c => <optgroup key={c.id} label={c.name}>{c.cities.filter(city => !city.destroyed).map(city => <option key={city.id} value={city.id}>{city.name} · {city.development}%</option>)}</optgroup>)}</select></label>
          <Button variant="secondary" disabled={!target || plan.launches.length >= (country.bombs ?? 0) + plan.bombs} onClick={() => edit({ launches: [...plan.launches, target] })}><Plus size={16} /> Добавить запуск</Button>
          <ul className="game-order-list">{plan.launches.map((id, i) => <li key={`${id}-${i}`}>{match.countries.flatMap(c => c.cities).find(c => c.id === id)?.name}<button aria-label={`Отменить запуск ${i + 1}`} onClick={() => edit({ launches: plan.launches.filter((_, n) => n !== i) })}><X size={16} /></button></li>)}</ul>
        </section>
        <section className="game-subpanel"><h3><Sprout size={20} /> Экология и поддержка</h3><p className="game-muted">Экология мира: <b>{match.ecology}%</b>. Её улучшение увеличивает доход всех стран.</p><label className="game-check"><input type="checkbox" checked={plan.ecology} disabled={retaliation} onChange={e => edit({ ecology: e.target.checked })} /><span>Улучшить экологию</span><b>200</b></label>
          <h4>Анонимные пожертвования</h4><p className="game-muted">Получатель увидит перевод после расчёта, без имени отправителя.</p>{others.map(c => <label key={c.id} className="game-inline-label">{c.name}<input aria-label={`Пожертвовать: ${c.name}`} type="number" min={0} max={1000000000} step="0.01" disabled={retaliation} value={(plan.donations.find(d => d.countryId === c.id)?.amountCents ?? 0) / 100} onChange={e => { const amountCents = Math.max(0, Math.round(Number(e.target.value) * 100)); edit({ donations: [...plan.donations.filter(d => d.countryId !== c.id), ...(amountCents ? [{ countryId: c.id, amountCents }] : [])] }); }} /></label>)}
          <h4><Shield size={17} /> Санкции · бесплатно</h4>{others.map(c => <label key={c.id} className="game-check"><input type="checkbox" checked={plan.sanctions.includes(c.id)} disabled={retaliation} onChange={() => toggle('sanctions', c.id)} /><span>{c.name}</span></label>)}
        </section>
      </div>
    </fieldset>
    <div className="game-save-bar"><div><small>После покупок</small><strong className={left < 0 ? 'game-error' : ''}>{coins(left)} монет</strong><small>{left < 0 ? 'Перерасход: при расчёте случайные покупки будут отменены.' : draft ? 'Есть несохранённые изменения' : saved ? 'План сохранён для всей команды' : 'Показан общий сохранённый план'}</small></div><Button disabled={locked || conflict || !draft} onClick={async () => { if (await command('plan', plan)) { setDraft(null); setSaved(true); } }}><Save size={17} /> Сохранить план</Button></div>
    {conflict && <p className="game-notice" role="alert">Другой участник изменил план. <button onClick={() => setDraft(null)}>Загрузить общий план</button> — ваш черновик будет заменён.</p>}
    {match.phase === 'council' && <p className="game-muted">Сейчас заседание ООН. Покупки откроются в следующей фазе.</p>}
  </section>;
}
