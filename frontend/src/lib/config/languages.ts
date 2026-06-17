import languagesData from './languages.json';

export type LanguageOption = {
	code: string;
	label: string;
};

/** Supported speech / caption languages (BCP-47 codes). */
export const LANGUAGES = languagesData as LanguageOption[];

export const DEFAULT_SOURCE_LANG = 'en-US';
export const DEFAULT_TARGET_LANG = 'ru-RU';

/** Strip region subtag: `en-US` → `en`. */
export function langBase(code: string): string {
	return code.split('-')[0] ?? code;
}

export function languageLabel(code: string): string {
	return LANGUAGES.find((l) => l.code === code)?.label ?? code;
}

export function firstLanguageOtherThan(code: string): LanguageOption | undefined {
	const base = langBase(code);
	return LANGUAGES.find((l) => langBase(l.code) !== base);
}
