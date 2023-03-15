package osmUtils

type SearchFile struct {
	Stations map[string]Station `json:"stations"`
	Halts    map[string]Halt    `json:"halts"`
	Signals  map[string]Signal  `json:"signals"`
}

type Station struct {
	Name string `json:"name"`
	Lat  string `json:"lat"`
	Lon  string `json:"lon"`
}

type Halt struct {
	Name string `json:"name"`
	Lat  string `json:"lat"`
	Lon  string `json:"lon"`
}

type Signal struct {
	Name string `json:"name"`
	Lat  string `json:"lat"`
	Lon  string `json:"lon"`
}
