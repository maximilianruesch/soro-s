#include <filesystem>

#include "utl/cmd_line_parser.h"
#include "utl/to_vec.h"
#include "utl/logging.h"

#include "soro/infrastructure/infrastructure.h"

#include "soro/server/import/import.h"
#include "soro/server/osm_export/osm_export.h"
#include "soro/server/soro_server.h"
#include "soro/server/search_util.h"

namespace fs = std::filesystem;

struct server_settings {
  utl::cmd_line_flag<fs::path, UTL_LONG("--resource_dir"),
                     UTL_DESC("where the server reads the resources from")>
      resource_dir_{"server_resources/resources/"};

  utl::cmd_line_flag<fs::path, UTL_LONG("--server_resource_dir"),
                     UTL_DESC("where the server puts the generated resources")>
      server_resource_dir_{"server_resources/"};

  utl::cmd_line_flag<std::string, UTL_LONG("--address"),
                     UTL_DESC("ip address the server listens on")>
      address_{"0.0.0.0"};

  utl::cmd_line_flag<soro::server::port_t, UTL_LONG("--port"),
                     UTL_DESC("port the server listens on")>
      port_{8080};

  utl::cmd_line_flag<bool, UTL_LONG("--regenerate"), UTL_SHORT("-r"),
                     UTL_DESC("regenerate server resources")>
      regenerate_{false};

  utl::cmd_line_flag<bool, UTL_LONG("--test"), UTL_SHORT("-t"),
                     UTL_DESC("start in test mode - quit after 1s")>
      test_{false};
};

bool is_infrastructure_file(fs::path const& possible_infrastructure) {
  return !fs::is_directory(possible_infrastructure) &&
         possible_infrastructure.has_extension() &&
         possible_infrastructure.extension() == ".iss";
}

void exists_or_create_dir(fs::path const& dir_path) {
  if (!fs::exists(dir_path)) {
    fs::create_directory(dir_path);
  }
}

int failed_startup() { return 1; }


// Filtering osm station and stops

// Filter for station and stop nodes
std::vector<pugi::xml_node> filter_halt(const pugi::xml_document& xml_doc) {
  std::vector<pugi::xml_node> filtered;
  auto osm_node = xml_doc.child("osm");

  auto children = osm_node.children();
  for (auto child:children) {
    if (std::string("node") == child.name()) {
      auto tags = child.children();

      for (auto tag:tags) {

        if (std::string("railway") == tag.attribute("k").as_string() && 
            (std::string("station") == tag.attribute("v").as_string() || std::string("halt") == tag.attribute("v").as_string())) {
          filtered.push_back(child);
        }

        if (std::string("subtype") == tag.attribute("k").as_string() && std::string("ms") == tag.attribute("v").as_string()) {
          filtered.push_back(child);
        }
      }
    }
  }

  return filtered;
}

// Extract the information from the station and stop nodes
std::vector<soro::server::osm_object> extract_halt_info(
    const std::vector<pugi::xml_node>& nodes) {
  std::vector<soro::server::osm_object> result;

  for (const auto& node : nodes) {
    const double lon = node.attribute("lon").as_double();
    const double lat = node.attribute("lat").as_double();
    std::string name = "default";
    soro::server::osm_type type = soro::server::osm_type::UNDEFINED;

    auto tags = node.children();

    // Determine Type
    for (auto tag : tags) {
      if (std::string("railway") == tag.attribute("k").as_string() &&
          std::string("station") == tag.attribute("v").as_string())
          type = soro::server::osm_type::STATION;
      else if (std::string("railway") == tag.attribute("k").as_string() &&
               std::string("halt") == tag.attribute("v").as_string())
          type = soro::server::osm_type::HALT;
      else if (std::string("subtype") == tag.attribute("k").as_string() &&
               std::string("ms") == tag.attribute("v").as_string()) {
          type = soro::server::osm_type::MAIN_SIGNAL;
      }
    }

    for (auto tag:tags) {
      switch (type) {
          case soro::server::osm_type::HALT: 
          case soro::server::osm_type::STATION: 
              if (std::string("name") == tag.attribute("k").as_string())
                  name = tag.attribute("v").as_string();
              break;
          case soro::server::osm_type::MAIN_SIGNAL: 
              if (std::string("id") == tag.attribute("k").as_string())
                  name = tag.attribute("v").as_string();
              break;
          default: break;
      }
    }

    result.emplace_back(name, type, lon, lat);
  }

  return result;
}

int main(int argc, char const** argv) {
  server_settings s;
  std::cout << "\n\t\t[SORO Server]\n\n";

  try {
    s = utl::parse<server_settings>(argc, argv);
  } catch (std::exception const& e) {
    std::cout << "options error: " << e.what() << "\n";
    return failed_startup();
  }

  auto const coord_file = s.server_resource_dir_ / "misc" / "btrs_geo.csv";

  fs::path const tt_dir = s.server_resource_dir_ / "timetable";
  fs::path const infra_dir = s.server_resource_dir_ / "infrastructure";

  exists_or_create_dir(s.server_resource_dir_);
  exists_or_create_dir(tt_dir);
  exists_or_create_dir(infra_dir);

  std::vector<fs::path> infra_todo;
  for (auto const& dir_entry :
       fs::directory_iterator{s.resource_dir_ / "infrastructure"}) {

    if (!dir_entry.is_directory()) {
      continue;
    }

    auto const res_path = infra_dir / dir_entry.path().filename();

    if (!fs::exists(res_path) ||
        last_write_time(res_path) < dir_entry.last_write_time()) {
      infra_todo.emplace_back(dir_entry.path());
    }
  }

  if (s.regenerate_) {
    infra_todo.clear();
    for (auto&& dir_entry :
         fs::directory_iterator{s.resource_dir_ / "infrastructure"}) {

      infra_todo.emplace_back(dir_entry.path());
    }
  }


  // Create paths for infraFiles files
  std::vector<fs::path> all_osm_paths;

  //OSM data is generated from the XML files in Infrastructure and stored in /Serverresources
  for (auto const& infra_file : infra_todo) {
    auto const infra_res_dir = infra_dir / infra_file.filename();
    exists_or_create_dir(infra_res_dir);

    soro::infra::infrastructure_options opts;
    opts.infrastructure_path_ = infra_file;
    opts.gps_coord_path_ = coord_file;
    opts.determine_layout_ = true;
    opts.determine_interlocking_ = false;
    opts.determine_conflicts_ = false;

      soro::infra::infrastructure const infra(opts);

      auto const osm_file =
          infra_res_dir / infra_file.filename().replace_extension(".osm");
      //This generates a new OSMFile from the infrasData
      soro::server::osm_export::export_and_write(*infra, osm_file);

      all_osm_paths.push_back(osm_file);

  }

  // Create paths for osm files
  std::vector<fs::path> osm_paths;

  //All real osm files are collected from folder /resources/osm
  auto osm_path = s.resource_dir_ / "osm";
  if (fs::exists(osm_path)) { // if folder "osm" folder exists, generate paths to osm files
      for (auto&& dir_entry : fs::directory_iterator{ osm_path }) {
          osm_paths.emplace_back(dir_entry);
      }
  }

  // Copy every osm file to server
  std::unordered_map<std::string, std::vector<soro::server::osm_object>> halts;
  for (const auto& osm_file : osm_paths) {
      auto const infra_res_dir = infra_dir / osm_file.filename().replace_extension("");
      exists_or_create_dir(infra_res_dir);

      auto const osm_server_file = infra_res_dir / osm_file.filename();

      // load, filter and save to a new location
      pugi::xml_document osm_data;
      auto const load_result = osm_data.load_file(osm_file.c_str());

      const auto filtered = filter_halt(osm_data);
      const auto fileName = osm_file.filename().replace_extension("").string();
      halts[fileName] = extract_halt_info(filtered);

      if (!load_result) {
          uLOG(utl::err)
              << "Failed to read real OSM Data. Will resume without it. Error: "
              << load_result.description()
              << "\n";
          continue;
      }

      osm_data.save_file(osm_server_file.c_str());
      all_osm_paths.push_back(osm_server_file);
  }


  //Generate Tiles and its filestructure for all osm files
  for(const auto& osm_file : all_osm_paths){
      auto const infra_res_dir = infra_dir / osm_file.filename().replace_extension("");
      auto const osm_server_file = infra_res_dir / osm_file.filename();

      auto const tiles_dir = infra_res_dir / "tiles";
      exists_or_create_dir(tiles_dir);

      auto const tmp_dir = infra_res_dir / "tmp";
      exists_or_create_dir(tmp_dir);

      soro::server::import_settings const import_settings(
        osm_server_file,
        infra_res_dir / "tiles" /
            osm_file.filename().replace_extension(".mdb"),
        tmp_dir, s.server_resource_dir_ / "profile" / "profile.lua");

    //tile server created tiles for given OSM file
    soro::server::import_tiles(import_settings);
  }

  soro::server::server const server(s.address_.val(), s.port_.val(),
                                    s.server_resource_dir_.val(), s.test_, halts);
}