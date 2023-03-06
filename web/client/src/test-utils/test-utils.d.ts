import { VueWrapper } from '@vue/test-utils';
import { ComponentPublicInstance } from 'vue';

export type Configuration = {
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

export type ExtendedVueWrapper<T = ComponentPublicInstance> = VueWrapper<T> & {
    vm: { [key: string]: any },
}
