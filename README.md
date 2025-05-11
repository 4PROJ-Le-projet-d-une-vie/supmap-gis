# supmap-gis

## 1. Introduction

Le microservice **supmap-gis** fournit des fonctionnalités avancées de traitement géographique pour l’écosystème Supmap : calcul d’itinéraires multimodaux, géocodage (et géocodage inverse), et prise en compte dynamique des incidents.

### 1.1. Rôles principaux

- **Exposer une API HTTP** : endpoints `/geocode`, `/address`, `/route`, `/health` via net/http (handlers personnalisés).
- **Routage multimodal** : calcul d’itinéraires avec options (type de véhicule, exclusions dynamiques, alternatives…).
- **Géocodage** : conversion d’adresses en coordonnées GPS (et inversement).
- **Prise en compte des incidents** : intégration du service interne supmap-incidents pour éviter des routes lors du calcul d’itinéraires.

### 1.2. Fonctionnement général

- **Entrée principale** : `cmd/api/main.go`
  - Instancie la configuration, le logger, les clients HTTP externes (Nominatim, Valhalla, supmap-incidents).
  - Compose les services métiers : GeocodingService, IncidentsService, RoutingService, injectés au serveur HTTP.
  - Lance l’API HTTP via `internal/api/server.go` (serveur, routing, doc Swagger).

- **Handlers HTTP** : `internal/api/handlers.go`
  - Chaque endpoint a un handler dédié, qui :
    - Valide les paramètres ou le corps de la requête.
    - Appelle le service métier correspondant.
    - Formate la réponse ou l’erreur.

- **Services métiers** : `internal/services/`
  - **GeocodingService** : encapsule les appels à Nominatim, standardise les résultats.
  - **RoutingService** : orchestre le calcul d’itinéraire avec Valhalla, l’enrichit en excluant dynamiquement les routes touchées par des incidents bloquants via IncidentsService.
  - **IncidentsService** : interroge supmap-incidents pour identifier les points à éviter.
  - **Polyline decoding utilitaire** : interprète les polylines Valhalla (tracé des trajets).

- **Providers** : `internal/providers/`
  - Clients HTTP spécialisés pour : Valhalla (routing), Nominatim (géocodage), supmap-incidents (incidents).

### 1.3. Technologies et principes d’architecture

- **Go natif** : `net/http`, pas de framework tiers lourd. On utilise la librairie standard.
- **Injection de dépendances explicite** : chaque service reçoit ses clients et dépendances à l’instanciation.
- **Découplage fort** : chaque handler et service métier a une responsabilité claire, testable et extensible.
- **Contrôle des erreurs centralisé** : propagation explicite des erreurs vers le handler HTTP.

---

## 2. Architecture générale

### 2.1. Schéma d’architecture

```mermaid
flowchart TD
    subgraph API
        A["Client HTTP<br/>(frontend, autres services)"]
        B["Serveur HTTP<br/>(internal/api)"]
    end

    subgraph Services internes
        C["GeocodingService<br/>(internal/services/geocoding.go)"]
        D["RoutingService<br/>(internal/services/routing.go)"]
        E["IncidentsService<br/>(internal/services/incidents.go)"]
    end

    subgraph Providers
        F["Nominatim<br/>(internal/providers/nominatim)"]
        G["Valhalla<br/>(internal/providers/valhalla)"]
        H["supmap-incidents<br/>(internal/providers/supmap-incidents)"]
    end

    A -->|HTTP Requests| B
    B -->|geocodeHandler / addressHandler| C
    B -->|routeHandler| D
    D -->|incidentsService| E
    C -->|Appelle| F
    D -->|Appelle| G
    E -->|Appelle| H
```

### 2.2. Description des interactions internes et externes

- **Entrée dans le service** :  
  Les requêtes HTTP arrivent sur le serveur (`internal/api/server.go`).  
  Chaque endpoint (`/geocode`, `/address`, `/route`, `/health`) possède un handler dédié dans `internal/api/handlers.go`.

- **Géocodage** :  
  Les handlers `/geocode` et `/address` font appel au `GeocodingService`. Celui-ci délègue les requêtes à un client Nominatim (provider interne) pour exécuter le géocodage direct ou inverse. Les résultats sont formatés et renvoyés au client.

- **Routage** :  
  Le handler `/route` utilise le `RoutingService`.  
  Avant de lancer le calcul d’itinéraire, ce service :
    1. Demande au `IncidentsService` si des incidents sont présents autour des points de passage (via le provider supmap-incidents).
    2. Exclut dynamiquement ces zones du calcul d’itinéraire.
    3. Lance la requête de routage auprès du provider Valhalla.
    4. Formate et retourne la réponse.

- **Gestion des incidents** :  
  Le service `IncidentsService` encapsule la logique d’appel à l’API supmap-incidents :  
  Il calcule un cercle englobant tous les points de départ/d’arrivée/intermédiaires et interroge le provider pour obtenir la liste des incidents à proximité.

- **Configuration et initialisation** :  
  Le point d’entrée (`cmd/api/main.go`) :
    - Charge la configuration (`internal/config`)
    - Instancie les clients providers (Nominatim, Valhalla, supmap-incidents)
    - Compose les services métiers
    - Instancie le serveur HTTP

### 2.3. Présentation des principaux composants

- **Serveur HTTP (`internal/api/server.go`)**
    - Démarre le serveur, gère le routage des endpoints, fournit la documentation Swagger.
    - Injecte les services métiers nécessaires aux handlers.

- **Handlers HTTP (`internal/api/handlers.go`)**
    - Orchestrent la validation des entrées, l’appel aux services métiers et la gestion des erreurs/réponses.

- **Services métiers (`internal/services/`)**
    - `GeocodingService` : Encapsule la logique de géocodage via Nominatim.
    - `RoutingService` : Gère le calcul d’itinéraires, l’intégration des incidents, la transformation des réponses Valhalla.
    - `IncidentsService` : Calcule la zone à surveiller, interroge supmap-incidents, filtre les incidents pertinents qui nécessitent d'être évités.

- **Providers (`internal/providers/`)**
    - Contiennent les clients HTTP pour chaque service externe :
        - Nominatim (géocodage)
        - Valhalla (routage)
        - supmap-incidents (incidents)

- **Configuration (`internal/config`)**
    - Centralise la configuration du service (ports, hôtes, etc.).

---

