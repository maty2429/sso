package handler

import (
	"net/http"
	"sso/internal/core/service"
	"sso/internal/utils"

	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	projectService *service.ProjectService
}

func NewProjectHandler(projectService *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
	}
}

type createProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	FrontendURL string `json:"frontend_url"`
}

func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var req createProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.projectService.CreateProject(c.Request.Context(), req.Name, req.Code, req.Description, req.FrontendURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, project)
}

type addMemberRequest struct {
	Rut   interface{} `json:"rut" binding:"required"`
	Roles []int       `json:"roles" binding:"required"`
}

func (h *ProjectHandler) AddMember(c *gin.Context) {
	projectCode := c.Param("projectCode")
	var req addMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var rutInt int
	var err error

	// Handle RUT input (int or string)
	switch v := req.Rut.(type) {
	case float64:
		rutInt = int(v)
	case string:
		rutInt, _, err = utils.ParseRut(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rut format: " + err.Error()})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rut format"})
		return
	}

	if err := h.projectService.AddMember(c.Request.Context(), projectCode, rutInt, req.Roles); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "member added successfully"})
}
