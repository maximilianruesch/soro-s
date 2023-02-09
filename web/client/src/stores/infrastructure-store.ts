import { Module } from 'vuex';
import { MapPosition } from '@/components/infrastructure/infrastructure-map.vue';
import { sendRequest, transformUrl } from '@/api/api-client';

type InfrastructureState = {
    infrastructures: string[],
    currentInfrastructure?: string,
    currentSearchedMapPosition?: MapPosition,
    currentSearchedMapPositions: {
        name: string,
        position: MapPosition,
    }[],
    currentSearchError?: string,
    highlightedSignalStationRouteID?: string,
    highlightedStationRouteID?: string,
}

type InfrastructureFetchResponse = { dirs: string[] }

export const InfrastructureNamespace = 'infrastructure';

export const InfrastructureStore: Module<InfrastructureState, undefined> = {
    namespaced: true,

    state() {
        return {
            infrastructures: [],
            currentInfrastructure: undefined,
            currentSearchedMapPosition: undefined,
            currentSearchedMapPositions: [],
            currentSearchError: undefined,
            highlightedSignalStationRouteID: undefined,
            highlightedStationRouteID: undefined,
        };
    },

    mutations: {
        setInfrastructures(state, infrastructures) {
            state.infrastructures = infrastructures;
        },

        setCurrentInfrastructure(state, currentInfrastructure) {
            state.currentInfrastructure = currentInfrastructure;
        },

        setCurrentSearchedMapPosition(state, currentSearchedMapPosition) {
            state.currentSearchedMapPosition = currentSearchedMapPosition;
        },

        setCurrentSearchedMapPositions(state, currentSearchedMapPositions) {
            state.currentSearchedMapPositions = currentSearchedMapPositions;
        },

        setCurrentSearchError(state, currentSearchError) {
            state.currentSearchError = currentSearchError;
        },

        setHighlightedSignalStationRouteID(state, highlightedSignalStationRouteID) {
            state.highlightedSignalStationRouteID = highlightedSignalStationRouteID;
        },

        setHighlightedStationRouteID(state, highlightedStationRouteID) {
            state.highlightedStationRouteID = highlightedStationRouteID;
        },
    },

    actions: {
        initialLoad({ commit }) {
            sendRequest({ url: 'infrastructure' })
                .then(response => response.json())
                .then((dir: InfrastructureFetchResponse) => {
                    commit('setInfrastructures', dir.dirs.filter((option: string) => option !== '.' && option !== '..'));
                });
        },

        load({ commit }, infrastructureFilename) {
            console.log('Switching to infrastructure to', infrastructureFilename);
            commit('setCurrentInfrastructure', infrastructureFilename);
        },

        unload({ commit }) {
            commit('setCurrentInfrastructure', null);
        },

        searchPositionFromName({ commit, state }, query) {
            if (!state.currentInfrastructure) {
                console.error('Tried search with no selected infrastructure');

                return;
            }

            if (!query) {
                commit('setCurrentSearchedMapPosition', null);

                return;
            }

            fetch(transformUrl(`search?query=${query}&infrastructure=${state.currentInfrastructure}`))
                .then(response => response.json())
                .then(position => {
                    const somePositions = [
                        {
                            name: 'Darmstadt-Eberstadt',
                            position: {
                                lat: 49.8144694,
                                lon: 8.6259571,
                            },
                        },
                        {
                            name: 'Frankfurt',
                            position: {
                                lat: 50.1039142,
                                lon: 8.6448659,
                            },
                        },
                        {
                            name: 'Kassel',
                            position: {
                                lat: 51.3113881,
                                lon: 9.4477049,
                            },
                        }
                    ];
                    const realPositions = Array.isArray(position) ? position : somePositions;

                    commit('setCurrentSearchedMapPositions', realPositions);
                    commit('setCurrentSearchedMapPosition', realPositions[0]?.position);
                    commit('setCurrentSearchError', undefined);
                })
                .catch(() => {
                    commit('setCurrentSearchError', 'Not found!');

                    return;
                });
        }
    },
};