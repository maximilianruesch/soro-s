import { Module } from 'vuex';
import { MapPosition } from '@/components/infrastructure/infrastructure-map.vue';
import { sendRequest } from '@/api/api-client';

type InfrastructureState = {
    infrastructures: string[],
    currentInfrastructure?: string,
    currentSearchedMapPosition?: MapPosition,
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

            fetch(`${window.origin}/search?query=${query}&infrastructure=${state.currentInfrastructure}`)
                .then(response => response.json())
                .then(position => commit('setCurrentSearchedMapPosition', position));
        }
    },
};