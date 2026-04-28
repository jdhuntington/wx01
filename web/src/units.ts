export type UnitSystem = 'metric' | 'imperial';

export function tempConvert(c: number | null, system: UnitSystem): number | null {
  if (c === null) return null;
  return system === 'imperial' ? c * 9 / 5 + 32 : c;
}

export function tempUnit(system: UnitSystem): string {
  return system === 'imperial' ? '°F' : '°C';
}

export function windConvert(ms: number | null, system: UnitSystem): number | null {
  if (ms === null) return null;
  return system === 'imperial' ? ms * 2.23694 : ms;
}

export function windUnit(system: UnitSystem): string {
  return system === 'imperial' ? 'mph' : 'm/s';
}

export function rainConvert(mm: number | null, system: UnitSystem): number | null {
  if (mm === null) return null;
  return system === 'imperial' ? mm * 0.03937 : mm;
}

export function rainUnit(system: UnitSystem): string {
  return system === 'imperial' ? 'in' : 'mm';
}

export function pressureConvert(mb: number | null, system: UnitSystem): number | null {
  if (mb === null) return null;
  return system === 'imperial' ? mb * 0.02953 : mb;
}

export function pressureUnit(system: UnitSystem): string {
  return system === 'imperial' ? 'inHg' : 'mb';
}

export function distanceConvert(km: number | null, system: UnitSystem): number | null {
  if (km === null) return null;
  return system === 'imperial' ? km * 0.621371 : km;
}

export function distanceUnit(system: UnitSystem): string {
  return system === 'imperial' ? 'mi' : 'km';
}

export function getStoredUnits(): UnitSystem {
  return (localStorage.getItem('wx01_units') as UnitSystem) || 'metric';
}

export function setStoredUnits(system: UnitSystem) {
  localStorage.setItem('wx01_units', system);
}
