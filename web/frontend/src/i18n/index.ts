import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';

const resources = {
    'zh-CN': { translation: {} },
    en: { translation: {} },
};

const saved = localStorage.getItem('lang');
const defaultLang = saved || 'zh-CN';

i18n
    .use(initReactI18next)
    .init({
        resources,
        lng: defaultLang,
        fallbackLng: 'zh-CN',
        interpolation: {
            escapeValue: false,
        },
    });

export default i18n;
