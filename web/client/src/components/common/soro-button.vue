<template>
    <button
        class="matter-button-contained soro-button"
        @click="(event: Event) => $emit('click', event)"
    >
        {{ label }}
    </button>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue';

export default defineComponent({
    name: 'SoroButton',

    props: {
        label: {
            type: String as PropType<string>,
            required: true,
        },
    },

    emits: ['click'],
});
</script>
