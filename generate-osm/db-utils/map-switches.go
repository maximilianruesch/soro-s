package dbUtils

import (
	OSMUtil "transform-osm/osm-utils"
)

func findAndMapAnchorSwitches(
	abschnitt *Spurplanabschnitt,
	osm *OSMUtil.Osm,
	anchors map[string][]*OSMUtil.Node,
	foundAchnorCount *int,
) {
	for _, knoten := range abschnitt.Knoten {
		for _, switchBegin := range knoten.WeichenAnf {
			for _, node := range osm.Node {
				if len(node.Tag) == 0 {
					continue
				}

				railwayTag, _ := OSMUtil.FindTagOnNode(node, "railway")
				refTag, _ := OSMUtil.FindTagOnNode(node, "ref")
				if railwayTag == "switch" && refTag == switchBegin.Name.Value {
					anchors[switchBegin.Kilometrierung.Value] = append(anchors[switchBegin.Kilometrierung.Value], node)
					node.Tag = append(node.Tag, []*OSMUtil.Tag{
						{XMLName: XML_TAG_NAME_CONST, K: "type", V: "element"},
						{XMLName: XML_TAG_NAME_CONST, K: "subtype", V: "simple_switch"},
						{XMLName: XML_TAG_NAME_CONST, K: "id", V: refTag},
					}...)
					*foundAchnorCount++
				}
			}
		}

		restSwitches := make([]*Weichenknoten, len(knoten.WeichenStamm)+len(knoten.WeichenAbzwLinks)+len(knoten.WeichenAbzwRechts))
		copy(restSwitches, knoten.WeichenStamm)
		copy(restSwitches[len(knoten.WeichenStamm):], knoten.WeichenAbzwLinks)
		copy(restSwitches[len(knoten.WeichenStamm)+len(knoten.WeichenAbzwLinks):], knoten.WeichenAbzwRechts)
		for _, switchBegin := range restSwitches {
			for _, node := range osm.Node {
				if len(node.Tag) == 0 {
					continue
				}

				railwayTag, _ := OSMUtil.FindTagOnNode(node, "railway")
				refTag, _ := OSMUtil.FindTagOnNode(node, "ref")

				if railwayTag == "switch" {
					partnerName := switchBegin.Partner.Name

					if partnerName == refTag && anchors[switchBegin.Kilometrierung.Value] == nil {
						anchors[switchBegin.Kilometrierung.Value] = append(anchors[switchBegin.Kilometrierung.Value], node)
						*foundAchnorCount++
					}
				}
			}
		}
	}
}
