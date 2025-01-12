<template>
    <div>
        <div
            v-if="showExtendedLink"
            class="station-search-extended-link"
        >
            <a
                href="/"
                @click="onClickExtended"
            >
                Go to extended Search
            </a>
        </div>

        <div class="station-search">
            <v-text-field
                ref="searchTextField"
                v-model="currentQuery"
                :disabled="!currentInfrastructure"
                label="Search for item by name:"
                :error-messages="currentSearchError"
                hide-details="auto"
                @keydown.enter.prevent="searchName"
            />

            <soro-button
                label="Search"
                class="search-button"
                @click="searchName"
            />
        </div>

        <div
            v-if="showExtendedOptions"
            class="station-search-extended-options"
        >
            <v-checkbox
                v-for="searchType in validSearchTypes"
                :key="searchType"
                v-model="currentSearchTypes"
                :value="searchType"
                color="primary"
                density="compact"
                hide-details
                multiple
            >
                <template #label>
                    {{ `Search for ${ElementTypeLabels[searchType]}` }}
                    <img
                        class="station-search-search-type-icon"
                        :src="iconUrl + searchType + iconExtension"
                        alt=""
                    >
                </template>
            </v-checkbox>
        </div>

        <v-list
            v-if="currentSearchedMapPositions.length > 1"
            density="compact"
            elevation="5"
            class="station-search-result-list"
        >
            <v-list-subheader>SEARCH RESULTS</v-list-subheader>
            <v-list-item
                v-for="(mapPosition, index) in currentSearchedMapPositions"
                :key="index"
                :value="mapPosition.name"
                @click="setCurrentSearchedMapPosition(mapPosition.position)"
            >
                <v-list-item-title>
                    <img
                        v-if="validSearchTypes.includes(mapPosition.type)"
                        class="station-search-search-type-icon"
                        :src="iconUrl + mapPosition.type + iconExtension"
                        alt=""
                    >
                    <span> {{ getSearchResultLabelParts(mapPosition.name).before }}</span>
                    <strong class="search-match">{{ currentSearchTerm }}</strong>
                    <span>{{ getSearchResultLabelParts(mapPosition.name).after }}</span>
                </v-list-item-title>
            </v-list-item>
        </v-list>
    </div>
</template>

<script setup lang="ts">
import SoroButton from '@/components/soro-button.vue';
import { ElementTypeLabels } from '@/components/infrastructure/element-types';
import { iconUrl, iconExtension } from '@/components/infrastructure/add-icons';
</script>

<script lang="ts">
import { defineComponent } from 'vue';
import { InfrastructureNamespace } from '@/stores/infrastructure-store';
import { mapActions, mapMutations, mapState } from 'vuex';
import { ElementType } from '@/components/infrastructure/element-types';

const validSearchTypes = [
    ElementType.STATION,
    ElementType.HALT,
    ElementType.MAIN_SIGNAL,
];

export default defineComponent({
    name: 'StationSearch',

    props: {
        showExtendedLink: {
            type: Boolean,
            required: false,
            default: false,
        },
        showExtendedOptions: {
            type: Boolean,
            required: false,
            default: false,
        },
    },

    emits: ['change-to-extended'],

    data(): {
        validSearchTypes: typeof validSearchTypes,
        currentQuery: string | null,
        currentSearchTypes: string[],
        } {
        return {
            validSearchTypes,
            currentQuery: null,
            currentSearchTypes: [
                ElementType.STATION,
                ElementType.HALT,
            ],
        };
    },

    computed: {
        ...mapState(InfrastructureNamespace, [
            'currentInfrastructure',
            'currentSearchTerm',
            'currentSearchError',
            'currentSearchedMapPositions',
        ]),
    },

    methods: {
        searchName() {
            if (!this.currentQuery) {
                return;
            }

            const includedTypes: { [key: string]: boolean } = {};
            validSearchTypes.forEach((searchType) => {
                includedTypes[searchType] = this.currentSearchTypes.includes(searchType);
            });

            this.searchPositionFromName({
                query: this.currentQuery,
                includedTypes,
            });
        },

        getSearchResultLabelParts(searchResult: string): { before: string, after: string } {
            const beginIndex = searchResult.indexOf(this.currentSearchTerm);

            return {
                before: searchResult.substring(0, beginIndex),
                after: searchResult.substring(beginIndex + this.currentSearchTerm.length),
            };
        },

        onClickExtended(event: Event) {
            event.preventDefault();

            this.$emit('change-to-extended');
        },

        ...mapMutations(InfrastructureNamespace, ['setCurrentSearchedMapPosition']),
        ...mapActions(InfrastructureNamespace, ['searchPositionFromName']),
    },
});
</script>

<style scoped>
.station-search {
    display: flex;
    margin-bottom: 10px;
}

.search-button {
    margin-left: 10px;
    height: auto;
}

.search-match {
    color: rgb(var(--v-theme-primary));
}

.station-search-extended-link {
    display: flex;
    color: rgb(var(--v-theme-primary));
    justify-content: right;
}

a,
a:visited,
a:hover,
a:active {
    color: inherit;
}

.station-search-result-list.v-list {
    overflow-y: auto;
    max-height: 500px;
}

.station-search-search-type-icon {
    margin-left: 5px;
    margin-right: 5px;
    height: 1.2em;
}
</style>