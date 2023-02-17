import { ResolvedLayoutConfig } from 'golden-layout';

export type GoldenLayoutAdapter = {
    addGLComponent: (componentType: string, title: string) => void,
    loadGLLayout: (layout: ResolvedLayoutConfig) => void,
    getLayoutConfig: () => ResolvedLayoutConfig,
}

declare module 'golden-layout-adapter.vue' {
    const adapter: GoldenLayoutAdapter;
    export default adapter;
}
