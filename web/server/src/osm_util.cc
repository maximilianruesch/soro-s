#include "soro/server/osm_util.h"

namespace soro::server {


	std::string map_type(const osm_type type) {
		switch (type) {
            case osm_type::HALT: 
				return "hlt"; 
				break;
            case osm_type::STATION: 
				return "station"; 
				break;
            case osm_type::MAIN_SIGNAL: 
				return "ms"; 
				break;
			default: 
				return "undefined";
				break;
		}
	}


}  // namespace soro::server