<template>
    <div class="station-search">
        <v-text-field
            :disabled="!currentInfrastructure"
            label="Search for station or halt by name:"
            :error-messages="currentSearchError"
            hide-details="auto"
            @change="event => currentQuery = event.target.value"
        />

        <soro-button
            label="Search"
            class="search-button"
            @click="searchName"
        />
    </div>
</template>

<script setup lang="ts">
import SoroButton from '@/components/soro-button.vue';
</script>

<script lang="ts">
import { defineComponent } from 'vue';
import { InfrastructureNamespace } from '@/stores/infrastructure-store';
import { mapActions, mapState } from 'vuex';

export default defineComponent({
    name: 'StationSearch',

    data() {
        return {
            currentQuery: null
        };
    },

    computed: {
        ...mapState(InfrastructureNamespace, [
            'currentInfrastructure',
            'currentSearchError',
        ]),
    },

    methods: {
        searchName() {
            if (!this.currentQuery) {
                return;
            }

            console.log(`Will now try searching for ${this.currentQuery}`);
            this.searchPositionFromName(this.currentQuery);
        },

        ...mapActions(InfrastructureNamespace, ['searchPositionFromName']),
    },
});
</script>

<style scoped>
.station-search {
    display: flex;
}

.search-button {
    margin-left: 10px;
    height: auto;
}
</style>