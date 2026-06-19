const TAIPEI = 'Asia/Taipei';

export function formatTaipeiDateTime(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return new Intl.DateTimeFormat('zh-TW', {
    timeZone: TAIPEI,
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(d);
}

export function formatTaipeiDateTimeLong(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return new Intl.DateTimeFormat('zh-TW', {
    timeZone: TAIPEI,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    weekday: 'short',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(d);
}

export function formatVenueLocalTime(localDate: string): string {
  if (!localDate) return '—';
  return localDate.replace(/^(\d{2})\/(\d{2})\/(\d{4})/, '$1/$2/$3');
}
