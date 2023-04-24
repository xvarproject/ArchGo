package routes

import (
	"ArchGo/controllers"
	"ArchGo/middleware"
	"github.com/gin-gonic/gin"
)

type UserRouteController struct {
	userController controllers.UserController
}

func NewRouteUserController(userController controllers.UserController) UserRouteController {
	return UserRouteController{userController}
}

func (uc *UserRouteController) UserRoute(rg *gin.RouterGroup) {

	router := rg.Group("users")
	router.Use(middleware.DeserializeUser())
	router.GET("/me", uc.userController.GetMe)
	router.POST("/admin_update_balance", uc.userController.AdminUpdateBalance)
	router.POST("/admin_bulk_create_recharging_cards", uc.userController.AdminBulkCreateRechargingCards)
	router.GET("/admin_get_recharging_cards", uc.userController.AdminFindAllCards)
	router.DELETE("/admin_deactivate_recharging_cards/:rechargingCardId", uc.userController.AdminDeactivateCard)
	router.POST("/recharge_with_recharging_card", uc.userController.RechargeMyselfWithRechargingCard)
}
