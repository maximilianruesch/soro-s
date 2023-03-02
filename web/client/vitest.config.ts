import { defineConfig } from 'vitest/config';
import viteConfig from './vite.config';
import { mergeConfig } from 'vite';

export default mergeConfig(viteConfig, defineConfig({
    test: {
        watch: false,
        globals: true,
        silent: true,
        environment: 'jsdom',
        deps: {
            inline: ['vuetify'],
        },
    },
}));
