import { flushPromises, MountingOptions, shallowMount } from '@vue/test-utils';
import { Module, createStore } from 'vuex';
import { allMocks } from './mocks';
import { Configuration } from './test-utils';

export async function shallowMountWithDefaults(vueComponent: any, configuration: Configuration = {}) {
    vueComponent.mixins = configuration.mixins || vueComponent.mixins;

    let mountConfiguration: MountingOptions<any> & Record<string, any> = {
        global: {
            ...configuration.global || {},
            plugins: [...(configuration.global?.plugins || [])],
        },
        data: configuration.data || function () {
            return {}; 
        },
        propsData: configuration.props || {},
        filters: configuration.filters || {},
        mocks: configuration.mocks || {},
        provide: configuration.injections || {},
        stubs: {
            ...allMocks,
            ...(configuration.stubs || {}),
        },
    };
    mountConfiguration = addStore(mountConfiguration, configuration);

    const wrapper = shallowMount(vueComponent, mountConfiguration);

    await flushPromises();

    return wrapper;
}

const addStore = (mountConfiguration: MountingOptions<any> & Record<string, any>, configuration: Configuration) => {
    if (!configuration.store) {
        return mountConfiguration;
    }

    const storesWithDefaults: { [moduleName: string]: Module<any, any> } = {};
    Object.keys(configuration.store).forEach((storeName) => {
        storesWithDefaults[storeName] = {
            namespaced: true,
            getters: {},
            actions: {},
            mutations: {},
            state: {},
            ...configuration.store[storeName],
        };
    });
    mountConfiguration.global?.plugins?.push(createStore({
        modules: storesWithDefaults,
    }));

    return mountConfiguration;
};