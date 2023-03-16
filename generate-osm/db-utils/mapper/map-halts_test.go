package mapper_test

import (
	"testing"
	"transform-osm/db-utils/mapper"
	osmUtils "transform-osm/osm-utils"

	"github.com/stretchr/testify/assert"
)

func TestMapHalts(t *testing.T) {
	type args struct {
		osm              *osmUtils.Osm
		anchor           map[float64]([]*osmUtils.Node)
		haltList         map[string]osmUtils.Halt
		nodeIdCounter    *int
		knoten           mapper.Spurplanknoten
		elementsNotFound map[string]([]string)
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "find one halt",
			args: args{
				osm: &osmUtils.Osm{
					Node: []*osmUtils.Node{
						{
							Id:  "10",
							Lat: "1.0",
							Lon: "1.0",
						},
						{
							Id:  "11",
							Lat: "2.0",
							Lon: "2.0",
						},
						{
							Id:  "12",
							Lat: "3.0",
							Lon: "3.0",
						},
						{
							Id:  "13",
							Lat: "4.0",
							Lon: "4.0",
						},
					},
					Way: []*osmUtils.Way{
						{
							Id: "100",
							Nd: []*osmUtils.Nd{
								{
									Ref: "10",
								},
								{
									Ref: "11",
								},
							},
						},
						{
							Id: "101",
							Nd: []*osmUtils.Nd{
								{
									Ref: "11",
								},
								{
									Ref: "12",
								},
							},
						},
						{
							Id: "102",
							Nd: []*osmUtils.Nd{
								{
									Ref: "12",
								},
								{
									Ref: "13",
								},
							},
						},
					},
				},
				anchor: map[float64]([]*osmUtils.Node){
					1.0: {
						{
							Id:  "11",
							Lat: "1.0",
							Lon: "1.0",
						},
					},
					3.0: {
						{
							Id:  "12",
							Lat: "2.0",
							Lon: "2.0",
						},
					},
				},
				haltList:      map[string]osmUtils.Halt{},
				nodeIdCounter: new(int),
				knoten: mapper.Spurplanknoten{
					HalteplGzF: []*mapper.NamedSimpleElement{
						{
							Name: mapper.Wert{
								Value: "Halt 1",
							},
							KnotenTyp: mapper.KnotenTyp{
								Kilometrierung: mapper.Wert{
									Value: "2.0",
								},
							},
						},
					},
				},
				elementsNotFound: map[string]([]string){},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mapper.MapHalts(tt.args.osm, tt.args.anchor, tt.args.haltList, tt.args.nodeIdCounter, tt.args.knoten, tt.args.elementsNotFound)
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			assert.Equal(t, 1, len(tt.args.haltList))
			assert.Equal(t, "Halt 1", tt.args.haltList["3"].Name)

		})
	}
}
