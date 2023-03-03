import { VueWrapper } from '@vue/test-utils';

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

export type ExtendedVueWrapper = VueWrapper & {
    vm: { [key: string]: any },
}
