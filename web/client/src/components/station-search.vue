<template>
    <div class="station-search">
        <soro-text-field
            label="Search for station or halt by name:"
            class="search-text-field"
            @change="event => currentQuery = event"
        />

        <soro-button
            label="Search"
            class="search-button"
            @click="searchName"
        />
    </div>
</template>

<script setup lang="ts">
import SoroTextField from '@/components/common/soro-text-field.vue';
import SoroButton from '@/components/soro-button.vue';
</script>

<script lang="ts">
import { defineComponent } from 'vue';
import { InfrastructureNamespace } from '@/stores/infrastructure-store';
import { mapActions } from 'vuex';

export default defineComponent({
    name: 'StationSearch',

    data() {
        return {
            currentQuery: null
        };
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
    height: 60px;
}

.search-text-field {
    width: 80%;
    height: 100%;
}

.search-button {
    width: 20%;
    height: 100%;
}
</style>