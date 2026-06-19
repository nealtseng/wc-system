import React from 'react';
import { cn } from '../lib/utils';

export function NationFlag({ iso2, alt, className }: {
  iso2: string;
  alt?: string;
  className?: string;
}) {
  return (
    <img
      src={`https://flagcdn.com/w80/${iso2.toLowerCase()}.png`}
      alt={alt ?? iso2}
      className={cn("inline-block object-cover border border-[#333] shadow-sm", className)}
      draggable={false}
      loading="lazy"
    />
  );
}
