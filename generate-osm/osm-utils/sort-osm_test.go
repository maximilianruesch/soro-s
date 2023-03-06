package osmUtils_test

import (
	"testing"
	osmUtils "transform-osm/osm-utils"
)

func TestSortAndRemoveDuplicates(t *testing.T) {
	type args struct {
		osm *osmUtils.Osm
	}
	tests := []struct {
		name string
		args args
		want *osmUtils.Osm
	}{
		{
			name: "test sort and remove duplicates",
			args: args{
				osm: &osmUtils.Osm{
					Relation: []*osmUtils.Relation{
						{
							Id: "2",
						},
						{
							Id: "1",
						},
						{
							Id: "1",
						},
					},
					Way: []*osmUtils.Way{
						{
							Id: "2",
						},
						{
							Id: "1",
						},
						{
							Id: "1",
						},
					},
					Node: []*osmUtils.Node{
						{
							Id: "2",
						},
						{
							Id: "1",
						},
						{
							Id: "1",
						},
					},
				},
			},
			want: &osmUtils.Osm{
				Relation: []*osmUtils.Relation{
					{
						Id: "1",
					},
					{
						Id: "2",
					},
				},
				Node: []*osmUtils.Node{
					{
						Id: "1",
					},
					{
						Id: "2",
					},
				},
				Way: []*osmUtils.Way{
					{
						Id: "1",
					},
					{
						Id: "2",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			osmUtils.SortAndRemoveDuplicatesOsm(tt.args.osm)
			for i := 0; i < len(tt.args.osm.Relation); i++ {
				if tt.args.osm.Relation[i].Id != tt.want.Relation[i].Id {
					t.Errorf("SortAndRemoveDuplicates() = %v, want %v", tt.args.osm.Relation[i].Id, tt.want.Relation[i].Id)
				}
			}

			for i := 0; i < len(tt.args.osm.Way); i++ {
				if tt.args.osm.Way[i].Id != tt.want.Way[i].Id {
					t.Errorf("SortAndRemoveDuplicates() = %v, want %v", tt.args.osm.Way[i].Id, tt.want.Way[i].Id)
				}
			}

			for i := 0; i < len(tt.args.osm.Node); i++ {
				if tt.args.osm.Node[i].Id != tt.want.Node[i].Id {
					t.Errorf("SortAndRemoveDuplicates() = %v, want %v", tt.args.osm.Node[i].Id, tt.want.Node[i].Id)
				}
			}
		})
	}
}
