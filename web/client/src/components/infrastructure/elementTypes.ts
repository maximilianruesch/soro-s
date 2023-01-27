export const elementTypeLabels = {
    'bumper': 'Bumper',
    'border': 'Border',
    'track_end': 'Track End',
    'simple_switch': 'Switch',
    'as': 'Approach Signal',
    'ms': 'Main Signal',
    'ps': 'Protection Signal',
    'eotd': 'End of Train Detector',
    'spl': 'Speed Limit',
    'tunnel': 'Tunnel',
    'hlt': 'Halt',
    'rtcp': 'RTCP',
    'km_jump': 'KM Jump',
    'line_switch': 'Line Switch',
    'slope': 'Slope',
    'cross': 'Cross',
    'ctc': 'CTC',
    'station': 'Station'
};

export enum ElementType {
    BUMPER = 'bumper',
    BORDER = 'border',
    TRACK_END = 'track_end',
    SIMPLE_SWITCH = 'simple_switch',
}

export const elementTypes = Object.values(elementTypeLabels);

export const elementTypesWithLabels = Object.values(ElementType);


