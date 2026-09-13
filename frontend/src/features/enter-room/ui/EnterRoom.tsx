import { useState } from 'react';
import { useNavigate } from 'react-router';
import { useCreateRoomMutation, useJoinRoomMutation } from '@/entities/room';
import { errorMessage } from '@/shared/api';
import { Button } from '@/shared/ui/button';
import styles from './EnterRoom.module.css';

export function EnterRoom({ gameId, initialCode = '' }: { gameId?: string; initialCode?: string }) {
  const navigate = useNavigate();
  const [name, setName] = useState('');
  const [code, setCode] = useState(initialCode);
  const [create, creating] = useCreateRoomMutation();
  const [join, joining] = useJoinRoomMutation();
  const [error, setError] = useState('');
  return <form className={styles.form} onSubmit={async event => {
    event.preventDefault(); setError('');
    try { const result = await (gameId ? create({ name, gameId }) : join({ name, code: code.trim().toUpperCase() })).unwrap(); navigate(`/rooms/${result.code}`); }
    catch (e) { setError(errorMessage(e)); }
  }}>
    <label>Ваше имя<input autoComplete="nickname" value={name} onChange={e => setName(e.target.value)} required minLength={1} maxLength={32} placeholder={gameId ? 'Имя ведущего' : 'Как вас представить?'} /></label>
    {!gameId && <label>Код комнаты<input autoCapitalize="characters" autoComplete="off" value={code} onChange={e => setCode(e.target.value.toUpperCase())} required pattern="[A-Z2-9]{6}" maxLength={6} placeholder="ABC234" /></label>}
    {error && <p role="alert">{error}</p>}
    <Button type="submit" disabled={creating.isLoading || joining.isLoading}>{creating.isLoading || joining.isLoading ? 'Подключаемся…' : gameId ? 'Создать комнату' : 'Присоединиться'}</Button>
    <small>Сохраните вкладку: доступ к вашей роли запоминается в этом браузере.</small>
  </form>;
}
