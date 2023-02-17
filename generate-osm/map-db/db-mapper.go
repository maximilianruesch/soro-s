package DBMapper 

import (
	"encoding/xml"
	"os"
	"log"
	"strconv"
	"strings"
	"errors"
	"fmt"
	"math"
	DBUtil "transform-osm/db-utils"
	OSMUtil "transform-osm/osm-utils"
)

var TagName = xml.Name{" ", "tag"}
var id_counter = 1
var anchors map[string]([]*OSMUtil.Node)
var osmData OSMUtil.Osm
var num_found = 0

func MapDB(refs []string, osmDir string, DBDir string) {
	for _, line := range refs {	
		anchors = make(map[string]([]*OSMUtil.Node))

		var findAnchors = true
		var dbData DBUtil.XmlIssDaten
		osmData = OSMUtil.Osm{}

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

		mainF, mainS := mapSignals(dbData, findAnchors)
		// mapPoints(&osmData, dbData)		

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

		mainF, mainS = mapSignals(restData, findAnchors) 
		// mapPoints(restData)
		// mapRest(dbData) 		

		if new_Data, err := xml.MarshalIndent(osmData, "", "	"); err != nil {
			panic(err)
		} else {
			if err := os.WriteFile(osmDir+"/"+line+".xml", []byte(xml.Header + string(new_Data)), 0644); err != nil {
				panic(err)
			}
		}
	}
	fmt.Printf("Found: %d \n", num_found)
}

func mapSignals(DBData DBUtil.XmlIssDaten, firstPass bool) ([]*DBUtil.Signal, []*DBUtil.Signal){
	var main_sigF []*DBUtil.Signal = []*DBUtil.Signal{}
	var main_sigS []*DBUtil.Signal = []*DBUtil.Signal{}
	for _, stelle := range DBData.Betriebsstellen {
		for _, abschnitt := range stelle.Abschnitte {
			for _, knoten := range abschnitt.Knoten {
				if firstPass {
					main_sigF = append(main_sigF, processHauptsigF(*knoten)...)	
					main_sigS = append(main_sigS, processHauptsigS(*knoten)...)
				} else {
					searchHauptsigF(*knoten)
				}
				
			}
		}
	}

	return main_sigF, main_sigS
}

func processHauptsigF(knoten DBUtil.Spurplanknoten) []*DBUtil.Signal {
	var notFound = []*DBUtil.Signal{}
	for _, signal := range knoten.HauptsigF {
		found := false
		kilometrage := signal.KnotenTyp.Kilometrierung[0].Value
		if node_list := anchors[kilometrage]; node_list != nil {
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

		for _, node := range osmData.Node {
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
				insertNewHauptsig(node, kilometrage, *signal, "falling", &notFound)
			} 
		}		

		if !found {
			notFound = append(notFound, signal)
		} 
	}

	return notFound	
}

func processHauptsigS(knoten DBUtil.Spurplanknoten) []*DBUtil.Signal {
	var notFound = []*DBUtil.Signal{}
	for _, signal := range knoten.HauptsigS {
		found := false
		kilometrage := signal.KnotenTyp.Kilometrierung[0].Value
		if node_list := anchors[kilometrage]; node_list != nil {
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

		for _, node := range osmData.Node {
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
				found = insertNewHauptsig(node, kilometrage, *signal, "rising", &notFound)
			} 
		}		

		if !found {
			notFound = append(notFound, signal)
		}
	}

	return notFound	
}

func searchHauptsigF(knoten DBUtil.Spurplanknoten) {
	var not_found = []*DBUtil.Signal{} 
	switch len(anchors) {
	case 0:
		fmt.Print("Could not find anchors! \n")
	case 1:
		fmt.Print("Could not find enough anchors! \n")
	default:
		for _, signal := range knoten.HauptsigF {
			kilometrage, _ := strconv.ParseFloat(strings.ReplaceAll(signal.KnotenTyp.Kilometrierung[0].Value, ",", "."), 64)

			maxNode, err := findBestOSMNode(kilometrage)
			if err != nil {
				not_found = append(not_found, signal)
				continue
			}			

			found := insertNewHauptsig(maxNode, signal.KnotenTyp.Kilometrierung[0].Value, *signal, "falling", &not_found)	
			if found {
				//fmt.Printf("Found node %s \n", maxNode.Id)
				num_found++
			}
		}
	}	
	// TODO: Node not found, find closest mapped Node and work from there
}

func findTwoNearest(kilometrage float64) (nearest float64, second_nearest float64) {	
	nearest = -1.0
	second_nearest = -1.0

	for key, _ := range anchors {
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

	if second_nearest != -1.0  { 
		return
	}
	for key, _ := range anchors {
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
	return
}

func formatKilometrage(in float64) (out string) {
	out = strings.ReplaceAll(strconv.FormatFloat(in, 'f', -1, 64), ".", ",")

	for ; len(anchors[out]) == 0; out += "0" {
		if !strings.Contains(out, ",") {
			out += ","
		}
	}
	return
}

func findBestOSMNode(kilometrage float64) (*OSMUtil.Node, error){
	nearest, second_nearest := findTwoNearest(kilometrage)

	if nearest == -1.0 || second_nearest == -1.0 {
		return nil, errors.New("Could not find node.")
	}

	nearest_string := formatKilometrage(nearest)
	second_nearest_string := formatKilometrage(second_nearest)

	newNode, err := DBUtil.FindNewNode((anchors[nearest_string])[0], (anchors[second_nearest_string])[0], math.Abs(nearest - kilometrage), math.Abs(second_nearest - kilometrage), &osmData)
	if err != nil {
		return nil, errors.New("Could not find node.")
	}

	return newNode, nil
}

func insertNewHauptsig(node *OSMUtil.Node, kilometrage string, signal DBUtil.Signal, direction string, notFound *[]*DBUtil.Signal) bool{
	for key, value_list := range anchors {
		for _, value := range value_list {
			if value != node {
				continue
			}
			if key == kilometrage {
				newNode := OSMUtil.Node{Id: strconv.Itoa(id_counter), Lat: node.Lat, Lon: node.Lon, Tag:[]*OSMUtil.Tag{
							&OSMUtil.Tag{TagName, "type", "element"}, 
							&OSMUtil.Tag{TagName, "subtype", "ms"}, 
							&OSMUtil.Tag{TagName, "id", signal.Name[0].Value},
							&OSMUtil.Tag{TagName, "direction", direction}}}
				OSMUtil.InsertNode(&newNode, node.Id, &osmData)
				id_counter++
				anchors[key] = append(anchors[key], &newNode)
				return true
			} else {
				for _, error_val := range value_list {
					*notFound = append(*notFound, &DBUtil.Signal{
						DBUtil.KnotenTyp{Kilometrierung: []*DBUtil.Wert{&DBUtil.Wert{Value: key}}}, 
						[]*DBUtil.Wert{&DBUtil.Wert{Value: signal.Name[0].Value}}})
					error_val.Tag = error_val.Tag[:(len(error_val.Tag)-4)]
				}
				delete(anchors, key)
				return false
			}
		}
	}

	node.Tag = append(node.Tag, []*OSMUtil.Tag{
			&OSMUtil.Tag{TagName, "type", "element"}, 
			&OSMUtil.Tag{TagName, "subtype", "ms"}, 
			&OSMUtil.Tag{TagName, "id", signal.Name[0].Value},
			&OSMUtil.Tag{TagName, "direction", direction}}...)
	if len(anchors[kilometrage]) == 0 {
		anchors[kilometrage] = []*OSMUtil.Node{node}
	} else {
		anchors[kilometrage] = append(anchors[kilometrage], node)
	}						
	return true
}