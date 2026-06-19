import React, { useEffect, useState } from 'react';
import { UserCircle2 } from 'lucide-react';
import { cn } from '../lib/utils';
import { fetchWikiThumbnail } from '../services/api';

interface PlayerPhotoProps {
  wikiSlug: string;
  name: string;
  size?: 'sm' | 'lg';
  className?: string;
}

export function PlayerPhoto({ wikiSlug, name, size = 'sm', className }: PlayerPhotoProps) {
  const [url, setUrl] = useState<string | null>(null);
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    let cancelled = false;
    setUrl(null);
    setFailed(false);

    fetchWikiThumbnail(wikiSlug)
      .then((data) => {
        if (!cancelled && data.thumbnail_url) {
          setUrl(data.thumbnail_url);
        }
      })
      .catch(() => {
        if (!cancelled) setFailed(true);
      });

    return () => {
      cancelled = true;
    };
  }, [wikiSlug]);

  const dim = size === 'lg' ? 'w-16 h-16' : 'w-8 h-8';
  const icon = size === 'lg' ? 'w-8 h-8' : 'w-4 h-4';

  if (url && !failed) {
    return (
      <img
        src={url}
        alt={name}
        className={cn(dim, 'rounded-sm object-cover border border-[#333] bg-[#1a1a1a]', className)}
        onError={() => setFailed(true)}
      />
    );
  }

  return (
    <div className={cn(dim, 'bg-[#1a1a1a] border border-[#333] flex items-center justify-center shrink-0', className)}>
      <UserCircle2 className={cn(icon, 'text-[#555]')} />
    </div>
  );
}
