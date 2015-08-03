package router

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type RouteData struct {
	view   http.HandlerFunc
	routes map[string]*RouteData
}

type Router struct {
	method map[string]map[string]*RouteData
}

func InitRouter() *Router {
	router := Router{method: make(map[string]map[string]*RouteData)}
	router.method["GET"] = make(map[string]*RouteData)
	router.method["GET"]["/"] = &RouteData{view: http.NotFound, routes: make(map[string]*RouteData)}

	router.method["POST"] = make(map[string]*RouteData)
	router.method["POST"]["/"] = &RouteData{view: http.NotFound, routes: make(map[string]*RouteData)}

	router.method["PUT"] = make(map[string]*RouteData)
	router.method["PUT"]["/"] = &RouteData{view: http.NotFound, routes: make(map[string]*RouteData)}

	router.method["DELETE"] = make(map[string]*RouteData)
	router.method["DELETE"]["/"] = &RouteData{view: http.NotFound, routes: make(map[string]*RouteData)}

	router.method["HEAD"] = make(map[string]*RouteData)
	router.method["HEAD"]["/"] = &RouteData{view: http.NotFound, routes: make(map[string]*RouteData)}
	return &router
}

func (r *Router) AddRoute(method string, url string, view http.HandlerFunc) {
	// Force a "/" at the beginning of every pattern
	url = CleanUrl(url)

	routes := r.method[strings.ToUpper(method)]
	// Root
	if len(url) == 1 {
		if route, ok := routes["/"]; ok {
			route.view = view
		} else {
			routes["/"] = &RouteData{view: view, routes: make(map[string]*RouteData)}
		}
	} else { // Child
		currentRoute := routes["/"]

		// /users/ -> ["","users", ""]
		// /users/:uid/ -> ["", "users", ":uid", ""]
		urlParts := strings.Split(url, "/")
		urlParts = urlParts[1 : len(urlParts)-1]
		urlPartsLength := len(urlParts)
		for i, urlPart := range urlParts {
			// check for route
			if nextRoute, ok := currentRoute.routes[urlPart]; ok {
				currentRoute = nextRoute
			} else {
				// route data not found, create
				routeData := &RouteData{view: http.NotFound, routes: make(map[string]*RouteData)}
				currentRoute.routes[urlPart] = routeData
			}
			if i == urlPartsLength-1 {
				currentRoute.routes[urlPart].view = view
			}
		}
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	routes := r.method[req.Method]
	urlPath := CleanUrl(req.URL.Path)
	currentView := http.NotFound

	if len(urlPath) == 1 {
		currentView = routes["/"].view
	} else {
		currentRoute := routes["/"]
		urlParts := strings.Split(urlPath, "/")
		urlParts = urlParts[1 : len(urlParts)-1]
		urlPartsLength := len(urlParts)
		for i, urlPart := range urlParts {
			routeFound := false
			if nextRoute, ok := currentRoute.routes[urlPart]; ok {
				routeFound = true
				currentRoute = nextRoute
			} else {
				// check for named params
				for routeUrl, routeData := range currentRoute.routes {
					if routeUrl[0:1] != ":" {
						routeFound = false
						continue
					} else {
						routeFound = true
						urlParam := routeUrl[1:]
						query := req.URL.Query()
						query.Set(urlParam, urlPart)
						req.URL.RawQuery = url.Values(query).Encode() + "&" + req.URL.RawQuery
						currentRoute = routeData
						break
					}
				}
			}
			if i == urlPartsLength-1 && routeFound {
				currentView = currentRoute.view
			}
		}
	}

	currentView(w, req)
}

func CleanUrl(url string) string {
	// Force a "/" at the beginning of every pattern
	if url[0:1] != "/" {
		url = "/" + url
	}
	// Force a "/" at the end of every pattern
	if url[len(url)-1:] != "/" {
		url += "/"
	}
	return url
}
