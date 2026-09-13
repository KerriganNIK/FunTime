import { useEffect, useId, useRef, type ReactNode } from 'react';
import { X } from 'lucide-react';
import styles from './Dialog.module.css';

type Props = { open: boolean; onClose: () => void; title: string; children: ReactNode };

export function Dialog({ open, onClose, title, children }: Props) {
  const ref = useRef<HTMLDialogElement>(null);
  const titleId = useId();
  useEffect(() => {
    const dialog = ref.current;
    if (!open || !dialog) return;
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    dialog.showModal();
    return () => { dialog.close(); document.body.style.overflow = previousOverflow; };
  }, [open]);
  return (
    <dialog ref={ref} className={styles.dialog} aria-labelledby={titleId}
      onCancel={onClose} onClick={event => { if (event.target === event.currentTarget) onClose(); }}>
      <div className={styles.content}>
        <button autoFocus className={styles.close} onClick={onClose} aria-label="Закрыть окно"><X size={21} /></button>
        <h2 id={titleId}>{title}</h2>
        {children}
      </div>
    </dialog>
  );
}
