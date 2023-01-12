<template>
    <div class="soro-text-field matter-textfield-filled">
        <input
            placeholder=" "
            type="text"
            @input="emitChange"
        >
        <span>{{ label }}</span>
    </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
    name: 'SoroTextField',

    props: {
        label: {
            type: String,
            required: true,
        },
    },

    emits: ['change'],

    methods: {
        emitChange(event: Event) {
            event.preventDefault();
            this.$emit('change', (event.target as HTMLInputElement).value);
        },
    },
});
</script>

<style scoped>
.soro-text-field {
    margin-left: 3%;
    margin-right: 3%;
    width: 100%;
}
</style>