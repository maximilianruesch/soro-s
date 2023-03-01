<template>
    <div>
        <station-search class="station-search-field" />
    </div>
</template>

<script setup lang="ts">
import StationSearch from './station-search.vue';
</script>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
    name: 'SoroNavigationSearchContent',
});
</script>

<style>
.station-search-field {
    padding: 3%;
}
</style>
