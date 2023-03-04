import { shallowMountWithDefaults } from '@test-utils/shallow-mount-with-defaults';
import StationSearch from './station-search.vue';
import { InfrastructureNamespace } from '@/stores/infrastructure-store';
import { ExtendedVueWrapper } from '@test-utils/test-utils';

describe('station-search', async () => {
    let stationSearch: ExtendedVueWrapper;
    const searchPositionFromName = vi.fn();
    const setCurrentSearchedMapPosition = vi.fn();
    const infrastructureState = {
        currentInfrastructure: '',
        currentSearchTerm: '',
        currentSearchError: '',
        currentSearchedMapPositions: [],
    };

    const defaults = {
        store: {
            [InfrastructureNamespace]: {
                namespaced: true,
                state: infrastructureState,
                mutations: { setCurrentSearchedMapPosition },
                actions: { searchPositionFromName },
            },
        },
    };

    beforeEach(async () => {
        vi.clearAllMocks();
        stationSearch = await shallowMountWithDefaults(StationSearch, defaults);
    });

    describe('when the search button emits a click event', async () => {
        it('does not call \'searchPositionFromName\' if no query is entered', function () {
            stationSearch.vm.currentQuery = null;

            const searchButton = stationSearch.findComponent({ ref: 'search-button' });
            searchButton.vm.$emit('click');

            expect(searchPositionFromName).not.toHaveBeenCalled();
        });
    });

    describe('when setting \'showExtendedLink\' to true', function () {
        beforeEach(async () => {
            stationSearch = await shallowMountWithDefaults(StationSearch, {
                ...defaults,
                props: { showExtendedLink: true },
            });
        });

        it('shows an extended link', async () => {
            const showExtendedLink = stationSearch.find('.station-search-extended-link');
            expect(showExtendedLink.exists()).toBe(true);
        });

        it('emits \'change-to-extended\' when clicking the extended link', async () => {
            const showExtendedLink = stationSearch.find('.station-search-extended-link');

            await showExtendedLink.find('a').trigger('click');

            expect(stationSearch.emitted('change-to-extended')).toHaveLength(1);
        });
    });

    it('does not show an extended link when setting \'showExtendedLink\' to false', async () => {
        stationSearch = await shallowMountWithDefaults(StationSearch, {
            ...defaults,
            props: { showExtendedLink: false },
        });

        const showExtendedLink = stationSearch.find('.station-search-extended-link');
        expect(showExtendedLink.exists()).toBe(false);
    });
});
