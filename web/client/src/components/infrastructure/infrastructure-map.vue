<template>
    <div ref="container">
        <div
            ref="map"
            class="map infrastructure-map"
        />
        <infrastructure-legend
            class="map-overlay"
            :checked-controls="checkedControls"
            @change="onLegendControlChanged"
            @reset="resetLegend"
        />
        <div
            ref="infrastructureTooltip"
            class="infrastructureTooltip infrastructure-tooltip"
        >
            <ul id="infrastructureTooltipList">
                <li id="kilometerPoint" />
                <li id="risingOrFalling" />
            </ul>
        </div>
    </div>
</template>

<script lang="ts">
import { mapState } from 'vuex';
import { InfrastructureNamespace } from '@/stores/infrastructure-store';
import {
    deHighlightSignalStationRoute,
    deHighlightStationRoute,
    highlightSignalStationRoute,
    highlightStationRoute,
} from './highlight-helpers';
import { FilterSpecification, Map } from 'maplibre-gl';
import { createInfrastructureMapStyle } from './map-style';
import { addIcons } from './add-icons';
import { ElementTypes, ElementType } from './element-types';
import { defineComponent } from 'vue';
import { transformUrl } from '@/api/api-client';
import { ThemeInstance, useTheme } from 'vuetify';
import { SpecialLegendControls, SpecialLegendControl } from '@/components/infrastructure/infrastructure-legend.vue';
import InfrastructureLegend from '@/components/infrastructure/infrastructure-legend.vue';

export const initiallyCheckedControls = [
    ElementType.STATION,
    ElementType.HALT,
    ElementType.MAIN_SIGNAL,
    ElementType.APPROACH_SIGNAL,
    ElementType.END_OF_TRAIN_DETECTOR,
    ...SpecialLegendControls,
];

const mapDefaults = {
    attributionControl: false,
    zoom: 10,
    hash: 'location',
    center: [10, 50],
    maxBounds: [[-5, 40], [25, 60]], // [SW Point] [NE Point] in LonLat
    bearing: 0,
};

export type MapPosition = {
    lat: number,
    lon: number,
};

export default defineComponent({
    name: 'InfrastructureMap',
    components: { InfrastructureLegend },
    inject: {
        goldenLayoutKeyInjection: {
            default: '',
        },
    },

    setup() {
        return { currentTheme: useTheme().global };
    },

    data(): {
        libreGLMap?: Map,
        checkedControls: typeof initiallyCheckedControls,
        } {
        return {
            libreGLMap: undefined,
            checkedControls: Array.from(initiallyCheckedControls),
        };
    },

    computed: {
        checkedLegendControlLocalStorageKey() {
            return `infrastructure[${this.goldenLayoutKeyInjection}].checkedControls`;
        },

        ...mapState(InfrastructureNamespace, [
            'currentInfrastructure',
            'currentSearchedMapPosition',
            'highlightedSignalStationRouteID',
            'highlightedStationRouteID',
        ]),
    },

    watch: {
        currentInfrastructure(newInfrastructure: string | null) {
            if (this.libreGLMap) {
                this.libreGLMap.remove();
                this.libreGLMap = undefined;
            }

            if (!newInfrastructure) {
                return;
            }

            // Re-instantiating the map on infrastructure change currently leads to duplicated icon fetching on change.
            this.createMap(newInfrastructure);
        },

        currentSearchedMapPosition(mapPosition: MapPosition) {
            if (!this.libreGLMap) {
                return;
            }

            this.libreGLMap.jumpTo({
                center: mapPosition,
                zoom: 14,
            });
        },

        currentTheme: {
            handler(newTheme: ThemeInstance) {
                if (!this.libreGLMap) {
                    return;
                }

                this.libreGLMap.setStyle(createInfrastructureMapStyle({
                    currentTheme: newTheme.current.value,
                    activatedElements: this.checkedControls,
                }));
            },
            deep: true,
        },

        highlightedSignalStationRouteID(newID, oldID) {
            if (!this.libreGLMap) {
                return;
            }

            if (newID) {
                // @ts-ignore
                highlightSignalStationRoute(this.libreGLMap, this.currentInfrastructure, newID);
            } else {
                // @ts-ignore
                deHighlightSignalStationRoute(this.libreGLMap, oldID);
            }
        },

        highlightedStationRouteID(newID, oldID) {
            if (!this.libreGLMap) {
                return;
            }

            if (newID) {
                // @ts-ignore
                highlightStationRoute(this.libreGLMap, this.currentInfrastructure, newID);
            } else {
                // @ts-ignore
                deHighlightStationRoute(this.libreGLMap, oldID);
            }
        },
    },

    created() {
        const checkedControlsString = window.localStorage.getItem(this.checkedLegendControlLocalStorageKey);
        if (checkedControlsString) {
            this.checkedControls = JSON.parse(checkedControlsString);
        }
    },

    mounted() {
        if (!this.currentInfrastructure) {
            return;
        }
        this.createMap(this.currentInfrastructure);
    },

    methods: {
        onLegendControlChanged(legendControl: string, checked: boolean) {
            if (checked) {
                this.checkedControls.push(legendControl);
            } else {
                this.checkedControls = this.checkedControls.filter((control) => control !== legendControl);
            }

            this.saveControls();

            if (!this.libreGLMap) {
                return;
            }

            if (SpecialLegendControls.includes(legendControl)) {
                this.evaluateSpecialLegendControls();

                return;
            }

            this.setElementTypeVisibility(legendControl, checked);
        },

        resetLegend() {
            this.checkedControls = initiallyCheckedControls;
            this.saveControls();
            this.setVisibilityOfAllControls();
        },

        saveControls() {
            window.localStorage.setItem(this.checkedLegendControlLocalStorageKey, JSON.stringify(this.checkedControls));
        },

        setVisibilityOfAllControls() {
            ElementTypes.forEach((type) => this.setElementTypeVisibility(type, this.checkedControls.includes(type)));
            this.evaluateSpecialLegendControls();
        },

        setElementTypeVisibility(elementType: string, visible: boolean) {
            if (elementType !== ElementType.STATION) {
                this.libreGLMap?.setLayoutProperty(
                    `circle-${elementType}-layer`,
                    'visibility',
                    visible ? 'visible': 'none',
                );
            }

            this.libreGLMap?.setLayoutProperty(
                `${elementType}-layer`,
                'visibility',
                visible ? 'visible': 'none',
            );
        },

        evaluateSpecialLegendControls() {
            const risingChecked = this.checkedControls.includes(SpecialLegendControl.RISING);
            const fallingChecked = this.checkedControls.includes(SpecialLegendControl.FALLING);

            let filter: FilterSpecification;
            if (!risingChecked && fallingChecked) {
                filter = ['!', ['get', 'rising']];
            } else if (risingChecked && !fallingChecked) {
                filter = ['get', 'rising'];
            } else if (!risingChecked && !fallingChecked) {
                filter = ['boolean', false];
            }

            ElementTypes.forEach((elementType) => {
                if (elementType === ElementType.STATION) {
                    return;
                }

                this.libreGLMap?.setFilter(elementType + '-layer', filter);
                this.libreGLMap?.setFilter('circle-' + elementType + '-layer', filter);
            });
        },

        createMap(infrastructure: string) {
            this.libreGLMap = new Map({
                ...mapDefaults,
                container: this.$refs.map as HTMLElement,
                // @ts-ignore
                transformRequest: (relative_url) => {
                    if (relative_url.startsWith('/')) {
                        return { url: transformUrl(`/${infrastructure}${relative_url}`) };
                    }
                },
                style: createInfrastructureMapStyle({
                    currentTheme: this.$vuetify.theme.current,
                    activatedElements: this.checkedControls,
                }),
            });

            this.libreGLMap.on('load', async () => {
                if (!this.libreGLMap) {
                    return;
                }

                await addIcons(this.libreGLMap as Map);
                this.setVisibilityOfAllControls();
            });

            this.libreGLMap.dragPan.enable({
                linearity: 0.01,
                maxSpeed: 1400,
                deceleration: 2500,
            });
        },
    },
});
</script>

<style>
.infrastructure-map {
    padding: 0;
    margin: 0;
    position: absolute;
    height: 100%;
    width: 100%;
}

.infrastructure-tooltip {
    display: none;
    left: 0;
    top: 0;
    background: white;
    border: 2px;
    border-radius: 5px;
}

.map-overlay {
    position: absolute;
    bottom: 0;
    right: 0;
    margin-right: 20px;
}
</style>

<style href="..e-gl.css" rel="stylesheet" />
<style href="..re.css" rel="stylesheet" />
