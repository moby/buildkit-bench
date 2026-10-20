export function formatValue(value, unit) {
  if (!Number.isFinite(value)) return 'Not recorded';
  if (unit === 'bytes') {
    const scale =
      value === 0 ? 0 : Math.min(4, Math.max(0, Math.floor(Math.log(Math.abs(value)) / Math.log(1024))));
    return `${new Intl.NumberFormat('en', { maximumFractionDigits: 2 }).format(value / 1024 ** scale)} ${
      ['B', 'KiB', 'MiB', 'GiB', 'TiB'][scale]
    }`;
  }
  return `${new Intl.NumberFormat('en', { maximumSignificantDigits: 5 }).format(value)}${
    unit ? ` ${unit}` : ''
  }`;
}

export function runLabel(name) {
  if (!/^\d{8}-\d{6}$/.test(name || '')) return name || 'Select a run';
  return `${name.slice(0, 4)}-${name.slice(4, 6)}-${name.slice(6, 8)} · ${name.slice(9, 11)}:${name.slice(
    11,
    13
  )}`;
}

export function runTimestamp(name) {
  if (!/^\d{8}-\d{6}$/.test(name)) return null;
  return Date.UTC(
    +name.slice(0, 4),
    +name.slice(4, 6) - 1,
    +name.slice(6, 8),
    +name.slice(9, 11),
    +name.slice(11, 13),
    +name.slice(13, 15)
  );
}
