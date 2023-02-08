package DBMapper 

import (
	"encoding/xml"
	"os"
	"log"
	"strconv"
	"strings"
	"fmt"
	"math"
	DBUtil "transform-osm/db-utils"
	OSMUtil "transform-osm/osm-utils"
)

var TagName = xml.Name{" ", "tag"}
var id_counter = 1

func MapDB(refs []string, osmDir string, DBDir string) {
	for _, line := range refs {	
		var mappedItems = make(map[string]([]*OSMUtil.Node))

		var findAnchors = true
		
		var osmData OSMUtil.Osm
		var dbData DBUtil.XmlIssDaten

		osm_file, err := os.ReadFile(osmDir+"/"+line+".xml")
		if err != nil {
			log.Fatal(err)
		}
		db_file, err := os.ReadFile(DBDir+"/"+line+"_DB.xml")
		if err != nil {
			log.Fatal(err)
		}

		if err := xml.Unmarshal([]byte(osm_file), &osmData); err != nil { 
			panic(err)	
		}
		if err := xml.Unmarshal([]byte(db_file), &dbData); err != nil { 
			panic(err)	
		}

		print("Processing line ")
		print(line)
		print("\n")

		mainF, mainS := mapSignals(&osmData, dbData, &mappedItems, findAnchors)
		// mapPoints(&osmData, dbData, &mappedItems)		

		var restData = DBUtil.XmlIssDaten{
			Betriebsstellen: []*DBUtil.Spurplanbetriebsstelle{
				&DBUtil.Spurplanbetriebsstelle{
					Abschnitte: []*DBUtil.Spurplanabschnitt{
						&DBUtil.Spurplanabschnitt{
							Knoten: []*DBUtil.Spurplanknoten{
								&DBUtil.Spurplanknoten{
									HauptsigF: mainF,
									HauptsigS: mainS } } } } } } }

		findAnchors = false

		mainF, mainS = mapSignals(&osmData, restData, &mappedItems, findAnchors) // TODO: Different function for searching signals!
		// mapPoints(&osmData, restData, &mappedItems)
		// mapRest(&osmData, dbData, &mappedItems) 

		if new_Data, err := xml.MarshalIndent(osmData, "", "	"); err != nil {
			panic(err)
		} else {
			if err := os.WriteFile(osmDir+"/"+line+".xml", []byte(xml.Header + string(new_Data)), 0644); err != nil {
				panic(err)
			}
		}
	}
}

func mapSignals(OSMData *OSMUtil.Osm, DBData DBUtil.XmlIssDaten, anchors *map[string]([]*OSMUtil.Node), firstPass bool) ([]*DBUtil.Signal, []*DBUtil.Signal){
	var main_sigF []*DBUtil.Signal = []*DBUtil.Signal{}
	var main_sigS []*DBUtil.Signal = []*DBUtil.Signal{}
	for _, stelle := range DBData.Betriebsstellen {
		for _, abschnitt := range stelle.Abschnitte {
			for _, knoten := range abschnitt.Knoten {
				if firstPass {
					main_sigF = append(main_sigF, processHauptsigF(*knoten, OSMData, anchors)...)	
					main_sigS = append(main_sigS, processHauptsigS(*knoten, OSMData, anchors)...)
				} else {
					searchHauptsigF(*knoten, OSMData, anchors)
				}
				
			}
		}
	}

	return main_sigF, main_sigS
}

func processHauptsigF(knoten DBUtil.Spurplanknoten, OSMData *OSMUtil.Osm, anchors *map[string]([]*OSMUtil.Node)) []*DBUtil.Signal {
	var notFound = []*DBUtil.Signal{}
	for _, signal := range knoten.HauptsigF {
		found := false
		mismatch := false
		kilometrage := signal.KnotenTyp.Kilometrierung[0].Value
		if node_list := (*anchors)[kilometrage]; node_list != nil {
			for _, node := range node_list {
				typ, err1 := OSMUtil.GetTag(*node, "type")
				subtyp, err2 := OSMUtil.GetTag(*node, "subtype")
				id, err3 := OSMUtil.GetTag(*node, "id")
				direction, err4 := OSMUtil.GetTag(*node, "direction")
				if err1 == nil && typ == "element" && err2 == nil && subtyp == "ms" && err3 == nil && id == signal.Name[0].Value && err4 == nil && direction == "falling" {
					found = true
					break
				}
			}
		}		

		if found {
			continue
		}

		for i, node := range OSMData.Node {
			if len(node.Tag) == 0 {
				continue
			}

			is_signal := false 
			has_correct_id := false
			if railwayTag, err := OSMUtil.GetTag(*node, "railway"); err == nil && railwayTag == "signal" {
				is_signal = true
			}
			if idTag, err := OSMUtil.GetTag(*node, "ref"); err == nil && strings.ReplaceAll(idTag, " ", "") == signal.Name[0].Value {
				has_correct_id = true
			}

			if is_signal && has_correct_id {
				for key, value_list := range (*anchors) {
					for _, value := range value_list {
						if value != node {
							continue
						}
						if key == kilometrage {
							newNode := OSMUtil.Node{Id: strconv.Itoa(id_counter), Lat: node.Lat, Lon: node.Lon, Tag:[]*OSMUtil.Tag{
										&OSMUtil.Tag{TagName, "type", "element"}, 
										&OSMUtil.Tag{TagName, "subtype", "ms"}, 
										&OSMUtil.Tag{TagName, "id", signal.Name[0].Value},
										&OSMUtil.Tag{TagName, "direction", "falling"}}}
							OSMUtil.InsertNode(&newNode, node.Id, OSMData)
							id_counter++
							(*anchors)[key] = append((*anchors)[key], &newNode)
							found = true
							break
						} else {
							for _, error_val := range value_list {
								notFound = append(notFound, &DBUtil.Signal{
									DBUtil.KnotenTyp{Kilometrierung: []*DBUtil.Wert{&DBUtil.Wert{Value: key}}}, 
									[]*DBUtil.Wert{&DBUtil.Wert{Value: signal.Name[0].Value}}})
								error_val.Tag = error_val.Tag[:(len(error_val.Tag)-4)]
							}
							delete((*anchors), key)
							mismatch = true
							break
						}
					}
				}
				if !found && !mismatch {
					tags := &OSMData.Node[i].Tag 
					*tags = append(*tags, []*OSMUtil.Tag{
							&OSMUtil.Tag{TagName, "type", "element"}, 
							&OSMUtil.Tag{TagName, "subtype", "ms"}, 
							&OSMUtil.Tag{TagName, "id", signal.Name[0].Value},
							&OSMUtil.Tag{TagName, "direction", "falling"}}...)
					if len((*anchors)[kilometrage]) == 0 {
						(*anchors)[kilometrage] = []*OSMUtil.Node{OSMData.Node[i]}
					} else {
						(*anchors)[kilometrage] = append((*anchors)[kilometrage], OSMData.Node[i])
					}						
					found = true
				}
			} 
		}		

		if !found {
			notFound = append(notFound, signal)
		}
	}

	return notFound	
}

func processHauptsigS(knoten DBUtil.Spurplanknoten, OSMData *OSMUtil.Osm, anchors *map[string]([]*OSMUtil.Node)) []*DBUtil.Signal {
	var notFound = []*DBUtil.Signal{}
	for _, signal := range knoten.HauptsigS {
		found := false
		mismatch := false
		kilometrage := signal.KnotenTyp.Kilometrierung[0].Value
		if node_list := (*anchors)[kilometrage]; node_list != nil {
			for _, node := range node_list {
				typ, err1 := OSMUtil.GetTag(*node, "type")
				subtyp, err2 := OSMUtil.GetTag(*node, "subtype")
				id, err3 := OSMUtil.GetTag(*node, "id")
				direction, err4 := OSMUtil.GetTag(*node, "direction")
				if err1 == nil && typ == "element" && err2 == nil && subtyp == "ms" && err3 == nil && id == signal.Name[0].Value && err4 == nil && direction == "rising" {
					found = true
					break
				}
			}
		}		

		if found {
			continue
		}

		for i, node := range OSMData.Node {
			if len(node.Tag) == 0 {
				continue
			}

			is_signal := false 
			has_correct_id := false
			if railwayTag, err := OSMUtil.GetTag(*node, "railway"); err == nil && railwayTag == "signal" {
				is_signal = true
			}
			if idTag, err := OSMUtil.GetTag(*node, "ref"); err == nil && strings.ReplaceAll(idTag, " ", "") == signal.Name[0].Value {
				has_correct_id = true
			}

			if is_signal && has_correct_id {
				for key, value_list := range (*anchors) {
					for _, value := range value_list {
						if value != node {
							continue
						}
						if key == kilometrage {
							newNode := OSMUtil.Node{Id: strconv.Itoa(id_counter), Lat: node.Lat, Lon: node.Lon, Tag:[]*OSMUtil.Tag{
										&OSMUtil.Tag{TagName, "type", "element"}, 
										&OSMUtil.Tag{TagName, "subtype", "ms"}, 
										&OSMUtil.Tag{TagName, "id", signal.Name[0].Value},
										&OSMUtil.Tag{TagName, "direction", "rising"}}}
							OSMUtil.InsertNode(&newNode, node.Id, OSMData)
							id_counter++
							(*anchors)[key] = append((*anchors)[key], &newNode)
							found = true
							break
						} else {
							for _, error_val := range value_list {
								notFound = append(notFound, &DBUtil.Signal{
									DBUtil.KnotenTyp{Kilometrierung: []*DBUtil.Wert{&DBUtil.Wert{Value: key}}}, 
									[]*DBUtil.Wert{&DBUtil.Wert{Value: signal.Name[0].Value}}})
								error_val.Tag = error_val.Tag[:(len(error_val.Tag)-4)]
							}
							delete((*anchors), key)
							mismatch = true
							break
						}
					}
				}
				if !found && !mismatch {
					tags := &OSMData.Node[i].Tag 
					*tags = append(*tags, []*OSMUtil.Tag{
							&OSMUtil.Tag{TagName, "type", "element"}, 
							&OSMUtil.Tag{TagName, "subtype", "ms"}, 
							&OSMUtil.Tag{TagName, "id", signal.Name[0].Value},
							&OSMUtil.Tag{TagName, "direction", "rising"}}...)
					if len((*anchors)[kilometrage]) == 0 {
						(*anchors)[kilometrage] = []*OSMUtil.Node{OSMData.Node[i]}
					} else {
						(*anchors)[kilometrage] = append((*anchors)[kilometrage], OSMData.Node[i])
					}						
					found = true
				}
			} 
		}		

		if !found {
			notFound = append(notFound, signal)
		}
	}

	return notFound	
}

func searchHauptsigF(knoten DBUtil.Spurplanknoten, OSMData *OSMUtil.Osm, anchors *map[string]([]*OSMUtil.Node)) {
	switch len(*anchors) {
	case 0:
		fmt.Print("Could not find anchors! \n")
	case 1:
		// TODO: Try to find a node and guess the correct one, or throw error also
	default:
		for _, signal := range knoten.HauptsigF {
			nearest := -1.0
			second_nearest := -1.0
			kilometrage, _ := strconv.ParseFloat(strings.ReplaceAll(signal.KnotenTyp.Kilometrierung[0].Value, ",", "."), 64)

			for key, _ := range *anchors {
				if strings.Contains(key, "+") {
					continue
				}
				float_key, _ := strconv.ParseFloat(strings.ReplaceAll(key, ",", "."), 64)
				if nearest == -1.0 {
					nearest = float_key
				}
				if math.Abs(float_key - kilometrage) < math.Abs(nearest - kilometrage) {
					second_nearest = nearest
					nearest = float_key
				}
			}	

			if second_nearest == -1.0  {
				for key, _ := range *anchors {
					if strings.Contains(key, "+") {
						continue
					}
					float_key, _ := strconv.ParseFloat(strings.ReplaceAll(key, ",", "."), 64)
					if float_key == nearest {
						continue
					}
					if second_nearest == -1.0 {
						second_nearest = float_key
					}
					if math.Abs(float_key - kilometrage) < math.Abs(second_nearest - kilometrage) {
						second_nearest = float_key
					}
				}	
			}

			if nearest == -1.0 || second_nearest == -1.0 {
				continue
			}

			nearest_string := strings.ReplaceAll(strconv.FormatFloat(nearest, 'f', -1, 64), ".", ",")
			second_nearest_string := strings.ReplaceAll(strconv.FormatFloat(second_nearest, 'f', -1, 64), ".", ",")

			for ; len((*anchors)[nearest_string]) == 0; nearest_string += "0" {
				if !strings.Contains(nearest_string, ",") {
					nearest_string += ","
				}
			}
			for ; len((*anchors)[second_nearest_string]) == 0; second_nearest_string += "0" {
				if !strings.Contains(second_nearest_string, ",") {
					second_nearest_string += ","
				}
			}	
			
			newLat, newLon := findNewCoordinates(*((*anchors)[nearest_string])[0], *((*anchors)[second_nearest_string])[0], nearest, second_nearest, kilometrage)
			fmt.Printf("%f, %f \n \n", newLat, newLon)
		}
	}	
	// TODO: Node not found, find closest mapped Node and work from there
}

func findNewCoordinates(nearestNode OSMUtil.Node, second_nearestNode OSMUtil.Node, nearest float64, second_nearest float64, kilometrage float64) (float64, float64) {
	dist1 := float64(math.Abs(nearest - kilometrage))	
	nearest_Lat, _ := strconv.ParseFloat(nearestNode.Lat, 64)
	nearest_Lon, _ := strconv.ParseFloat(nearestNode.Lon, 64)
	if dist1 == 0.0 {
		return nearest_Lat, nearest_Lon
	}

	dist2 := float64(math.Abs(second_nearest - kilometrage))
	second_nearest_Lat, _ := strconv.ParseFloat(second_nearestNode.Lat, 64)
	second_nearest_Lon, _ := strconv.ParseFloat(second_nearestNode.Lon, 64)
	
	diffx, diffy := second_nearest_Lat - nearest_Lat, second_nearest_Lon - nearest_Lon
	length := math.Sqrt(diffx*diffx + diffy*diffy)
	diffx, diffy = diffx/length, diffy/length

	var newLat, newLon float64

	if length > dist2 {
		newLat, newLon = nearest_Lat + diffx*dist1, nearest_Lon + diffy*dist1
	} else {
		newLat, newLon = nearest_Lat + diffx*dist1, nearest_Lon + diffy*dist1
	}

	fmt.Printf("%f, %f \n", nearest_Lat, nearest_Lon)
	fmt.Printf("%f, %f \n", second_nearest_Lat, second_nearest_Lon)

	return newLat, newLon
}