import type { TeamInfo, TeamResponse } from '../types';

export function teamFromResponse(t: TeamResponse): TeamInfo {
  return {
    id: t.id,
    name: t.name,
    iso2: t.iso2 || 'un',
    group: t.wc_group || '',
    elo: t.elo,
    gdp:
      t.gdp_per_capita > 0
        ? `$${t.gdp_per_capita.toLocaleString(undefined, { maximumFractionDigits: 0 })}`
        : 'N/A',
    gdpValue: t.gdp_per_capita,
  };
}
