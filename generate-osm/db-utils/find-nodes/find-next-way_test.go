package findNodes_test

import (
	"testing"
	findNodes "transform-osm/db-utils/find-nodes"
	osmUtils "transform-osm/osm-utils"

	"github.com/stretchr/testify/assert"
)

func TestFindNextWay(t *testing.T) {
	type args struct {
		osm         *osmUtils.Osm
		wayDirUp    bool
		index       int
		runningNode *osmUtils.Node
		oldNode     *osmUtils.Node
		runningWay  osmUtils.Way
	}
	type want struct {
		nextWay  osmUtils.Way
		newIndex int
		wayDirUp bool
		errNil   bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "find next way",
			args: args{
				osm: &osmUtils.Osm{
					Way: []*osmUtils.Way{
						{
							Id: "1",
							Nd: []*osmUtils.Nd{
								{
									Ref: "1",
								},
								{
									Ref: "2",
								},
							},
						},
						{
							Id: "2",
							Nd: []*osmUtils.Nd{
								{
									Ref: "2",
								},
								{
									Ref: "3",
								},
							},
						},
						{
							Id: "3",
							Nd: []*osmUtils.Nd{
								{
									Ref: "3",
								},
								{
									Ref: "4",
								},
							},
						},
						{
							Id: "4",
							Nd: []*osmUtils.Nd{
								{
									Ref: "4",
								},
								{
									Ref: "5",
								},
							},
						},
					},
				},
				wayDirUp: true,
				index:    0,
				runningNode: &osmUtils.Node{
					Id: "3",
				},
				oldNode: &osmUtils.Node{
					Id: "1",
				},
				runningWay: osmUtils.Way{
					Id: "1",
					Nd: []*osmUtils.Nd{
						{
							Ref: "1",
						},
						{
							Ref: "2",
						},
					},
				},
			},
			want: want{
				nextWay: osmUtils.Way{
					Id: "2",
					Nd: []*osmUtils.Nd{
						{
							Ref: "2",
						},
						{
							Ref: "3",
						},
					},
				},
				newIndex: 1,
				wayDirUp: true,
				errNil:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextWay, newIndex, wayDirUp, err := findNodes.FindNextWay(tt.args.osm, tt.args.wayDirUp, tt.args.index, tt.args.runningNode, tt.args.oldNode, tt.args.runningWay)
			if tt.want.errNil {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
			}
			assert.Equal(t, tt.want.nextWay, nextWay)
			assert.Equal(t, tt.want.newIndex, newIndex)
			assert.Equal(t, tt.want.wayDirUp, wayDirUp)
		})
	}

}

func TestGetBothCorrectWays(t *testing.T) {
	type args struct {
		osm           *osmUtils.Osm
		runningNodeId string
	}
	type want struct {
		firstWay  osmUtils.Way
		secondWay osmUtils.Way
		errNil    bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "get both correct ways",
			args: args{
				osm: &osmUtils.Osm{
					Way: []*osmUtils.Way{
						{
							Id: "1",
							Nd: []*osmUtils.Nd{
								{
									Ref: "1",
								},
								{
									Ref: "2",
								},
							},
						},
						{
							Id: "2",
							Nd: []*osmUtils.Nd{
								{
									Ref: "2",
								},
								{
									Ref: "3",
								},
							},
						},
						{
							Id: "3",
							Nd: []*osmUtils.Nd{
								{
									Ref: "3",
								},
								{
									Ref: "4",
								},
							},
						},
					},
				},
				runningNodeId: "2",
			},
			want: want{
				firstWay: osmUtils.Way{
					Id: "1",
					Nd: []*osmUtils.Nd{
						{
							Ref: "1",
						},
						{
							Ref: "2",
						},
					},
				},
				secondWay: osmUtils.Way{
					Id: "2",
					Nd: []*osmUtils.Nd{
						{
							Ref: "2",
						},
						{
							Ref: "3",
						},
					},
				},
				errNil: true,
			},
		},
		{
			name: "throws error if no way found",
			args: args{
				osm: &osmUtils.Osm{
					Way: []*osmUtils.Way{
						{
							Id: "1",
							Nd: []*osmUtils.Nd{
								{
									Ref: "1",
								},
								{
									Ref: "2",
								},
							},
						},
						{
							Id: "2",
							Nd: []*osmUtils.Nd{
								{
									Ref: "2",
								},
								{
									Ref: "3",
								},
							},
						},
						{
							Id: "3",
							Nd: []*osmUtils.Nd{
								{
									Ref: "3",
								},
								{
									Ref: "4",
								},
							},
						},
					},
				},
				runningNodeId: "5",
			},
			want: want{
				firstWay:  osmUtils.Way{},
				secondWay: osmUtils.Way{},
				errNil:    false,
			},
		},
		{
			name: "throws error if only one way found",
			args: args{
				osm: &osmUtils.Osm{
					Way: []*osmUtils.Way{
						{
							Id: "1",
							Nd: []*osmUtils.Nd{
								{
									Ref: "1",
								},
								{
									Ref: "2",
								},
							},
						},
						{
							Id: "2",
							Nd: []*osmUtils.Nd{
								{
									Ref: "2",
								},
								{
									Ref: "3",
								},
							},
						},
						{
							Id: "3",
							Nd: []*osmUtils.Nd{
								{
									Ref: "3",
								},
								{
									Ref: "4",
								},
							},
						},
					},
				},
				runningNodeId: "1",
			},
			want: want{
				firstWay:  osmUtils.Way{},
				secondWay: osmUtils.Way{},
				errNil:    false,
			},
		},
		{
			name: "throws error if more than two ways found",
			args: args{
				osm: &osmUtils.Osm{
					Way: []*osmUtils.Way{
						{
							Id: "1",
							Nd: []*osmUtils.Nd{
								{
									Ref: "1",
								},
								{
									Ref: "3",
								},
							},
						},
						{
							Id: "2",
							Nd: []*osmUtils.Nd{
								{
									Ref: "2",
								},
								{
									Ref: "3",
								},
							},
						},
						{
							Id: "3",
							Nd: []*osmUtils.Nd{
								{
									Ref: "3",
								},
								{
									Ref: "4",
								},
							},
						},
					},
				},
				runningNodeId: "3",
			},
			want: want{
				firstWay:  osmUtils.Way{},
				secondWay: osmUtils.Way{},
				errNil:    false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			firstWay, secondWay, err := findNodes.GetBothCorrectWays(tt.args.osm, tt.args.runningNodeId)
			assert.Equal(t, tt.want.firstWay, firstWay)
			assert.Equal(t, tt.want.secondWay, secondWay)
			if tt.want.errNil {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}
