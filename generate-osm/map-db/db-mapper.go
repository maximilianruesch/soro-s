package DBMapper 

import (
	"encoding/xml"
	"os"
	"log"
	"strconv"
	"strings"
	//"math"
	DBUtil "transform-osm/db-utils"
	OSMUtil "transform-osm/osm-utils"
)

var TagName = xml.Name{" ", "tag"}
var id_counter = 1

func MapDB(refs []string, osmDir string, DBDir string) {
	for _, line := range refs {	
		var mappedItems = make(map[string](*OSMUtil.Node))
		
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

		mainF, mainS := mapSignals(&osmData, dbData, &mappedItems)
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

		mainF, mainS = mapSignals(&osmData, restData, &mappedItems) // TODO: Different function for searching signals!
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

func mapSignals(OSMData *OSMUtil.Osm, DBData DBUtil.XmlIssDaten, anchors *map[string](*OSMUtil.Node)) ([]*DBUtil.Signal, []*DBUtil.Signal){
	var main_sigF, main_sigS []*DBUtil.Signal
	for _, stelle := range DBData.Betriebsstellen {
		for _, abschnitt := range stelle.Abschnitte {
			for _, knoten := range abschnitt.Knoten {
				main_sigF = processHauptsigF(*knoten, OSMData, anchors)							
				print("final length: ")
				print(len(*anchors))
				print("\n")		
				main_sigS = processHauptsigS(*knoten, OSMData, anchors)
			}
		}
	}
	return main_sigF, main_sigS
}

func processHauptsigF(knoten DBUtil.Spurplanknoten, OSMData *OSMUtil.Osm, anchors *map[string](*OSMUtil.Node)) []*DBUtil.Signal {
	var notFound = []*DBUtil.Signal{}
	for _, signal := range knoten.HauptsigF {
		found := false
		kilometrage := signal.KnotenTyp.Kilometrierung[0].Value
		if node := (*anchors)[kilometrage]; node != nil {
			typ, err1 := OSMUtil.GetTag(*node, "type")
			subtyp, err2 := OSMUtil.GetTag(*node, "subtype")
			id, err3 := OSMUtil.GetTag(*node, "id")
			direction, err4 := OSMUtil.GetTag(*node, "direction")
			if err1 == nil && typ == "element" && err2 == nil && subtyp == "ms" && err3 == nil && id == signal.Name[0].Value && err4 == nil && direction == "falling" {
				continue
			}			
			newNode := OSMUtil.Node{Id: strconv.Itoa(id_counter), Lat: node.Lat, Lon: node.Lon, Tag:[]*OSMUtil.Tag{
				&OSMUtil.Tag{TagName, "type", "element"}, 
				&OSMUtil.Tag{TagName, "subtype", "ms"}, 
				&OSMUtil.Tag{TagName, "id", signal.Name[0].Value},
				&OSMUtil.Tag{TagName, "direction", "falling"}}}
			OSMUtil.InsertNode(&newNode, node.Id, OSMData)
			id_counter++
			continue 
		}

		for i, node := range OSMData.Node {
			if len(node.Tag) == 0 {
				continue
			}
			is_signal := false
			has_correct_id := false
			for _, tag := range (*node).Tag {
				if tag.K == "railway" && tag.V == "signal" {
					is_signal = true
				}	
				if tag.K == "ref" && strings.ReplaceAll(tag.V, " ", "") == signal.Name[0].Value {
					has_correct_id = true
				}	
			}

			if has_correct_id && is_signal {
				tags := &OSMData.Node[i].Tag 
				*tags = append(*tags, []*OSMUtil.Tag{
						&OSMUtil.Tag{TagName, "type", "element"}, 
						&OSMUtil.Tag{TagName, "subtype", "ms"}, 
						&OSMUtil.Tag{TagName, "id", signal.Name[0].Value},
						&OSMUtil.Tag{TagName, "direction", "falling"}}...)
				(*anchors)[kilometrage] = OSMData.Node[i]	
				print("Length before: ")
				print(len(*anchors))
				print("\n")
				found = true
				break
			} 
		}					
		print("Length after: ")
		print(len(*anchors))
		print("\n")					
		if found {
			continue
		}

		notFound = append(notFound, signal)

		/*
		nearest := -1.0
		for key, _ := range *anchors {
			if nearest == -1.0 {
				nearest = key
			}
			if math.Abs(key - kilometrage) < math.Abs(nearest - kilometrage) {
				nearest = key
			}
			print("Doing something... \n")
		}
		/*
		nearest_Lat, nearest_Lon := (*anchors)[nearest].Lat, (*anchors)[nearest].Lon
		dist := math.Abs(nearest - kilometrage)

		print(dist)
		*/
		// TODO: Node not found, find closest mapped Node and work from there
	}
	return notFound
}

func processHauptsigS(knoten DBUtil.Spurplanknoten, OSMData *OSMUtil.Osm, anchors *map[string](*OSMUtil.Node)) []*DBUtil.Signal {
	var notFound = []*DBUtil.Signal{}
	for _, signal := range knoten.HauptsigS {		
		found := false
		kilometrage := signal.KnotenTyp.Kilometrierung[0].Value
		if node := (*anchors)[kilometrage]; node != nil {
			typ, err1 := OSMUtil.GetTag(*node, "type")
			subtyp, err2 := OSMUtil.GetTag(*node, "subtype")
			id, err3 := OSMUtil.GetTag(*node, "id")
			direction, err4 := OSMUtil.GetTag(*node, "direction")
			if err1 == nil && typ == "element" && err2 == nil && subtyp == "ms" && err3 == nil && id == signal.Name[0].Value && err4 == nil && direction == "rising" {
				continue
			}			
			newNode := OSMUtil.Node{Id: strconv.Itoa(id_counter), Lat: node.Lat, Lon: node.Lon, Tag:[]*OSMUtil.Tag{
				&OSMUtil.Tag{TagName, "type", "element"}, 
				&OSMUtil.Tag{TagName, "subtype", "ms"}, 
				&OSMUtil.Tag{TagName, "id", signal.Name[0].Value},
				&OSMUtil.Tag{TagName, "direction", "rising"}}}
			OSMUtil.InsertNode(&newNode, node.Id, OSMData)
			id_counter++
			continue 
		}

		for i, node := range OSMData.Node {
			if len(node.Tag) == 0 {
				continue
			}
			is_signal := false
			has_correct_id := false
			for _, tag := range (*node).Tag {
				if tag.K == "railway" && tag.V == "signal" {
					is_signal = true
				}	
				if tag.K == "ref" && strings.ReplaceAll(tag.V, " ", "") == signal.Name[0].Value {
					has_correct_id = true
				}	
			}

			if has_correct_id && is_signal {
				tags := &OSMData.Node[i].Tag 
				*tags = append(*tags, []*OSMUtil.Tag{
						&OSMUtil.Tag{TagName, "type", "element"}, 
						&OSMUtil.Tag{TagName, "subtype", "ms"}, 
						&OSMUtil.Tag{TagName, "id", signal.Name[0].Value},
						&OSMUtil.Tag{TagName, "direction", "rising"}}...)
				(*anchors)[kilometrage] = OSMData.Node[i]	
				found = true
				break
			} 
		}
		if found {
			continue
		}

		notFound = append(notFound, signal)
	}
	return notFound
}
