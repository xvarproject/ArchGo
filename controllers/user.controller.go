package controllers

import (
	"ArchGo/constants"
	"ArchGo/logger"
	"ArchGo/utils"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ArchGo/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserController struct {
	DB *gorm.DB
	cs *CreditSystem
}

func NewUserController(DB *gorm.DB, creditSystem *CreditSystem) UserController {
	return UserController{DB, creditSystem}
}

func (uc *UserController) GetMe(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(models.User)

	userResponse := &models.UserResponse{
		ID:        currentUser.ID,
		Name:      currentUser.Name,
		Email:     currentUser.Email,
		Photo:     currentUser.Photo,
		Role:      currentUser.Role,
		Verified:  currentUser.Verified,
		Balance:   currentUser.Balance,
		Provider:  currentUser.Provider,
		CreatedAt: currentUser.CreatedAt,
		UpdatedAt: currentUser.UpdatedAt,
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"user": userResponse}})
}

func (uc *UserController) AdminBulkCreateRechargingCards(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(models.User)
	if currentUser.Role != constants.RoleAdmin {
		ctx.JSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": "Only Admin can create recharging cards"})
		return
	}
	var payload *models.BulkCreateRechargingCardsInput

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	var rechargingCards []models.RechargingCard
	i := int64(0)
	for ; i < payload.Amount; i++ {
		card := models.RechargingCard{
			Password:       utils.GenerateRandomString(10),
			BalanceInCents: payload.BalanceInCents,
			Status:         constants.RechargingCardActive,
			VendorName:     currentUser.Email,
		}
		rechargingCards = append(rechargingCards, card)
	}

	err := uc.DB.Create(&rechargingCards).Error
	if err != nil {
		logger.Warning("Failed to create recharging cards", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Failed to create recharging cards"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"rechargingCards": rechargingCards}})
}

func (uc *UserController) AdminFindAllCards(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(models.User)
	if currentUser.Role != constants.RoleAdmin {
		ctx.JSON(http.StatusForbidden, gin.H{"status": "fail", "message": "You are not authorized to perform this action"})
		return
	}

	var count int64
	uc.DB.Model(&models.RechargingCard{}).Count(&count)
	var page = ctx.DefaultQuery("page", "1")
	var limit = ctx.DefaultQuery("limit", "10")

	intPage, _ := strconv.Atoi(page)
	intLimit, _ := strconv.Atoi(limit)
	offset := (intPage - 1) * intLimit

	var rechargingCards []models.RechargingCard
	results := uc.DB.Offset(offset).Limit(intLimit).Find(&rechargingCards)
	if results.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": results.Error})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "results": len(rechargingCards),
		"total_records": count, "data": rechargingCards})
}

func (uc *UserController) AdminDeactivateCard(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(models.User)
	if currentUser.Role != constants.RoleAdmin {
		ctx.JSON(http.StatusForbidden, gin.H{"status": "fail", "message": "You are not authorized to perform this action"})
		return
	}
	rechargingCardId := ctx.Param("rechargingCardId")

	cardToDeactivate := models.RechargingCard{}
	result := uc.DB.First(&cardToDeactivate, "id = ?", rechargingCardId)
	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": "No recharging card with that id exists"})
		return
	}
	result = uc.DB.Model(&cardToDeactivate).Update("status", constants.RechargingCardInactive)
	if result.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": result.Error})
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

func (uc *UserController) AdminUpdateBalance(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(models.User)
	if currentUser.Role != constants.RoleAdmin {
		ctx.JSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": "Only Admin can update credits"})
		return
	}
	var payload *models.UpdateBalanceInput

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	newBalance, err := uc.cs.UpdateBalanceByUserEmail(payload.Email, payload.Amount)
	if err != nil {
		logger.Warning(err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Failed to update credits"})
		return
	}
	now := time.Now()
	bill := models.BillingHistory{
		Operator:          currentUser.Email,
		AccountEmail:      payload.Email,
		Amount:            payload.Amount,
		TransactionType:   constants.TransactionTypeAdmin,
		TransactionDetail: "Admin updates balance",
		TransactionTime:   now,
	}
	err = uc.DB.Create(&bill).Error
	if err != nil {
		logger.Warning("Failed to create billing history", err)
	}
	logger.Info("Create billing history", bill)

	var user models.User
	err = uc.DB.First(&user, "email = ?", payload.Email).Error
	if err != nil {
		logger.Warning(err.Error())
	}
	var firstName = user.Name

	if strings.Contains(firstName, " ") {
		firstName = strings.Split(firstName, " ")[1]
	}

	emailData := utils.EmailData{
		FirstName: firstName,
		Subject:   "Pointer.ai 充值提醒",
		Amount:    payload.Amount,
		Balance:   newBalance,
	}
	go func() {
		utils.SendEmail(&user, &emailData, "chargeSucceed.html")

	}()
	logger.Info("Charge balance of user", payload.Email, "for amount", payload.Amount, "newBalance: ", newBalance)

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"updated_email": payload.Email, "new_balance": newBalance}})
}

func (uc *UserController) RechargeMyselfWithRechargingCard(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(models.User)
	var payload *models.ChargeMyselfWithRechargingCardInput

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	var rechargingCard models.RechargingCard
	err := uc.DB.First(&rechargingCard, payload.CardID).Error
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": "No recharging card with that id exists"})
		return
	}

	newBalance, err := uc.cs.rechargeUserEmailWithChargingCardId(currentUser.Email, payload.CardID, payload.Password)
	if err != nil {
		logger.Warning(err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "fail", "message": "Failed to recharge"})
		return
	}

	now := time.Now()
	// Billing history is non-critical at this point, so we don't put it in the same transaction with the recharge process
	bill := models.BillingHistory{
		Operator:          currentUser.Email,
		AccountEmail:      currentUser.Email,
		Amount:            rechargingCard.BalanceInCents,
		TransactionType:   constants.TransactionTypeRecharge,
		TransactionDetail: fmt.Sprintf("Recharge with recharging card %s", rechargingCard.ID),
		TransactionTime:   now,
	}
	err = uc.DB.Create(&bill).Error
	if err != nil {
		logger.Warning("Failed to create billing history", err)
	}
	logger.Info("Create billing history", bill)

	var firstName = currentUser.Name

	if strings.Contains(firstName, " ") {
		firstName = strings.Split(firstName, " ")[1]
	}

	emailData := utils.EmailData{
		FirstName: firstName,
		Subject:   "Pointer.ai 充值提醒",
		Amount:    rechargingCard.BalanceInCents,
		Balance:   newBalance,
	}
	go func() {
		utils.SendEmail(&currentUser, &emailData, "chargeSucceed.html")
	}()
	logger.Info("Recharge balance of user", currentUser.Email, "for amount", rechargingCard.BalanceInCents, "using card, newBalance: ", newBalance)

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"recharged_email": currentUser.Email,
		"recharge_amount": rechargingCard.BalanceInCents, "new_balance": newBalance}})
}
