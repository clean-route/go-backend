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
	departureTime := localTime.Add(30 * time.Duration(delayCode) * time.Minute).Format("2006-01-02T15:04")

	params := url.Values{}
	params.Add("steps", "true")
	params.Add("geometries", "geojson")
	params.Add("alternatives", "true")
	params.Add("waypoints_per_route", "true")
	params.Add("access_token", config.AppConfig.MapboxAPIKey)
	params.Add("depart_at", departureTime)

	url := baseUrl + "?" + params.Encode()

	resp, err := http.Get(url)
	if err != nil {
		return mapboxroutes.RouteData{}, fmt.Errorf("error calling Mapbox API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return mapboxroutes.RouteData{}, fmt.Errorf("Mapbox API returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return mapboxroutes.RouteData{}, fmt.Errorf("error reading response body: %w", err)
	}

	var routes mapboxroutes.RouteData
	if err := json.Unmarshal(body, &routes); err != nil {
		return mapboxroutes.RouteData{}, fmt.Errorf("error unmarshaling JSON: %w", err)
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

	resp, err := http.Get(url)
	if err != nil {
		return graphhopperroutes.RouteData{}, fmt.Errorf("error calling GraphHopper API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return graphhopperroutes.RouteData{}, fmt.Errorf("error reading response body: %w", err)
	}

	var routes graphhopperroutes.RouteData
	if err := json.Unmarshal(body, &routes); err != nil {
		return graphhopperroutes.RouteData{}, fmt.Errorf("error unmarshaling JSON: %w", err)
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

	if mode == "driving-traffic" && (routePref == "fastest" || routePref == "balanced") {
		// Use Mapbox for car routes with fastest/balanced preference
		routes, err := rs.FindMapboxRoute(source, destination, delayCode)
		if err != nil {
			return nil, err
		}

		// Get energy data from GraphHopper
		energyMode := "car"
		energyRoute, err := rs.FindGraphhopperRoute(source, destination, energyMode)
		if err != nil {
			return nil, err
		}

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
			routes.Routes[i].Duration *= 1000
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
			return routes.Routes[index], nil
		} else if routePref == "balanced" {
			return rs.selectBalancedRoute(routes.Routes), nil
		}
	} else {
		// Use GraphHopper for other modes
		if mode == "driving-traffic" {
			mode = "car"
		}

		routes, err := rs.FindGraphhopperRoute(source, destination, mode)
		if err != nil {
			return nil, err
		}

		// Calculate exposure and energy
		for i := 0; i < len(routes.Paths); i++ {
			routes.Paths[i] = utils.CalculateRouteExposureGraphhopper(routes.Paths[i], delayCode)
			routes.Paths[i].TotalEnergy = utils.CalculateRouteEnergy(routes.Paths[i], mode, vehicleMass, condition, engineType)
		}

		// Return based on preference
		switch routePref {
		case "shortest":
			sort.SliceStable(routes.Paths, func(i, j int) bool {
				return routes.Paths[i].Distance < routes.Paths[j].Distance
			})
			return routes.Paths[0], nil
		case "fastest":
			sort.SliceStable(routes.Paths, func(i, j int) bool {
				return routes.Paths[i].Time < routes.Paths[j].Time
			})
			return routes.Paths[0], nil
		case "leap":
			sort.SliceStable(routes.Paths, func(i, j int) bool {
				return routes.Paths[i].TotalExposure < routes.Paths[j].TotalExposure
			})
			return routes.Paths[0], nil
		case "emission":
			sort.SliceStable(routes.Paths, func(i, j int) bool {
				return routes.Paths[i].TotalEnergy < routes.Paths[j].TotalEnergy
			})
			return routes.Paths[0], nil
		case "balanced":
			return rs.selectBalancedGraphhopperRoute(routes.Paths), nil
		}
	}

	return nil, fmt.Errorf("unsupported route preference: %s", routePref)
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

	// Calculate exposure and energy
	for i := 0; i < len(routes.Paths); i++ {
		routes.Paths[i] = utils.CalculateRouteExposureGraphhopper(routes.Paths[i], req.DelayCode)
		routes.Paths[i].TotalEnergy = utils.CalculateRouteEnergy(routes.Paths[i], req.Mode, req.VehicleMass, req.Condition, req.EngineType)
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

	return routeList, nil
}

// findAllCarRoutes finds all routes for car mode
func (rs *RouteService) findAllCarRoutes(req models.RouteRequest) (*mapboxroutes.RouteList, error) {
	mapboxRoute, err := rs.FindMapboxRoute(req.Source, req.Destination, req.DelayCode)
	if err != nil {
		return nil, err
	}

	graphhopperRoute, err := rs.FindGraphhopperRoute(req.Source, req.Destination, "car")
	if err != nil {
		return nil, err
	}

	// Calculate exposure and energy
	for i := 0; i < len(mapboxRoute.Routes); i++ {
		mapboxRoute.Routes[i] = utils.CalculateRouteExposureMapbox(mapboxRoute.Routes[i], req.DelayCode)
		graphhopperRoute.Paths[i].TotalExposure = mapboxRoute.Routes[i].TotalExposure
		mapboxRoute.Routes[i].Duration *= 1000
		mapboxRoute.Routes[i].TotalEnergy = utils.CalculateRouteEnergy(graphhopperRoute.Paths[i], req.Mode, req.VehicleMass, req.Condition, req.EngineType)
		graphhopperRoute.Paths[i].TotalEnergy = mapboxRoute.Routes[i].TotalEnergy
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
	routeList.LeapG = rs.findBestGraphhopperRoute(graphhopperRoute.Paths, "exposure")
	routeList.Lco2G = rs.findBestGraphhopperRoute(graphhopperRoute.Paths, "energy")
	routeList.Balanced = rs.selectBalancedMapboxRoute(mapboxRoute.Routes)

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
		}
	}
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
		if routes[0].Duration-routes[1].Duration < 5*60*1000 && routes[0].Distance-routes[1].Distance < 500 {
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
		if routes[0].Time-routes[1].Time < 5*60*1000 && routes[0].Distance-routes[1].Distance < 500 {
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
		if routes[0].Duration-routes[1].Duration < 5*60*1000 && routes[0].Distance-routes[1].Distance < 500 {
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
