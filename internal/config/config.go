package config

import (
	"fmt"
	"log" // Pour logger les informations ou erreurs de chargement de config

	"github.com/spf13/viper" // La bibliothèque pour la gestion de configuration
)

// TODO Créer Config qui est la structure principale qui mappe l'intégralité de la configuration de l'application.
// Les tags `mapstructure` sont utilisés par Viper pour mapper les clés du fichier de config
// (ou des variables d'environnement) aux champs de la structure Go.
type Config struct {
	Server struct {
		Port    int    `mapstructure:"port"`
		BaseURL string `mapstructure:"base_url"`
	} `mapstructure:"server"`
	Database struct {
		Name string `mapstructure:"name"`
	} `mapstructure:"database"`
	Analytics struct {
		BufferSize int `mapstructure:"buffer_size"`
	} `mapstructure:"analytics"`
	Monitor struct {
		IntervalMinutes int `mapstructure:"interval_minutes"`
	} `mapstructure:"monitor"`
}

// LoadConfig charge la configuration de l'application en utilisant Viper.
// Elle recherche un fichier 'config.yaml' dans le dossier 'configs/'.
// Elle définit également des valeurs par défaut si le fichier de config est absent ou incomplet.
func LoadConfig() (*Config, error) {
	// TODO Spécifie le chemin où Viper doit chercher les fichiers de config.
	// on cherche dans le dossier 'configs' relatif au répertoire d'exécution.
	viper.AddConfigPath("configs/")

	// TODO Spécifie le nom du fichier de config (sans l'extension).
	viper.SetConfigName("config")

	// TODO Spécifie le type de fichier de config.
	viper.SetConfigType("yaml")

	// TODO : Définir les valeurs par défaut pour toutes les options de configuration.
	// Ces valeurs seront utilisées si les clés correspondantes ne sont pas trouvées dans le fichier de config
	// ou si le fichier n'existe pas.
	// server.port, server.base_url etc.
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.base_url", "http://localhost:8080")
	viper.SetDefault("database.name", "url_shortener.db")
	viper.SetDefault("analytics.buffer_size", 1000)
	viper.SetDefault("monitor.interval_minutes", 5)

	// TODO : Lire le fichier de configuration.
	err := viper.ReadInConfig()
	if err != nil {
		// Si le fichier de config est introuvable
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("Warning: Config file not found, using default configuration values.")
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	} else {
		log.Printf("Config file loaded successfully: %s", viper.ConfigFileUsed())
	}

	// TODO 4: Démapper (unmarshal) la configuration lue (ou les valeurs par défaut) dans la structure Config.
	var cfg Config
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to decode into config struct: %w", err)
	}

	// Log  pour vérifier la config chargée
	log.Printf("Configuration loaded: Server Port=%d, DB Name=%s, Analytics Buffer=%d, Monitor Interval=%dmin",
		cfg.Server.Port, cfg.Database.Name, cfg.Analytics.BufferSize, cfg.Monitor.IntervalMinutes)

	return &cfg, nil // Retourne la configuration chargée
}
