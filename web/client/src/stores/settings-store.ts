import { Module } from 'vuex';
import { VuetifyExtension } from '@/vuetify';

type SettingsState = {
    darkLightModePreference: 'dark' | 'light',
    primaryColor: string | null,
}

export const SettingsNamespace = 'settings';

export const SettingsStore: Module<SettingsState, undefined> = {
    namespaced: true,

    state() {
        return {
            darkLightModePreference: 'light',
            primaryColor: null,
        };
    },

    mutations: {
        setDarkLightModePreference(this: VuetifyExtension, state, darkLightModePreference) {
            state.darkLightModePreference = darkLightModePreference;
            // We need to set the theme globally in vuetify to access its properties in components
            this.$vuetify.theme.global.name.value = darkLightModePreference;
        },

        setPrimaryColor(this: VuetifyExtension, state, primaryColor) {
            state.primaryColor = primaryColor;
            // We need to set the primary color globally in vuetify to access its properties in components
            const themes = this.$vuetify.theme.themes.value;
            Object.keys(themes).forEach((themeKey) => themes[themeKey].colors.primary = primaryColor);
        },
    },
};