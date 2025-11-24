package api

import (
    "errors"
    "log"
    "net/http"
    "time"

    "github.com/axellelanca/urlshortener/internal/models"
    "github.com/axellelanca/urlshortener/internal/services"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

// ClickEventsChannel est un channel bufferisé pour l'envoi asynchrone des événements de clic
var ClickEventsChannel chan models.ClickEvent

// SetupRoutes configure toutes les routes de l'API Gin et initialise le channel avec la taille du buffer
func SetupRoutes(router *gin.Engine, linkService *services.LinkService, clickChannelBuffer int) {
    if ClickEventsChannel == nil {
        ClickEventsChannel = make(chan models.ClickEvent, clickChannelBuffer)
    }

    router.GET("/health", HealthCheckHandler)
    router.POST("/api/v1/links", CreateShortLinkHandler(linkService))
    router.GET("/api/v1/links/:shortCode/stats", GetLinkStatsHandler(linkService))
    router.GET("/:shortCode", RedirectHandler(linkService))
}

// HealthCheckHandler retourne simplement {"status": "ok"}
func HealthCheckHandler(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// CreateLinkRequest est le JSON attendu lors de la création d'un lien
type CreateLinkRequest struct {
    LongURL string `json:"long_url" binding:"required,url"`
}

// CreateShortLinkHandler crée un lien court et renvoie le résultat JSON
func CreateShortLinkHandler(linkService *services.LinkService) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req CreateLinkRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "error":   "Invalid request",
                "message": err.Error(),
            })
            return
        }

        // Validation de la longueur de l'URL
        if len(req.LongURL) > 2048 {
            c.JSON(http.StatusBadRequest, gin.H{
                "error":   "Invalid URL",
                "message": "URL is too long (maximum 2048 characters)",
            })
            return
        }

        link, err := linkService.CreateLink(req.LongURL)
        if err != nil {
            log.Printf("Error creating link: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{
                "error":   "Internal server error",
                "message": "Failed to create short link",
            })
            return
        }

        c.JSON(http.StatusCreated, gin.H{
            "short_code":     link.ShortCode,
            "long_url":       link.LongURL,
            "full_short_url": "http://localhost:8080/" + link.ShortCode,
            "created_at":     link.CreatedAt,
        })
    }
}

// RedirectHandler redirige vers l'URL longue et enregistre le clic de façon asynchrone
func RedirectHandler(linkService *services.LinkService) gin.HandlerFunc {
    return func(c *gin.Context) {
        shortCode := c.Param("shortCode")

        // Validation du shortCode
        if len(shortCode) == 0 || len(shortCode) > 10 {
            c.JSON(http.StatusBadRequest, gin.H{
                "error":   "Invalid short code",
                "message": "Short code must be between 1 and 10 characters",
            })
            return
        }

        link, err := linkService.GetLinkByShortCode(shortCode)
        if err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                c.JSON(http.StatusNotFound, gin.H{
                    "error":   "Link not found",
                    "message": "The requested short link does not exist",
                })
                return
            }
            log.Printf("Error retrieving link for %s: %v", shortCode, err)
            c.JSON(http.StatusInternalServerError, gin.H{
                "error":   "Internal server error",
                "message": "Failed to retrieve link",
            })
            return
        }

        clickEvent := models.ClickEvent{
            LinkID:    link.ID,
            Timestamp: time.Now(),
            IPAddress:        c.ClientIP(),
            UserAgent: c.Request.UserAgent(),
        }

        select {
        case ClickEventsChannel <- clickEvent:
            // clic envoyé au worker
        default:
            log.Printf("Warning: ClickEventsChannel is full, dropping click event for %s.", shortCode)
        }

        c.Redirect(http.StatusFound, link.LongURL)
    }
}

// GetLinkStatsHandler renvoie les statistiques (nombre total de clics) pour un lien donné
func GetLinkStatsHandler(linkService *services.LinkService) gin.HandlerFunc {
    return func(c *gin.Context) {
        shortCode := c.Param("shortCode")

        // Validation du shortCode
        if len(shortCode) == 0 || len(shortCode) > 10 {
            c.JSON(http.StatusBadRequest, gin.H{
                "error":   "Invalid short code",
                "message": "Short code must be between 1 and 10 characters",
            })
            return
        }

        link, totalClicks, err := linkService.GetLinkStats(shortCode)
        if err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                c.JSON(http.StatusNotFound, gin.H{
                    "error":   "Link not found",
                    "message": "The requested short link does not exist",
                })
                return
            }
            log.Printf("Error retrieving stats for %s: %v", shortCode, err)
            c.JSON(http.StatusInternalServerError, gin.H{
                "error":   "Internal server error",
                "message": "Failed to retrieve statistics",
            })
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "short_code":   link.ShortCode,
            "long_url":    link.LongURL,
            "total_clicks": totalClicks,
            "created_at":   link.CreatedAt,
        })
    }
}
