import { Module } from 'vuex';

type SettingsState = {
    darkLightModePreference: typeof DarkLightModes[keyof typeof DarkLightModes],
    theme: typeof DarkLightModes.DARK | typeof DarkLightModes.LIGHT,
}

export const DarkLightModes = {
    DARK: 'dark',
    LIGHT: 'light',
    OS: 'os',
};

export const SettingsNamespace = 'settings';

export const SettingsStore: Module<SettingsState, undefined> = {
    namespaced: true,

    state() {
        return {
            darkLightModePreference: 'light',
            theme: 'light',
        };
    },

    mutations: {
        setDarkLightModePreference(state, darkLightModePreference) {
            state.darkLightModePreference = darkLightModePreference;
        },
        
        setTheme(state, theme) {
            state.theme = theme;
        }
    },

    actions: {
        loadSettings({ dispatch }) {
            dispatch('initThemeListener');
            dispatch('applyTheme');
        },

        setDarkLightModePreference({ commit, dispatch }, preference) {
            commit('setDarkLightModePreference', preference);
            dispatch('applyTheme');
        },

        initThemeListener({ commit, state }) {
            const themeMediaQuery = window.matchMedia('(prefers-color-scheme: dark)');

            themeMediaQuery.addEventListener('change', (event) => {
                if (state.darkLightModePreference !== DarkLightModes.OS) {
                    return;
                }

                commit('setTheme', event.matches ? DarkLightModes.DARK : DarkLightModes.LIGHT);
            });
        },

        applyTheme({ commit, state }) {
            const themeMediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
            if (state.darkLightModePreference === DarkLightModes.OS) {
                commit('setTheme', themeMediaQuery.matches ? DarkLightModes.DARK : DarkLightModes.LIGHT);

                return;
            }

            commit('setTheme', state.darkLightModePreference);
        }
    }
};