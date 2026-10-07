import { useRef, useState } from 'react';
import { Copy, Share2, Smartphone } from 'lucide-react';
import { copyText, roomInviteUrl } from '@/shared/lib/invite';
import { Button } from '@/shared/ui/button';
import { Dialog } from '@/shared/ui/dialog';
import styles from './InviteRoom.module.css';

export function InviteRoom({ code }: { code: string }) {
  const [open, setOpen] = useState(false);
  const [message, setMessage] = useState('');
  const input = useRef<HTMLInputElement>(null);
  const url = roomInviteUrl(code);
  const isLocal = import.meta.env.DEV && new URL(url).hostname !== location.hostname;
  return <>
    <button aria-label="Пригласить игроков" onClick={() => { setMessage(''); setOpen(true); }}><Smartphone size={17} /> Пригласить</button>
    <Dialog open={open} onClose={() => setOpen(false)} title="Пригласить игроков">
      <p className={styles.description}>Откройте ссылку на телефоне и выберите команду. {isLocal && 'Для локальной игры подключите телефон и компьютер к одной Wi-Fi сети.'}</p>
      <div className={styles.code}><span>КОД КОМНАТЫ</span><strong>{code}</strong></div>
      <label className={styles.link}>Ссылка для игроков<input ref={input} readOnly value={url} onFocus={e => e.target.select()} /></label>
      <div className={styles.actions}><Button onClick={async () => { if (await copyText(url)) setMessage('Ссылка скопирована'); else { input.current?.focus(); input.current?.select(); setMessage('Ссылка выделена — скопируйте её через меню браузера'); } }}><Copy size={17} /> Скопировать ссылку</Button>
        {typeof navigator.share === 'function' && <Button variant="secondary" onClick={async () => { try { await navigator.share({ title: 'FunTime', text: `Присоединяйся к комнате ${code}`, url }); } catch (error) { if (!(error instanceof DOMException && error.name === 'AbortError')) setMessage('Отправьте ссылку с помощью кнопки копирования'); } }}><Share2 size={17} /> Поделиться</Button>}
      </div>
      <p className={styles.feedback} role="status">{message}</p>
    </Dialog>
  </>;
}
