import { shallowMountWithDefaults } from '@test-utils/shallow-mount-with-defaults';
import StationSearch from '@/components/station-search.vue';
import { InfrastructureNamespace } from '@/stores/infrastructure-store';
import { Mock } from 'vitest';

describe('station-search', async () => {
    let stationSearch: any;
    let searchPositionFromName: Mock;
    let setCurrentSearchedMapPosition: Mock;
    const infrastructureState = {
        currentInfrastructure: '',
        currentSearchTerm: '',
        currentSearchError: '',
        currentSearchedMapPositions: [],
    };

    beforeEach(async () => {
        searchPositionFromName = vi.fn();
        setCurrentSearchedMapPosition = vi.fn();
        stationSearch = await shallowMountWithDefaults(StationSearch, {
            store: {
                [InfrastructureNamespace]: {
                    namespaced: true,
                    state: infrastructureState,
                    mutations: { setCurrentSearchedMapPosition },
                    actions: { searchPositionFromName },
                },
            },
        });
    });

    describe('when the search button emits a click event', async () => {
        it('does not call \'searchPositionFromName\' when no query is entered', function () {
            expect.assertions(1);

            stationSearch.vm.currentQuery = null;

            const searchButton = stationSearch.findComponent({ ref: 'search-button' });
            searchButton.vm.$emit('click');

            expect(searchPositionFromName).not.toHaveBeenCalled();
        });
    });
});
