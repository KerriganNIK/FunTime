import { useState } from 'react';
import { MessageCircle } from 'lucide-react';
import type { Country, WorldMatch } from '@/entities/world-domination';
import { Button } from '@/shared/ui/button';

export function WorldDiplomacy({ country, match, command, busy }: { country: Country; match: WorldMatch; command: (kind: string, payload: unknown) => Promise<boolean>; busy: boolean }) {
  const [target, setTarget] = useState(''); const [double, setDouble] = useState(false); const [texts, setTexts] = useState<Record<string, string>>({});
  const locked = busy || match.phase !== 'headquarters' || match.paused || country.retaliationRound > 0;
  const modern = match.rules.version >= 2;
  const capital = country.cities.find(c => c.role === 'capital' && !c.destroyed)?.level ?? 0;
  const name = (id: string) => match.countries.find(c => c.id === id)?.name ?? id;
  return <section className="game-panel"><h2><MessageCircle size={23} /> Дипломатия</h2><p className="game-muted">Один принятый визит за раунд бесплатно. {modern ? `Каждый дополнительный принятый визит в другую страну стоит один уровень столицы. Сейчас доступно: ${capital}.` : 'Один раз за игру можно посетить две страны.'} Отказ не расходует визит или уровень. Обещания не обязывают их выполнять.</p>
    {match.phase === 'council' && <p className="game-notice">Переговоры с другими странами откроются в фазе штаба.</p>}
    {match.phase === 'headquarters' && !country.retaliationRound && <><p className="game-muted">{match.paused ? 'Переговоры приостановлены вместе с игрой.' : 'Отправьте запрос другой стране и дождитесь принятия встречи.'}</p><fieldset className="game-fieldset" disabled={locked}><div className="game-actions"><label className="game-label">Отправить министра<select value={target} onChange={e => setTarget(e.target.value)}><option value="">Выберите страну</option>{match.countries.filter(c => c.id !== country.id && !c.retaliationRound).map(c => <option key={c.id} value={c.id}>{c.name}</option>)}</select></label><Button variant="secondary" disabled={!target} onClick={() => void command('meeting.request', { countryId: target, double })}>Предложить встречу</Button></div><label className="game-check"><input type="checkbox" checked={double} disabled={modern ? capital < 1 : !!country.doubleRound && country.doubleRound !== match.round} onChange={e => setDouble(e.target.checked)} /><span>{modern ? 'Разрешить дополнительный визит · 1 уровень столицы при принятии' : `Использовать право на второй визит ${country.doubleRound ? `(использовано в раунде ${country.doubleRound})` : ''}`}</span></label></fieldset></>}
    <div className="game-meetings">{match.meetings.filter(m => m.from === country.id || m.to === country.id).slice().reverse().map(m => <article key={m.id} className="game-subpanel"><div className="game-section-heading"><h3>{name(m.from)} → {name(m.to)}</h3><span className="game-badge">Раунд {m.round} · {m.status === 'pending' ? 'Ожидает ответа' : m.status === 'accepted' ? 'Встреча принята' : 'Отказ'}</span></div>
      {m.status === 'pending' && m.to === country.id && m.round === match.round && <div className="game-actions"><Button disabled={locked} onClick={() => void command('meeting.respond', { id: m.id, accept: true })}>Принять</Button><Button variant="secondary" disabled={locked} onClick={() => void command('meeting.respond', { id: m.id, accept: false })}>Отказать</Button></div>}
      {m.messages.map((msg, i) => <p key={i} className="game-message"><b>{msg.name} · {name(msg.countryId)}</b><span>{msg.text}</span></p>)}
      {m.status === 'accepted' && m.round === match.round && <form className="game-actions" onSubmit={async e => { e.preventDefault(); if (await command('meeting.message', { id: m.id, text: texts[m.id] ?? '' })) setTexts(v => ({ ...v, [m.id]: '' })); }}><input aria-label={`Сообщение: ${name(m.from)} — ${name(m.to)}`} placeholder="Ваше дипломатическое предложение…" required maxLength={1000} value={texts[m.id] ?? ''} disabled={locked} onChange={e => setTexts(v => ({ ...v, [m.id]: e.target.value }))} /><Button type="submit" disabled={locked || !texts[m.id]?.trim()}>Отправить</Button></form>}
    </article>)}</div>
  </section>;
}
