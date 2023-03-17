## Generate OSM Setup

This README contains information on how to use the generate-osm build for production and development environments.

### Tools
- [Golang](https://golang.org/doc/install)
- [Osmium](https://osmcode.org/osmium-tool/) (Need to be in PATH variable as osmium)
- For documentation: run `go install golang.org/x/tools/cmd/godoc@latest` in the GO root directory

1. Download the `.osm.pbf` file corresponding to your area from [Geofabrik](https://download.geofabrik.de/europe/germany.html) to the folder `generate-osm/temp`.
2. If you want to map any data in the DB-XMLIss format, you need to add all files to the folder `generate-osm/temp/DBResources`.
3. Run the following command to generate the annotated OSM file:
```bash 
go build
./transform-osm
```

The final OSM file as well as the generated search indices will be in the folder `generate-osm/temp` and named `finalOsm.{xml,json}`.

<!-- Alternatively you can provide an OSM-file, in which the generated data will be inserted. In this case, the new output-file will be named `[name provided file]DB.xml` and the JSON will also be named accordingly.
-->

Both generated files (OSM and JSON search indices) now have to be copied as by the following:
1. OSM file: `/soro-s/resources/osm`
2. JSON file: `/soro-s/resources/serach_indices`

Note that it is crucial that the files have the same name (up to their file extension) in order for search indices to be discovered properly.
Also note that **you may have to rebuild your cmake project** to apply the changes.

#### CLI-Flags
```bash
--generate-lines, --gl   Generate lines all lines new (default: false)
--mapDB, --mdb   Generate lines all lines new as well as map an DB data (default: false)
--input value, -i value  The input file to read as OSM PBF file (default: "./temp/base.osm.pbf")
--output value, -o value  The output file to write result to as XML file (default: "./finalOsm.xml")
--help, -h               show help
```

### Miscellaneous

To access comprehensible documentation for all public methods, run 
```bash
godoc -http=:6060
```
in the `generate-osm` directory and access it via the browser under `localhost:6060/`. You will find it under the tab `standard library / transform-osm`.
