import { readFile } from 'node:fs/promises';

const app = await readFile(new URL('../web/app.js', import.meta.url), 'utf8');
const html = await readFile(new URL('../web/index.html', import.meta.url), 'utf8');
const objectMatch = app.match(/const translations = (\{[\s\S]*?\});\n\nconst state/);

if (!objectMatch) {
  throw new Error('Could not locate the translation dictionary');
}

const dictionaries = Function(`"use strict"; return (${objectMatch[1]});`)();
const locales = Object.keys(dictionaries);
const referenceKeys = new Set(Object.keys(dictionaries[locales[0]]));
const usedKeys = new Set([
  ...Array.from(html.matchAll(/data-i18n(?:-aria-label|-placeholder)?="([^"]+)"/g), match => match[1]),
  ...Array.from(app.matchAll(/\bt\('([^']+)'/g), match => match[1])
]);
const errors = [];

for (const locale of locales) {
  const keys = new Set(Object.keys(dictionaries[locale]));
  for (const key of referenceKeys) {
    if (!keys.has(key)) errors.push(`${locale} is missing ${key}`);
  }
  for (const key of keys) {
    if (!referenceKeys.has(key)) errors.push(`${locale} has unexpected ${key}`);
  }
  for (const key of usedKeys) {
    if (!keys.has(key)) errors.push(`${locale} does not define used key ${key}`);
  }
}

if (errors.length) {
  console.error(errors.join('\n'));
  process.exit(1);
}

for (const fragment of [
  "setLocale(event.target.value, true)",
  "setLocale(saved.dashboard?.language || state.locale)"
]) {
  if (!app.includes(fragment)) {
    throw new Error(`Missing unified language behavior: ${fragment}`);
  }
}

console.log(`Checked ${referenceKeys.size} keys across ${locales.length} locales.`);
