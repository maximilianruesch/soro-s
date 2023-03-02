import { flushPromises, MountingOptions, shallowMount } from '@vue/test-utils';
import { Module, createStore } from 'vuex';
import { allMocks } from './mocks';

type Configuration = {
    props?: any,
    data?: any,
    mixins?: any,
    global?: any,
    mocks?: any,
    filters?: any,
    injections?: any,
    stubs?: any,
    store?: any,
};

export async function shallowMountWithDefaults(vueComponent: any, configuration: Configuration) {
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
        mocks: {
            ...allMocks,
            ...(configuration.mocks || {}),
        },
        provide: configuration.injections || {},
        stubs: configuration.stubs || {},
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