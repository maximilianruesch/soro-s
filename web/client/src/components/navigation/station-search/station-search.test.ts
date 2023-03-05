import { shallowMountWithDefaults } from '@test-utils/shallow-mount-with-defaults';
import StationSearch from './station-search.vue';
import { InfrastructureNamespace } from '@/stores/infrastructure-store';
import { ExtendedVueWrapper } from '@test-utils/test-utils';

describe('station-search', async () => {
    let stationSearch: ExtendedVueWrapper<typeof StationSearch>;
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

    it('updates the current query when the search text field emits \'change\' event', async () => {
        const searchTextField = stationSearch.findComponent({ ref: 'searchTextField' }) as ExtendedVueWrapper;
        searchTextField.vm.$emit('change', { target: { value: 'some-query' } });

        expect(stationSearch.vm.$data.currentQuery).toBe('some-query');
    });

    // Testing the following is difficult with shallowMount (as of event modifiers like '.enter' and '.prevent'), we may
    // have to think of a workaround
    describe.todo('when the search text field emits an event following a \'enter\' key press');

    describe('when the search button emits a \'click\' event', async () => {
        it('does not call \'searchPositionFromName\' if no query is entered', async () => {
            await stationSearch.setData({
                currentQuery: null,
            });

            const searchButton = stationSearch.findComponent('.search-button') as ExtendedVueWrapper;
            searchButton.vm.$emit('click');

            expect(searchPositionFromName).not.toHaveBeenCalled();
        });

        it(
            'calls \'searchPositionFromName\' with the query and all selected search types if a query is entered',
            async () => {
                await stationSearch.setData({
                    currentQuery: 'some-query',
                    currentSearchTypes: [
                        'station',
                        'foo',
                    ],
                });

                const searchButton = stationSearch.findComponent('.search-button') as ExtendedVueWrapper;
                searchButton.vm.$emit('click');

                expect(searchPositionFromName).toHaveBeenCalledWith(
                    expect.any(Object),
                    {
                        query: 'some-query',
                        includedTypes: {
                            station: true,
                            hlt: false,
                            ms: false,
                        },
                    },
                );
            },
        );
    });

    it('shows a checkbox for each of the valid search types when setting \'showExtendedOptions\' to true', async () => {
        stationSearch = await shallowMountWithDefaults(StationSearch, {
            ...defaults,
            props: { showExtendedOptions: true },
        });

        const extendedOptionList = stationSearch.find('.station-search-extended-options');

        expect(extendedOptionList.exists()).toBe(true);
        const checkboxes = extendedOptionList.findAllComponents({ name: 'v-checkbox' });
        expect(checkboxes).toHaveLength(3);
    });

    it('does not show checkboxes when setting \'showExtendedOptions\' to false', async () => {
        stationSearch = await shallowMountWithDefaults(StationSearch, {
            ...defaults,
            props: { showExtendedOptions: false },
        });

        const extendedOptionList = stationSearch.find('.station-search-extended-options');
        expect(extendedOptionList.exists()).toBe(false);
    });
});
