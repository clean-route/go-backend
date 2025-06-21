package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/clean-route/go-backend/internal/config"
	"github.com/clean-route/go-backend/internal/errors"
	"github.com/clean-route/go-backend/internal/logger"
	"github.com/clean-route/go-backend/internal/models"
	graphhopperroutes "github.com/clean-route/go-backend/internal/models/graphhopper"
	mapboxroutes "github.com/clean-route/go-backend/internal/models/mapbox"
	"github.com/clean-route/go-backend/internal/utils"
)

// RouteService handles route planning operations
type RouteService struct{}

// NewRouteService creates a new route service instance
func NewRouteService() *RouteService {
	return &RouteService{}
}

// FindMapboxRoute finds routes using Mapbox API
func (rs *RouteService) FindMapboxRoute(source [2]float64, destination [2]float64, delayCode uint8) (mapboxroutes.RouteData, error) {
	baseUrl := "https://api.mapbox.com/directions/v5/mapbox/driving-traffic/" + fmt.Sprintf("%f,%f;%f,%f", source[0], source[1], destination[0], destination[1])

	localTime := time.Now()
	departureTime := localTime.Add(60 * time.Duration(delayCode) * time.Minute).Format("2006-01-02T15:04")

	params := url.Values{}
	params.Add("steps", "true")
	params.Add("geometries", "geojson")
	params.Add("alternatives", "true")
	params.Add("waypoints_per_route", "true")
	params.Add("access_token", config.AppConfig.MapboxAPIKey)
	params.Add("depart_at", departureTime)

	url := baseUrl + "?" + params.Encode()

	logger.Debug("Calling Mapbox API",
		"url", baseUrl,
		"departure_time", departureTime,
		"delay_code", delayCode,
	)

	resp, err := http.Get(url)
	if err != nil {
		logger.Error("Failed to call Mapbox API",
			"error", err.Error(),
			"url", baseUrl,
		)
		return mapboxroutes.RouteData{}, errors.NewExternalError("error calling Mapbox API", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Mapbox API returned error status",
			"status_code", resp.StatusCode,
			"url", baseUrl,
		)
		return mapboxroutes.RouteData{}, errors.NewExternalError(fmt.Sprintf("Mapbox API returned status code: %d", resp.StatusCode), nil)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read Mapbox API response",
			"error", err.Error(),
			"url", baseUrl,
		)
		return mapboxroutes.RouteData{}, errors.NewExternalError("error reading response body", err)
	}

	var routes mapboxroutes.RouteData
	if err := json.Unmarshal(body, &routes); err != nil {
		logger.Error("Failed to unmarshal Mapbox API response",
			"error", err.Error(),
			"url", baseUrl,
			"response_body", string(body),
		)
		return mapboxroutes.RouteData{}, errors.NewExternalError("error unmarshaling JSON", err)
	}

	logger.Debug("Successfully retrieved Mapbox routes",
		"routes_count", len(routes.Routes),
		"response_body_length", len(body),
	)

	// Log if no routes were found
	if len(routes.Routes) == 0 {
		logger.Warn("Mapbox API returned no routes",
			"source", source,
			"destination", destination,
			"departure_time", departureTime,
			"response_body", string(body),
		)
	}

	return routes, nil
}

// FindGraphhopperRoute finds routes using GraphHopper API
func (rs *RouteService) FindGraphhopperRoute(source [2]float64, destination [2]float64, mode string) (graphhopperroutes.RouteData, error) {
	baseUrl := "https://graphhopper.com/api/1/route?"

	params := url.Values{}
	params.Add("point", fmt.Sprintf("%f,%f", source[1], source[0]))
	params.Add("point", fmt.Sprintf("%f,%f", destination[1], destination[0]))
	params.Add("vehicle", mode)
	params.Add("debug", "true")
	params.Add("key", config.AppConfig.GraphhopperAPIKey)
	params.Add("type", "json")
	params.Add("points_encoded", "false")
	params.Add("algorithm", "alternative_route")
	params.Add("alternative_route.max_paths", "4")
	params.Add("alternative_route.max_weight_factor", "1.4")
	params.Add("alternative_route.max_share_factor", "0.6")
	params.Add("elevation", "true")

	url := baseUrl + params.Encode()

	logger.Debug("Calling GraphHopper API",
		"url", baseUrl,
		"mode", mode,
	)

	resp, err := http.Get(url)
	if err != nil {
		logger.Error("Failed to call GraphHopper API",
			"error", err.Error(),
			"url", baseUrl,
			"mode", mode,
		)
		return graphhopperroutes.RouteData{}, errors.NewExternalError("error calling GraphHopper API", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read GraphHopper API response",
			"error", err.Error(),
			"url", baseUrl,
			"mode", mode,
		)
		return graphhopperroutes.RouteData{}, errors.NewExternalError("error reading response body", err)
	}

	var routes graphhopperroutes.RouteData
	if err := json.Unmarshal(body, &routes); err != nil {
		logger.Error("Failed to unmarshal GraphHopper API response",
			"error", err.Error(),
			"url", baseUrl,
			"mode", mode,
			"response_body", string(body),
		)
		return graphhopperroutes.RouteData{}, errors.NewExternalError("error unmarshaling JSON", err)
	}

	logger.Debug("Successfully retrieved GraphHopper routes",
		"mode", mode,
		"paths_count", len(routes.Paths),
		"response_body_length", len(body),
	)

	// Log if no paths were found
	if len(routes.Paths) == 0 {
		logger.Warn("GraphHopper API returned no paths",
			"mode", mode,
			"source", source,
			"destination", destination,
			"response_body", string(body),
		)
	}

	return routes, nil
}

// FindSingleRoute finds a single route based on preferences
func (rs *RouteService) FindSingleRoute(req models.RouteRequest) (interface{}, error) {
	source := req.Source
	destination := req.Destination
	delayCode := req.DelayCode
	mode := req.Mode
	routePref := req.RoutePreference
	vehicleMass := req.VehicleMass
	condition := req.Condition
	engineType := req.EngineType

	logger.Debug("Finding single route",
		"mode", mode,
		"route_preference", routePref,
		"delay_code", delayCode,
		"vehicle_mass", vehicleMass,
		"condition", condition,
		"engine_type", engineType,
		"source", source,
		"destination", destination,
	)

	if mode == "driving-traffic" && (routePref == "fastest" || routePref == "balanced") {
		// Use Mapbox for car routes with fastest/balanced preference
		logger.Debug("Using Mapbox API for car route",
			"route_preference", routePref,
		)

		routes, err := rs.FindMapboxRoute(source, destination, delayCode)
		if err != nil {
			logger.Error("Failed to find Mapbox route",
				"error", err.Error(),
				"source", source,
				"destination", destination,
				"delay_code", delayCode,
			)
			return nil, errors.Wrap(err, "failed to find Mapbox route")
		}

		// Check if routes are available
		if len(routes.Routes) == 0 {
			logger.Error("No Mapbox routes found",
				"source", source,
				"destination", destination,
				"mode", mode,
			)
			return nil, errors.NewNotFoundError("No routes found for the given coordinates", nil)
		}

		// Get energy data from GraphHopper
		energyMode := "car"
		logger.Debug("Getting energy data from GraphHopper",
			"energy_mode", energyMode,
		)

		energyRoute, err := rs.FindGraphhopperRoute(source, destination, energyMode)
		if err != nil {
			logger.Error("Failed to find GraphHopper energy route",
				"error", err.Error(),
				"source", source,
				"destination", destination,
				"energy_mode", energyMode,
			)
			return nil, errors.Wrap(err, "failed to find GraphHopper energy route")
		}

		// Check if energy routes are available
		if len(energyRoute.Paths) == 0 {
			logger.Error("No GraphHopper energy routes found",
				"source", source,
				"destination", destination,
				"mode", energyMode,
			)
			return nil, errors.NewNotFoundError("No energy data available for the route", nil)
		}

		logger.Debug("Calculating route exposure and energy",
			"mapbox_routes_count", len(routes.Routes),
			"energy_routes_count", len(energyRoute.Paths),
		)

		// Align routes by distance
		sort.SliceStable(energyRoute.Paths, func(i, j int) bool {
			return energyRoute.Paths[i].Distance < energyRoute.Paths[j].Distance
		})

		sort.SliceStable(routes.Routes, func(i, j int) bool {
			return routes.Routes[i].Distance < routes.Routes[j].Distance
		})

		// Calculate exposure and energy
		for i := 0; i < len(routes.Routes) && i < len(energyRoute.Paths); i++ {
			routes.Routes[i] = utils.CalculateRouteExposureMapbox(routes.Routes[i], delayCode)
			// Keep Mapbox duration in seconds (no conversion needed)
			routes.Routes[i].TotalEnergy = utils.CalculateRouteEnergy(energyRoute.Paths[i], mode, vehicleMass, condition, engineType)
		}

		// Return based on preference
		if routePref == "fastest" {
			index := 0
			for i := 0; i < len(routes.Routes); i++ {
				if routes.Routes[i].Duration < routes.Routes[index].Duration {
					index = i
				}
			}
			logger.Debug("Selected fastest route",
				"route_index", index,
				"duration", routes.Routes[index].Duration,
			)
			return routes.Routes[index], nil
		} else if routePref == "balanced" {
			logger.Debug("Selecting balanced route")
			return rs.selectBalancedRoute(routes.Routes), nil
		}
	} else {
		// Use GraphHopper for other modes
		if mode == "driving-traffic" {
			mode = "car"
		}

		logger.Debug("Using GraphHopper API for route",
			"mode", mode,
			"route_preference", routePref,
		)

		routes, err := rs.FindGraphhopperRoute(source, destination, mode)
		if err != nil {
			logger.Error("Failed to find GraphHopper route",
				"error", err.Error(),
				"source", source,
				"destination", destination,
				"mode", mode,
			)
			return nil, errors.Wrap(err, "failed to find GraphHopper route")
		}

		// Check if routes are available
		if len(routes.Paths) == 0 {
			logger.Error("No GraphHopper routes found",
				"source", source,
				"destination", destination,
				"mode", mode,
			)
			return nil, errors.NewNotFoundError("No routes found for the given coordinates", nil)
		}

		logger.Debug("Calculating route exposure and energy",
			"routes_count", len(routes.Paths),
		)

		// Calculate exposure and energy
		for i := 0; i < len(routes.Paths); i++ {
			routes.Paths[i] = utils.CalculateRouteExposureGraphhopper(routes.Paths[i], delayCode)
			routes.Paths[i].TotalEnergy = utils.CalculateRouteEnergy(routes.Paths[i], mode, vehicleMass, condition, engineType)
			// Convert GraphHopper time from milliseconds to seconds
			routes.Paths[i].Time = routes.Paths[i].Time / 1000
		}

		// Return based on preference
		switch routePref {
		case "shortest":
			sort.SliceStable(routes.Paths, func(i, j int) bool {
				return routes.Paths[i].Distance < routes.Paths[j].Distance
			})
			logger.Debug("Selected shortest route",
				"distance", routes.Paths[0].Distance,
			)
			return routes.Paths[0], nil
		case "fastest":
			sort.SliceStable(routes.Paths, func(i, j int) bool {
				return routes.Paths[i].Time < routes.Paths[j].Time
			})
			logger.Debug("Selected fastest route",
				"time", routes.Paths[0].Time,
			)
			return routes.Paths[0], nil
		case "leap":
			sort.SliceStable(routes.Paths, func(i, j int) bool {
				return routes.Paths[i].TotalExposure < routes.Paths[j].TotalExposure
			})
			logger.Debug("Selected lowest exposure route",
				"exposure", routes.Paths[0].TotalExposure,
			)
			return routes.Paths[0], nil
		case "emission":
			sort.SliceStable(routes.Paths, func(i, j int) bool {
				return routes.Paths[i].TotalEnergy < routes.Paths[j].TotalEnergy
			})
			logger.Debug("Selected lowest emission route",
				"energy", routes.Paths[0].TotalEnergy,
			)
			return routes.Paths[0], nil
		case "balanced":
			logger.Debug("Selecting balanced route")
			return rs.selectBalancedGraphhopperRoute(routes.Paths), nil
		}
	}

	logger.Error("Unsupported route preference",
		"route_preference", routePref,
		"mode", mode,
	)
	return nil, errors.NewValidationError(fmt.Sprintf("unsupported route preference: %s", routePref), nil)
}

// FindAllRoutes finds all route types for a given request
func (rs *RouteService) FindAllRoutes(req models.RouteRequest) (interface{}, error) {
	if req.Mode == "scooter" {
		return rs.findAllScooterRoutes(req)
	} else if req.Mode == "driving-traffic" {
		return rs.findAllCarRoutes(req)
	}

	return nil, fmt.Errorf("unsupported mode: %s", req.Mode)
}

// findAllScooterRoutes finds all routes for scooter mode
func (rs *RouteService) findAllScooterRoutes(req models.RouteRequest) (*graphhopperroutes.RouteList, error) {
	routes, err := rs.FindGraphhopperRoute(req.Source, req.Destination, req.Mode)
	if err != nil {
		return nil, err
	}

	// Check if routes are available
	if len(routes.Paths) == 0 {
		logger.Error("No GraphHopper routes found for scooter",
			"source", req.Source,
			"destination", req.Destination,
			"mode", req.Mode,
		)
		return nil, errors.NewNotFoundError("No routes found for the given coordinates", nil)
	}

	// Calculate exposure and energy
	for i := 0; i < len(routes.Paths); i++ {
		routes.Paths[i] = utils.CalculateRouteExposureGraphhopper(routes.Paths[i], req.DelayCode)
		routes.Paths[i].TotalEnergy = utils.CalculateRouteEnergy(routes.Paths[i], req.Mode, req.VehicleMass, req.Condition, req.EngineType)
		// Convert GraphHopper time from milliseconds to seconds
		routes.Paths[i].Time = routes.Paths[i].Time / 1000
	}

	routeList := &graphhopperroutes.RouteList{
		Source:      req.Source[:],
		Destination: req.Destination[:],
		DelayCode:   req.DelayCode,
		Mode:        req.Mode,
		RoutePref:   req.RoutePreference,
	}

	// Find best routes for each preference
	routeList.Fastest = rs.findBestRoute(routes.Paths, "time")
	routeList.Shortest = rs.findBestRoute(routes.Paths, "distance")
	routeList.LeapG = rs.findBestRoute(routes.Paths, "exposure")
	routeList.Lco2G = rs.findBestRoute(routes.Paths, "energy")
	routeList.Balanced = rs.selectBalancedGraphhopperRoute(routes.Paths)

	// Debug logging for route selection
	logger.Debug("Route selection results for scooter routes",
		"shortest_distance", routeList.Shortest.Distance,
		"shortest_exposure", routeList.Shortest.TotalExposure,
		"leap_distance", routeList.LeapG.Distance,
		"leap_exposure", routeList.LeapG.TotalExposure,
		"fastest_time", routeList.Fastest.Time,
		"fastest_exposure", routeList.Fastest.TotalExposure,
	)

	return routeList, nil
}

// findAllCarRoutes finds all routes for car mode
func (rs *RouteService) findAllCarRoutes(req models.RouteRequest) (*mapboxroutes.RouteList, error) {
	logger.Debug("Starting findAllCarRoutes",
		"source", req.Source,
		"destination", req.Destination,
		"mode", req.Mode,
		"delay_code", req.DelayCode,
	)

	mapboxRoute, err := rs.FindMapboxRoute(req.Source, req.Destination, req.DelayCode)
	if err != nil {
		logger.Error("Failed to find Mapbox routes",
			"error", err.Error(),
			"source", req.Source,
			"destination", req.Destination,
		)
		return nil, err
	}

	// Check if Mapbox routes are available
	if len(mapboxRoute.Routes) == 0 {
		logger.Error("No Mapbox routes found for car",
			"source", req.Source,
			"destination", req.Destination,
			"mode", req.Mode,
		)
		return nil, errors.NewNotFoundError("No routes found for the given coordinates", nil)
	}

	logger.Debug("Successfully retrieved Mapbox routes",
		"routes_count", len(mapboxRoute.Routes),
	)

	graphhopperRoute, err := rs.FindGraphhopperRoute(req.Source, req.Destination, "car")
	if err != nil {
		logger.Error("Failed to find GraphHopper routes",
			"error", err.Error(),
			"source", req.Source,
			"destination", req.Destination,
		)
		return nil, err
	}

	// Check if GraphHopper routes are available
	if len(graphhopperRoute.Paths) == 0 {
		logger.Error("No GraphHopper routes found for car energy calculation",
			"source", req.Source,
			"destination", req.Destination,
			"mode", "car",
		)
		return nil, errors.NewNotFoundError("No energy data available for the route", nil)
	}

	logger.Debug("Successfully retrieved GraphHopper routes",
		"paths_count", len(graphhopperRoute.Paths),
	)

	// Calculate exposure and energy
	for i := 0; i < len(mapboxRoute.Routes) && i < len(graphhopperRoute.Paths); i++ {
		mapboxRoute.Routes[i] = utils.CalculateRouteExposureMapbox(mapboxRoute.Routes[i], req.DelayCode)
		// Calculate energy for Mapbox route using corresponding GraphHopper path
		mapboxRoute.Routes[i].TotalEnergy = utils.CalculateRouteEnergy(graphhopperRoute.Paths[i], req.Mode, req.VehicleMass, req.Condition, req.EngineType)
		// Convert GraphHopper time from milliseconds to seconds
		graphhopperRoute.Paths[i].Time = graphhopperRoute.Paths[i].Time / 1000
		// Calculate exposure and energy for GraphHopper route
		graphhopperRoute.Paths[i] = utils.CalculateRouteExposureGraphhopper(graphhopperRoute.Paths[i], req.DelayCode)
		graphhopperRoute.Paths[i].TotalEnergy = utils.CalculateRouteEnergy(graphhopperRoute.Paths[i], req.Mode, req.VehicleMass, req.Condition, req.EngineType)

		// Debug logging for each route
		logger.Debug("Route calculation results",
			"route_index", i,
			"mapbox_distance", mapboxRoute.Routes[i].Distance,
			"mapbox_exposure", mapboxRoute.Routes[i].TotalExposure,
			"mapbox_energy", mapboxRoute.Routes[i].TotalEnergy,
			"graphhopper_distance", graphhopperRoute.Paths[i].Distance,
			"graphhopper_exposure", graphhopperRoute.Paths[i].TotalExposure,
			"graphhopper_energy", graphhopperRoute.Paths[i].TotalEnergy,
		)
	}

	routeList := &mapboxroutes.RouteList{
		Source:      req.Source[:],
		Destination: req.Destination[:],
		DelayCode:   req.DelayCode,
		Mode:        req.Mode,
		RoutePref:   req.RoutePreference,
	}

	// Find best routes for each preference
	routeList.Fastest = rs.findBestMapboxRoute(mapboxRoute.Routes, "duration")
	routeList.Shortest = rs.findBestMapboxRoute(mapboxRoute.Routes, "distance")
	routeList.Leap = rs.findBestMapboxRoute(mapboxRoute.Routes, "exposure")
	routeList.Lco2 = rs.findBestMapboxRoute(mapboxRoute.Routes, "energy")
	routeList.Balanced = rs.selectBalancedMapboxRoute(mapboxRoute.Routes)
	// Keep GraphHopper routes for energy calculation only
	routeList.LeapG = rs.findBestGraphhopperRoute(graphhopperRoute.Paths, "exposure")
	routeList.Lco2G = rs.findBestGraphhopperRoute(graphhopperRoute.Paths, "energy")

	// Validate that we have non-zero values for exposure and energy
	// If all routes have zero exposure, use the shortest route for LEAP
	if routeList.Leap.TotalExposure == 0 {
		logger.Warn("All routes have zero exposure, using shortest route for LEAP")
		routeList.Leap = routeList.Shortest
	}

	// If all routes have zero energy, use the shortest route for LCO2
	if routeList.Lco2.TotalEnergy == 0 {
		logger.Warn("All routes have zero energy, using shortest route for LCO2")
		routeList.Lco2 = routeList.Shortest
	}

	// Debug logging for route selection
	logger.Debug("Route selection results for car routes",
		"shortest_distance", routeList.Shortest.Distance,
		"shortest_exposure", routeList.Shortest.TotalExposure,
		"leap_distance", routeList.Leap.Distance,
		"leap_exposure", routeList.Leap.TotalExposure,
		"fastest_duration", routeList.Fastest.Duration,
		"fastest_exposure", routeList.Fastest.TotalExposure,
	)

	return routeList, nil
}

// Helper functions for finding best routes
func (rs *RouteService) findBestRoute(routes []graphhopperroutes.Path, criteria string) graphhopperroutes.Path {
	if len(routes) == 0 {
		return graphhopperroutes.Path{}
	}

	index := 0
	for i := 1; i < len(routes); i++ {
		switch criteria {
		case "time":
			if routes[i].Time < routes[index].Time {
				index = i
			}
		case "distance":
			if routes[i].Distance < routes[index].Distance {
				index = i
			}
		case "exposure":
			if routes[i].TotalExposure < routes[index].TotalExposure {
				index = i
			}
		case "energy":
			if routes[i].TotalEnergy < routes[index].TotalEnergy {
				index = i
			}
		}
	}
	return routes[index]
}

func (rs *RouteService) findBestMapboxRoute(routes []mapboxroutes.Route, criteria string) mapboxroutes.Route {
	if len(routes) == 0 {
		logger.Debug("No Mapbox routes available for selection")
		return mapboxroutes.Route{}
	}

	index := 0
	for i := 1; i < len(routes); i++ {
		switch criteria {
		case "duration":
			if routes[i].Duration < routes[index].Duration {
				index = i
			}
		case "distance":
			if routes[i].Distance < routes[index].Distance {
				index = i
			}
		case "exposure":
			logger.Debug("Comparing exposure values",
				"current_index", index,
				"current_exposure", routes[index].TotalExposure,
				"comparing_index", i,
				"comparing_exposure", routes[i].TotalExposure,
			)
			// Only select route with zero exposure if all routes have zero exposure
			if routes[i].TotalExposure > 0 && routes[index].TotalExposure == 0 {
				index = i
			} else if routes[i].TotalExposure < routes[index].TotalExposure && routes[index].TotalExposure > 0 {
				index = i
			}
		case "energy":
			logger.Debug("Comparing energy values",
				"current_index", index,
				"current_energy", routes[index].TotalEnergy,
				"comparing_index", i,
				"comparing_energy", routes[i].TotalEnergy,
			)
			// Only select route with zero energy if all routes have zero energy
			if routes[i].TotalEnergy > 0 && routes[index].TotalEnergy == 0 {
				index = i
			} else if routes[i].TotalEnergy < routes[index].TotalEnergy && routes[index].TotalEnergy > 0 {
				index = i
			}
		}
	}

	// Debug logging for route selection
	logger.Debug("Selected best Mapbox route",
		"criteria", criteria,
		"selected_index", index,
		"selected_distance", routes[index].Distance,
		"selected_exposure", routes[index].TotalExposure,
		"selected_energy", routes[index].TotalEnergy,
		"total_routes", len(routes),
	)

	return routes[index]
}

func (rs *RouteService) findBestGraphhopperRoute(routes []graphhopperroutes.Path, criteria string) graphhopperroutes.Path {
	if len(routes) == 0 {
		return graphhopperroutes.Path{}
	}

	index := 0
	for i := 1; i < len(routes); i++ {
		switch criteria {
		case "exposure":
			if routes[i].TotalExposure < routes[index].TotalExposure {
				index = i
			}
		case "energy":
			if routes[i].TotalEnergy < routes[index].TotalEnergy {
				index = i
			}
		}
	}
	return routes[index]
}

// selectBalancedRoute selects the best balanced route
func (rs *RouteService) selectBalancedRoute(routes []mapboxroutes.Route) mapboxroutes.Route {
	if len(routes) == 0 {
		return mapboxroutes.Route{}
	}
	if len(routes) == 1 {
		return routes[0]
	}
	if len(routes) == 2 {
		if routes[0].Duration-routes[1].Duration < 300 && routes[0].Distance-routes[1].Distance < 500 {
			if routes[0].TotalExposure < routes[1].TotalExposure {
				return routes[0]
			}
			return routes[1]
		}
		if routes[0].Duration < routes[1].Duration {
			return routes[0]
		}
		return routes[1]
	}
	return routes[0] // Default to first route for more than 2 routes
}

func (rs *RouteService) selectBalancedGraphhopperRoute(routes []graphhopperroutes.Path) graphhopperroutes.Path {
	if len(routes) == 0 {
		return graphhopperroutes.Path{}
	}
	if len(routes) == 1 {
		return routes[0]
	}
	if len(routes) == 2 {
		if routes[0].Time-routes[1].Time < 300 && routes[0].Distance-routes[1].Distance < 500 {
			if routes[0].TotalExposure < routes[1].TotalExposure {
				return routes[0]
			}
			return routes[1]
		}
		if routes[0].Time < routes[1].Time {
			return routes[0]
		}
		return routes[1]
	}

	// For more than 2 routes, sort by exposure, then by time for top 3, then by energy for top 2
	sort.Slice(routes, func(i, j int) bool {
		return routes[i].TotalExposure < routes[j].TotalExposure
	})

	if len(routes) > 3 {
		sort.SliceStable(routes[:3], func(i, j int) bool {
			return routes[i].Time < routes[j].Time
		})
	}

	if len(routes) > 2 {
		sort.Slice(routes[:2], func(i, j int) bool {
			return routes[i].TotalEnergy < routes[j].TotalEnergy
		})
	}

	return routes[0]
}

// selectBalancedMapboxRoute selects the best balanced route for Mapbox routes
func (rs *RouteService) selectBalancedMapboxRoute(routes []mapboxroutes.Route) mapboxroutes.Route {
	if len(routes) == 0 {
		return mapboxroutes.Route{}
	}
	if len(routes) == 1 {
		return routes[0]
	}
	if len(routes) == 2 {
		if routes[0].Duration-routes[1].Duration < 300 && routes[0].Distance-routes[1].Distance < 500 {
			if routes[0].TotalExposure < routes[1].TotalExposure {
				return routes[0]
			}
			return routes[1]
		}
		if routes[0].Duration < routes[1].Duration {
			return routes[0]
		}
		return routes[1]
	}
	return routes[0] // Default to first route for more than 2 routes
}
