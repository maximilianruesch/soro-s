import { Module } from 'vuex';
import { MapPosition } from '@/components/infrastructure/infrastructure-map.vue';
import { sendPostData, sendRequest } from '@/api/api-client';
import { ElementType } from '@/components/infrastructure/elementTypes';

type InfrastructureState = {
    infrastructures: string[],
    currentInfrastructure?: string,
    currentSearchedMapPosition?: MapPosition,
    currentSearchedMapPositions: {
        name: string,
        position: MapPosition,
    }[],
    currentSearchTerm?: string,
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
            currentSearchTerm: undefined,
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

        setCurrentSearchTerm(state, currentSearchTerm) {
            state.currentSearchTerm = currentSearchTerm;
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
            commit('setCurrentInfrastructure', infrastructureFilename);
        },

        unload({ commit }) {
            commit('setCurrentInfrastructure', null);
        },

        searchPositionFromName({ commit, state }, { query, includedTypes }) {
            if (!state.currentInfrastructure) {
                console.error('Tried search with no selected infrastructure');

                return;
            }

            if (!query) {
                commit('setCurrentSearchedMapPosition', null);

                return;
            }

            sendPostData({
                url: 'search',
                data: {
                    query,
                    infrastructure: state.currentInfrastructure,
                    options: { includedTypes },
                },
            })
                .then(response => response.json())
                .then(positions => {
                    commit('setCurrentSearchedMapPositions', positions);
                    commit('setCurrentSearchTerm', query);

                    if (positions.length === 0) {
                        commit('setCurrentSearchError', 'Not found!');

                        return;
                    }

                    commit('setCurrentSearchedMapPosition', positions[0]?.position);
                    commit('setCurrentSearchError', undefined);
                })
                .catch(() => {
                    commit('setCurrentSearchError', 'An error occurred!');

                    return;
                });
        },
    },
};