package mapper_test

import (
	"testing"
	"transform-osm/db-utils/mapper"
	osmUtils "transform-osm/osm-utils"

	"github.com/stretchr/testify/assert"
)

func TestFindAndMapAnchorSwitches(t *testing.T) {
	type args struct {
		knoten           mapper.Spurplanknoten
		osm              *osmUtils.Osm
		anchors          map[float64][]*osmUtils.Node
		notFoundSwitches *[]*mapper.Weichenanfang
		foundAnchorCount *int
		nodeIdCounter    *int
	}
	type want struct {
		isError          bool
		notFoundSwitches int
		foundAnchors     map[float64][]*osmUtils.Node
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "find and map anchor switches",
			args: args{
				knoten: mapper.Spurplanknoten{
					WeichenAnf: []*mapper.Weichenanfang{
						{
							Name: mapper.Wert{
								Value: "Weiche 1",
							},
							KnotenTyp: mapper.KnotenTyp{
								Kilometrierung: mapper.Wert{
									Value: "1.000",
								},
							},
						},
						{
							Name: mapper.Wert{
								Value: "Weiche 2",
							},
							KnotenTyp: mapper.KnotenTyp{
								Kilometrierung: mapper.Wert{
									Value: "2.000",
								},
							},
						},
					},
				},
				osm: &osmUtils.Osm{
					Node: []*osmUtils.Node{
						{
							Id:  "10",
							Lat: "1.0",
							Lon: "1.0",
							Tag: []*osmUtils.Tag{
								{
									K: "railway",
									V: "switch",
								},
								{
									K: "ref",
									V: "Weiche 1",
								},
							},
						},
						{
							Id:  "11",
							Lat: "2.0",
							Lon: "2.0",
							Tag: []*osmUtils.Tag{
								{
									K: "railway",
									V: "switch",
								},
								{
									K: "ref",
									V: "Weiche 2",
								},
							},
						},
					},
				},
				anchors:          map[float64][]*osmUtils.Node{},
				notFoundSwitches: &[]*mapper.Weichenanfang{},
				foundAnchorCount: new(int),
				nodeIdCounter:    new(int),
			},
			want: want{
				isError:          false,
				notFoundSwitches: 0,
				foundAnchors: map[float64][]*osmUtils.Node{
					1: {
						{
							Id:  "1",
							Lat: "1.0",
							Lon: "1.0",
							Tag: []*osmUtils.Tag{
								{
									K: "railway",
									V: "switch",
								},
								{
									K: "ref",
									V: "Weiche 1",
								},
							},
						},
					},
					2: {
						{
							Id:  "2",
							Lat: "2.0",
							Lon: "2.0",
							Tag: []*osmUtils.Tag{
								{
									K: "railway",
									V: "switch",
								},
								{
									K: "ref",
									V: "Weiche 2",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mapper.FindAndMapAnchorSwitches(
				tt.args.knoten,
				tt.args.osm,
				tt.args.anchors,
				tt.args.notFoundSwitches,
				tt.args.foundAnchorCount,
				tt.args.nodeIdCounter,
			)
			if tt.want.isError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want.notFoundSwitches, len(*tt.args.notFoundSwitches))
			assert.Equal(t, len(tt.want.foundAnchors), *tt.args.foundAnchorCount)
		})
	}
}

func TestMapUnanchoredSwitches(t *testing.T) {
	// TODO
}

func TestMapCrosses(t *testing.T) {
	// TODO
}
