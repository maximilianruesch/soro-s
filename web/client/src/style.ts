import '../deps/GoldenLayout/css/goldenlayout-base.css';
import '../deps/GoldenLayout/css/themes/goldenlayout-mdl-theme.css';
import 'material-design-icons-iconfont/dist/material-design-icons.css';
import 'vuetify/styles';
import './style/style.css';
import { ThemeDefinition } from 'vuetify';

// These are partial overrides of the light and dark themes provided by vuetify.
export const customLightTheme: ThemeDefinition = {
    colors: {
        primary: '#2196F3',
    },
};

export const customDarkTheme: ThemeDefinition = {
    colors: {
        primary: '#2196F3',
    },
};
