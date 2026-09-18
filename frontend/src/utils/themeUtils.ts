import type {ThemeName} from '../constants/themes'; export function applyTheme(theme:ThemeName){document.documentElement.dataset.theme=theme;localStorage.setItem('mindgarden_theme',theme)}
