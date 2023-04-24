package main

import (
	"ArchGo/controllers"
	"ArchGo/initializers"
	"ArchGo/openai-config"
	"ArchGo/routes"
	"github.com/alecthomas/kong"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

var (
	server              *gin.Engine
	AuthController      controllers.AuthController
	AuthRouteController routes.AuthRouteController

	UserController      controllers.UserController
	UserRouteController routes.UserRouteController

	PostController      controllers.PostController
	PostRouteController routes.PostRouteController

	ChatController      controllers.ChatController
	ChatRouteController routes.ChatRouteController
	CreditSystem        *controllers.CreditSystem

	ConversationHistoryController      controllers.ConversationHistoryController
	ConversationHistoryRouteController routes.ConversationHistoryRouteController
)

func init() {
	conf, err := initializers.LoadEnv(".")
	if err != nil {
		log.Fatal("🚀 Could not load environment variables", err)
	}

	initializers.ConnectDB(&conf)

	CreditSystem = controllers.NewCreditSystem(initializers.DB)

	AuthController = controllers.NewAuthController(initializers.DB)
	AuthRouteController = routes.NewAuthRouteController(AuthController)

	UserController = controllers.NewUserController(initializers.DB, CreditSystem)
	UserRouteController = routes.NewRouteUserController(UserController)

	PostController = controllers.NewPostController(initializers.DB)
	PostRouteController = routes.NewRoutePostController(PostController)

	ChatController = controllers.NewChatController(CreditSystem)
	ChatRouteController = routes.NewChatRouteController(ChatController)

	ConversationHistoryController = controllers.NewConversationHistoryController(initializers.DB)
	ConversationHistoryRouteController = routes.NewConversationHistoryRouteController(ConversationHistoryController)

	server = gin.Default()

}

func main() {
	kong.Parse(&openai_config.CLI)

	conf, err := initializers.LoadEnv(".")
	if err != nil {
		log.Fatal("🚀 Could not load environment variables", err)
	}

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowCredentials = true

	server.Use(cors.New(corsConfig))

	router := server.Group("/api")
	router.GET("/healthcheck", func(ctx *gin.Context) {
		message := "Welcome to ChatGPT!"
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": message})
	})

	AuthRouteController.AuthRoute(router)
	UserRouteController.UserRoute(router)
	PostRouteController.PostRoute(router)
	ChatRouteController.ChatRoute(router)
	ConversationHistoryRouteController.ConversationHistoryRoute(router)
	log.Fatal(server.Run(":" + conf.ServerPort))
}
