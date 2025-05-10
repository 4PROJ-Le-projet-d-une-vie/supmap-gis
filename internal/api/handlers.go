package api

import (
	"errors"
	"fmt"
	"github.com/matheodrd/httphelper/handler"
	"net/http"
	"strconv"
	"supmap-gis/internal/domain/services"
	"supmap-gis/internal/providers/valhalla"
)

// ErrResponse is here for the sole purpose of being used in swaggo annotations.
type ErrResponse struct {
	Message string `json:"message"`
}

// @Summary Géocode une adresse
// @Description Convertit une adresse en coordonnées. Plusieurs résultats peuvent être renvoyés.
// @Tags geocoding
// @Produce json
// @Param address query string true "Adresse dont on souhaite avoir les coordonnées GPS. Exemple: 'Abbaye aux Dames Caen'"
// @Success 200 {object} handler.Response[[]services.Place]
// @Failure 500 {object} ErrResponse "Erreur interne du serveur"
// @Router /geocode [get]
func (s *Server) geocodeHandler() http.HandlerFunc {
	return handler.Handler(func(w http.ResponseWriter, r *http.Request) error {
		address := r.URL.Query().Get("address")
		if address == "" {
			return handler.NewErrWithStatus(http.StatusBadRequest, fmt.Errorf("missing 'address' query parameter"))
		}

		result, err := s.geocodingService.Search(r.Context(), address)
		if err != nil {
			return handler.NewErrWithStatus(http.StatusInternalServerError, fmt.Errorf("geocoding address: %w", err))
		}

		resp := handler.Response[[]services.Place]{
			Data:    &result,
			Message: "success",
		}

		if err := handler.Encode[handler.Response[[]services.Place]](resp, http.StatusOK, w); err != nil {
			return handler.NewErrWithStatus(http.StatusInternalServerError, err)
		}

		return nil
	})
}

type RouteRequest struct {
	Locations        []valhalla.LocationRequest  `json:"locations"`
	ExcludeLocations []valhalla.ExcludeLocations `json:"exclude_locations"`
	Costing          valhalla.Costing            `json:"costing"`
	CostingOptions   *valhalla.CostingOptions    `json:"costing_options,omitempty"`
	Language         *string                     `json:"language,omitempty"`
	Alternates       *int                        `json:"alternates,omitempty"`
}

func (r RouteRequest) Validate() error {
	if len(r.Locations) < 2 {
		return errors.New("at least 2 locations must be provided")
	}
	if !r.Costing.IsValid() {
		return errors.New(fmt.Sprintf("costing %q is invalid", r.Costing))
	}
	return nil
}

// ToValhallaRequest converts a API request to a [valhalla.RouteRequest],
// and applies default values if necessary.
func (r RouteRequest) ToValhallaRequest() valhalla.RouteRequest {
	// Default values
	language := "fr-FR"
	alternates := 2

	if r.Language != nil {
		language = *r.Language
	}

	if r.Alternates != nil {
		alternates = *r.Alternates
	}

	return valhalla.RouteRequest{
		Locations:        r.Locations,
		ExcludeLocations: r.ExcludeLocations,
		Costing:          r.Costing,
		CostingOptions:   r.CostingOptions,
		Language:         language,
		Alternates:       alternates,
	}
}

// @Summary Calcul d'itinéraires.
// @Description Calcule un ou plusieurs itinéraires à partir de plusieurs localisations.
// @Tags routing
// @Accept json
// @Produce json
// @Param routeRequest body RouteRequest true "Liste de localisation accompagnés d'options permettant de paramétrer le calcul d'itinéraire. Optionnels: 'language', 'costing_options', 'alternates', 'exclude_locations'."
// @Success 200 {object} handler.Response[[]services.Trip]
// @Failure 400 {object} ErrResponse "Corps de la requête invalide"
// @Failure 500 {object} ErrResponse "Erreur interne du serveur"
// @Router /route [post]
func (s *Server) routeHandler() http.HandlerFunc {
	return handler.Handler(func(w http.ResponseWriter, r *http.Request) error {
		req, err := handler.Decode[RouteRequest](r)
		if err != nil {
			return handler.NewErrWithStatus(http.StatusBadRequest, err)
		}

		valhallaReq := req.ToValhallaRequest()

		route, err := s.routingService.CalculateRoute(r.Context(), valhallaReq)
		if err != nil {
			return handler.NewErrWithStatus(http.StatusInternalServerError, err)
		}

		resp := handler.Response[[]services.Trip]{
			Data:    route,
			Message: "success",
		}

		if err := handler.Encode[handler.Response[[]services.Trip]](resp, http.StatusOK, w); err != nil {
			return handler.NewErrWithStatus(http.StatusInternalServerError, err)
		}

		return nil
	})
}

type AddressResponse struct {
	DisplayName string `json:"display_name,omitempty"`
}

// @Summary Adresse à partir de coordonnées
// @Description Retourne l'adresse correspondante aux coordonnées géographiques fournies.
// @Tags geocoding
// @Accept json
// @Produce json
// @Param lat query number true "Latitude (ex: 49.0677)"
// @Param lon query number true "Longitude (ex: -0.6658)"
// @Success 200 {object} AddressResponse "Adresse trouvée à partir des coordonnées"
// @Failure 400 {object} ErrResponse "Paramètre de requête manquant ou invalide"
// @Failure 404 {object} ErrResponse "Aucune adresse trouvée pour les coordonnées spécifiées"
// @Failure 500 {object} ErrResponse "Erreur interne du serveur"
// @Router /address [get]
func (s *Server) addressHandler() http.HandlerFunc {
	return handler.Handler(func(w http.ResponseWriter, r *http.Request) error {
		query := r.URL.Query()
		if !query.Has("lat") {
			w.WriteHeader(http.StatusBadRequest)
			return nil
		}

		if !query.Has("lon") {
			w.WriteHeader(http.StatusBadRequest)
			return nil
		}

		lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
		lon, _ := strconv.ParseFloat(r.URL.Query().Get("lon"), 64)

		resp, err := s.geocodingService.Reverse(r.Context(), lat, lon)
		if err != nil {
			return handler.NewErrWithStatus(http.StatusInternalServerError, err)
		}

		// Récupérer le premier élément de feature et retourner son displayname dans la struct
		if resp == nil || len(resp.Features) == 0 || resp.Features[0].Properties.DisplayName == "" {
			return handler.NewErrWithStatus(http.StatusNotFound, fmt.Errorf("failed to retrieve data"))
		}

		address := AddressResponse{
			DisplayName: resp.Features[0].Properties.DisplayName,
		}

		if err := handler.Encode(address, http.StatusOK, w); err != nil {
			return handler.NewErrWithStatus(http.StatusInternalServerError, err)
		}

		return nil
	})
}
