package controllers

import (
	"ArchGo/constants"
	"ArchGo/logger"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ArchGo/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PostController struct {
	DB *gorm.DB
}

func NewPostController(DB *gorm.DB) PostController {
	return PostController{DB}
}

func (pc *PostController) CreatePost(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(models.User)
	if currentUser.Role != constants.RoleAdmin {
		ctx.JSON(http.StatusForbidden, gin.H{"status": "fail", "message": "You are not authorized to perform this action"})
		return
	}
	var payload *models.CreatePostInput

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	now := time.Now()
	newPost := models.Post{
		Title:   payload.Title,
		Content: payload.Content,
	}
	newPost.CreatedAt = now
	newPost.UpdatedAt = now

	result := pc.DB.Create(&newPost)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key") {
			ctx.JSON(http.StatusConflict, gin.H{"status": "fail", "message": "Post with that title already exists"})
			return
		}
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": result.Error.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": newPost})
}

func (pc *PostController) UpdatePost(ctx *gin.Context) {
	postId := ctx.Param("postId")
	currentUser := ctx.MustGet("currentUser").(models.User)
	if currentUser.Role != constants.RoleAdmin {
		ctx.JSON(http.StatusForbidden, gin.H{"status": "fail", "message": "You are not authorized to perform this action"})
		return
	}
	var payload *models.UpdatePostInput
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "fail", "message": err.Error()})
		return
	}
	var updatedPost models.Post
	result := pc.DB.First(&updatedPost, "id = ?", postId)
	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": "No post with that title exists"})
		return
	}
	postToUpdate := models.Post{
		Title:   payload.Title,
		Content: payload.Content,
	}
	postToUpdate.CreatedAt = updatedPost.CreatedAt
	postToUpdate.UpdatedAt = time.Now()

	pc.DB.Model(&updatedPost).Updates(postToUpdate)

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": updatedPost})
}

func (pc *PostController) FindLatestPost(ctx *gin.Context) {
	// Get the user with the largest ID
	var largestIDPost models.Post
	if err := pc.DB.Order("id desc").First(&largestIDPost).Error; err != nil {
		logger.Warning("Error retrieving latest post: %v\n", err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": largestIDPost})
}

func (pc *PostController) FindPosts(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(models.User)
	if currentUser.Role != constants.RoleAdmin {
		ctx.JSON(http.StatusForbidden, gin.H{"status": "fail", "message": "You are not authorized to perform this action"})
		return
	}

	var count int64
	pc.DB.Model(&models.Post{}).Count(&count)

	var page = ctx.DefaultQuery("page", "1")
	var limit = ctx.DefaultQuery("limit", "10")

	intPage, _ := strconv.Atoi(page)
	intLimit, _ := strconv.Atoi(limit)
	offset := (intPage - 1) * intLimit

	var posts []models.Post
	results := pc.DB.Offset(offset).Limit(intLimit).Find(&posts)
	if results.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": results.Error})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "results": len(posts), "total_records": count, "data": posts})
}

func (pc *PostController) DeletePost(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(models.User)
	if currentUser.Role != constants.RoleAdmin {
		ctx.JSON(http.StatusForbidden, gin.H{"status": "fail", "message": "You are not authorized to perform this action"})
		return
	}
	postId := ctx.Param("postId")

	result := pc.DB.Delete(&models.Post{}, "id = ?", postId)

	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": "No post with that title exists"})
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}
