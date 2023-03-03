import { shallowMountWithDefaults } from '@test-utils/shallow-mount-with-defaults';
import SoroNavigationSearchContent from './soro-navigation-search-content.vue';
import { VueWrapper } from '@vue/test-utils';

describe('station-search', async () => {
    it('contains a station search with extended options', async () => {
        expect.assertions(2);

        const soroNavigationSearchContent = await shallowMountWithDefaults(SoroNavigationSearchContent);
        const stationSearch = soroNavigationSearchContent.findComponent('station-search-stub') as VueWrapper;

        expect(stationSearch.exists()).toBe(true);
        expect(stationSearch.vm.$props).toStrictEqual(expect.objectContaining({
            showExtendedOptions: true,
        }));
    });
});
