import { mountWithDefaults } from '@/test-utils/mount-with-defaults';
import SoroNavigationMenuContent from './soro-navigation-menu-content.vue';
import { VueWrapper } from '@vue/test-utils';
import { GoldenLayoutNamespace } from '@/stores/golden-layout-store';
import { ComponentTechnicalName, GLComponentTitles } from '@/golden-layout/golden-layout-constants';
import { VExpansionPanels } from 'vuetify/components';

vi.mock('@/golden-layout/golden-layout-constants', () => ({
    ComponentTechnicalName: {
        INFRASTRUCTURE: 'infrastructure',
        ORDERING_GRAPH: 'ordering_graph',
    },
    GLComponentTitles: {
        infrastructure: 'some-infrastructure-title',
        ordering_graph: 'some-ordering-graph-title',
    },
}));

describe('soro-navigation-menu-content', async () => {
    let soroNavigationMenuContent: VueWrapper<any>;
    const goldenLayoutActions = { addGoldenLayoutTab: vi.fn() };

    const defaults = {
        store: {
            [GoldenLayoutNamespace]: { actions: goldenLayoutActions },
        },
    };

    beforeEach(async () => {
        goldenLayoutActions.addGoldenLayoutTab.mockImplementation(() => ({}));
        soroNavigationMenuContent = await mountWithDefaults(SoroNavigationMenuContent, defaults);
        vi.clearAllMocks();
        window.localStorage.clear();
    });

    it('displays several buttons to allow adding new golden layout tabs', async () => {
        const windowControls = soroNavigationMenuContent.find('.window-controls');

        const tabButtons = windowControls.findAllComponents({ name: 'soro-button' });
        tabButtons[0].vm.$emit('click');
        tabButtons[3].vm.$emit('click');

        expect(tabButtons).toHaveLength(4);
        // Buttons 2 and 3 should be disabled
        expect(tabButtons[1].attributes('disabled')).toBeDefined();
        expect(tabButtons[2].attributes('disabled')).toBeDefined();
        // Buttons 1 and 4 should have added golden layout tabs as in their click order above
        expect(goldenLayoutActions.addGoldenLayoutTab).toHaveBeenCalledTimes(2);
        expect(goldenLayoutActions.addGoldenLayoutTab).toHaveBeenNthCalledWith(
            1,
            expect.any(Object),
            {
                componentTechnicalName: ComponentTechnicalName.INFRASTRUCTURE,
                title: GLComponentTitles[ComponentTechnicalName.INFRASTRUCTURE],
            },
        );
        expect(goldenLayoutActions.addGoldenLayoutTab).toHaveBeenNthCalledWith(
            2,
            expect.any(Object),
            {
                componentTechnicalName: ComponentTechnicalName.ORDERING_GRAPH,
                title: GLComponentTitles[ComponentTechnicalName.ORDERING_GRAPH],
            },
        );
    });

    it('contains a station search with extended options', async () => {
        const stationSearch = soroNavigationMenuContent.findComponent({ name: 'station-search' });

        expect(stationSearch.exists()).toBe(true);
        expect(stationSearch.vm.$props).toStrictEqual(expect.objectContaining({
            showExtendedLink: true,
        }));
    });

    it('displays a dev tool to clear local storage', async () => {
        const devTools = soroNavigationMenuContent.findComponent<VExpansionPanels>('.dev-tools');
        window.localStorage.setItem('some-key', 'some-value');

        const devToolButtons = devTools.findAllComponents({ name: 'soro-button' });
        devToolButtons[0].vm.$emit('click');
        expect(window.localStorage).toHaveLength(0);
    });
});
