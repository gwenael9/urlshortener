# TP Go Final : URL Shortener

## Objectif du Projet

Ce TP vous met au défi de construire un service web performant de raccourcissement et de gestion d'URLs en Go. Votre application permettra de transformer une URL longue en une URL courte et unique. Chaque fois qu'une URL courte est visitée, le système redirigera instantanément l'utilisateur vers l'URL originale tout en enregistrant le clic de manière asynchrone, pour ne jamais ralentir la redirection.

Le service inclura également un moniteur pour vérifier périodiquement la disponibilité des URLs longues et notifier tout changement d'état. L'interaction se fera via une API RESTful et une interface en ligne de commande (CLI) complète.

## Connaissances Mobilisées

Ce projet est une synthèse complète et pratique de tous les concepts abordés durant ce module de Go (normalement il n'y aura pas trop de surprise) :

- Syntaxe Go de base (structs, maps, boucles, conditions, etc.)
- Concurrence (Goroutines, Channels) pour les tâches asynchrones et non-bloquantes
- Interfaces CLI avec [Cobra](https://cobra.dev/)
- Gestion des erreurs
- Manipulation de données (JSON) pour les APIs
- APIs RESTful avec le framework web [Gin](https://gin-gonic.com/)
- Persistance des données avec l'ORM [GORM](https://gorm.io/) et SQLite
- Gestion de configuration avec [Viper](https://github.com/spf13/viper)
- Design patterns courants (Repository, Service) pour une architecture propre

## Fonctionnalités Attendues

### Core Features (Obligatoires)

1. **Raccourcissement d'URLs** :

- Générer des codes courts uniques (6 caractères alphanumériques).
- Gérer les collisions lors de la génération de codes via une logique de retry.

2. **Redirection instantanée** :

- Rediriger les utilisateurs vers l'URL originale sans latence (code HTTP 302).
- Analytics asynchrones :
- Enregistrer les détails de chaque clic en arrière-plan via des Goroutines et un Channel bufferisé. La redirection ne doit jamais être bloquée par l'enregistrement du clic.

3. **Surveillance de l'état des URLs** :

- Le service doit vérifier périodiquement (intervalle configurable via Viper) si les URLs longues sont toujours accessibles (réponse HTTP 200/3xx).
- Si l'état d'une URL change (accessible leftrightarrow inaccessible), une fausse notification doit être générée dans les logs du serveur (ex: "[NOTIFICATION] L'URL ... est maintenant INACCESSIBLE.").

4. **APIs REST (via Gin)** :

- `GET /health` : Vérifie l'état de santé du service.
- `POST /api/v1/links` : Crée une nouvelle URL courte (attend un JSON {"long_url": "..."}).
- `GET /{shortCode}` : Gère la redirection et déclenche l'analytics asynchrone.
- `GET /api/v1/links/{shortCode}/stats` : Récupère les statistiques d'un lien (nombre total de clics).

5. **Interface CLI (via Cobra)** :

- `./url-shortener run-server` : Lance le serveur API, les workers de clics et le moniteur d'URLs.
- `./url-shortener create --url="https://..."` : Crée une URL courte depuis la ligne de commande.
- `./url-shortener stats --code="xyz123"` : Affiche les statistiques d'un lien donné.
- `./url-shortener migrate` : Exécute les migrations GORM pour la base de données.

6. **Features Avancées (Bonus - si le temps le permet)**

- URLs personnalisées : Permettre aux utilisateurs de proposer leur propre alias (ex: /mon-alias-perso).
- Expiration des liens : Les URLs courtes peuvent avoir une durée de vie limitée.
- Rate limiting : Protection simple par IP pour les créations de liens.

## Architecture du Projet

Le projet suit une structure modulaire classique pour les applications Go, qui sépare bien les différences préoccupations du projet :

```
url-shortener/
├── cmd/
│   ├── root.go             # Initialise la commande racine Cobra et ses sous-commandes
│   ├── server/
│   │   └── server.go       # Logique pour la commande 'run-server' (lance le serveur Gin, les workers de clics, le moniteur)
│   └── cli/
│       ├── create.go       # Logique pour la commande 'create' (crée un lien via CLI)
│       ├── stats.go        # Logique pour la commande 'stats' (affiche les statistiques d'un lien via CLI)
│       └── migrate.go      # Logique pour la commande 'migrate' (exécute les migrations GORM)
├── internal/
│   ├── api/
│   │   └── handlers.go     # Fonctions de gestion des requêtes HTTP (handlers Gin pour les routes API)
│   ├── models/
│   │   ├── link.go         # Définition de la structure GORM 'Link'
│   │   └── click.go        # Définition de la structure GORM 'Click'
│   ├── services/
│   │   ├── link_service.go # Logique métier pour les liens (ex: génération de code, validation)
│   │   └── click_service.go # Logique métier pour les clics (optionnel, peut être directement dans le worker si simple)
│   ├── workers/
│   │   └── click_worker.go # Goroutine et logique pour l'enregistrement asynchrone des clics
│   ├── monitor/
│   │   └── url_monitor.go  # Logique pour la surveillance périodique de l'état des URLs
│   ├── config/
│   │   └── config.go       # Chargement et structure de la configuration de l'application (Viper)
│   └── repository/
│       ├── link_repository.go # Interface et implémentation GORM pour les opérations CRUD sur 'Link'
│       └── click_repository.go # Interface et implémentation GORM pour les opérations CRUD sur 'Click'
├── configs/
│   └── config.yaml         # Fichier de configuration par défaut pour Viper
├── go.mod                  # Fichier de module Go (liste des dépendances du projet)
├── go.sum                  # Sommes de contrôle pour la sécurité des dépendances
└── README.md               # Documentation du projet (installation, utilisation, etc.)

```

## Démarrage et Utilisation du Projet

Suivez ces étapes pour mettre en place le projet et tester votre application (quand elle fonctionnera, évidemment).

### 1. Préparation Initiale

1. **Clonez le dépôt :**

```bash
git clone https://github.com/axellelanca/urlshortener.git
cd urlshortener # Naviguez vers le dossier du projet cloné
```

2. **Téléchargez et nettoyez les dépendances :**

```bash
go mod tidy
```

## Pour tester votre projet :

### Construisez l'exécutable :

Ceci compile votre application et crée un fichier url-shortener à la racine du projet.

```bash
go build -o url-shortener
```

Désormais, toutes les commandes seront lancées avec ./url-shortener.

### Initialisation de la Base de Données

Avant de démarrer le serveur, créez le fichier de base de données SQLite et ses tables :

1.  **Exécutez les migrations :**

```bash
./url-shortener migrate
```

Un message de succès confirmera la création des tables. Un fichier url_shortener.db sera créé à la racine du projet.

### Lancer le Serveur et les Processus de Fond

C'est l'étape qui démarre le cœur de votre application. Elle démarre le serveur web, les workers qui enregistrent les clics, et le moniteur d'URLs.

Démarrez le service :

```bash
./url-shortener run-server
```

Laissez ce terminal ouvert et actif. Il affichera les logs du serveur HTTP, des workers de clics et du moniteur d'URLs.

### 4. Interagir avec le Service (Utilise un **Nouveau Terminal**)

Ouvre une **nouvelle fenêtre de terminal** pour exécuter les commandes CLI et tester les APIs pendant que le serveur est en cours d'exécution.

#### 4.1. Créer une URL courte (via la CLI)

Raccourcis une URL longue en utilisant la commande `create` :

```bash
./url-shortener create --url="https://www.example.com/ma-super-url-de-test-pour-le-tp-go-final"
```

Tu obtiendras un message similaire à :

```bash
URL courte créée avec succès:
Code: XYZ123
URL complète: http://localhost:8080/XYZ123
```

Note le Code (ex: XYZ123) et l'URL complète pour les étapes suivantes.

#### 4.2. Accéder à l'URL courte (via Navigateur)

1. Ouvre ton navigateur web et accède à l'URL complète que tu as obtenue (par exemple, http://localhost:8080/XYZ123).
2. Le navigateur devrait te rediriger instantanément vers l'URL longue originale. Dans le terminal où le serveur tourne (./url-shortener run-server), tu devrais voir des logs indiquant qu'un clic a été détecté et envoyé au worker asynchrone.

#### 4.3. Consulter les Statistiques (via la CLI)

Vérifie combien de fois ton URL courte a été visitée :

1. Affiche les statistiques :

```
./url-shortener stats --code="8LKkDt"
```

Le terminal affichera :

```
Statistiques pour le code court: XYZ123
URL longue: [https://www.example.com/ma-super-url-de-test-pour-le-tp-go-final](https://www.example.com/ma-super-url-de-test-pour-le-tp-go-final)
Total de clics: 1
```

(Le nombre de clics augmentera à chaque fois que tu accèderas à l'URL courte via ton navigateur).

#### 4.4. Tester l'API de Santé (via curl)

Vérifie si ton serveur est bien opérationnel :

1. Exécute la commande curl :

```
curl http://localhost:8080/health
```

Tu devrais obtenir :

```
{"status":"ok"}
```

#### 4.5. Observer le Moniteur d'URLs

Le moniteur fonctionne en arrière-plan et vérifie la disponibilité des URLs longues toutes les 5 minutes (par défaut).

Observe les logs dans le terminal où run-server tourne. Si l'état d'une URL que tu as raccourcie change (par exemple, si le site devient inaccessible), tu verras un message [NOTIFICATION] similaire à :

```
[NOTIFICATION] Le lien XYZ123 ([https://url-hors-ligne.com](https://url-hors-ligne.com)) est passé de ACCESSIBLE à INACCESSIBLE !
```

(Pour tester cela, tu pourrais raccourcir une URL vers un site que tu sais hors ligne ou une adresse IP inexistante, et attendre l'intervalle de surveillance.)

### 5. Arrêter le Serveur

Quand tu as terminé tes tests et que tu souhaites arrêter le service :

1. Dans le terminal où ./url-shortener run-server tourne, appuie sur :

```
Ctrl + C
```

Tu verras des logs confirmant l'arrêt propre du serveur.

### Contributeurs

Merci à toutes les personnes qui ont contribué à ce projet ! 🎉

<table>
  <tr>
    <td align="center">
      <a href="https://github.com/gwenael9">
        <sub><b>Gwenael GEHO</b></sub>
      </a>
    </td>
    <td align="center">
      <a href="https://github.com/VoutsaStevie">
        <sub><b>Voutsa Stevie</b></sub>
      </a>
    </td>
    <td align="center">
      <a href="https://github.com/norab0">
        <sub><b>Nora</b></sub>
      </a>
    </td>
      <td align="center">
      <a href="https://github.com/iimAtomic">
        <sub><b>VEGBA Lux</b></sub> 
      </a>
    </td>
  </tr>
</table>
