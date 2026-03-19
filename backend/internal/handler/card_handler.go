package handler

import (
	"strconv"

	"backend/internal/middleware"
	"backend/internal/service"
	"backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type CardHandler struct {
	cardService service.CardService
}

func NewCardHandler(cardService service.CardService) *CardHandler {
	return &CardHandler{cardService: cardService}
}

func (h *CardHandler) List(c *gin.Context) {
	agentID, ok := middleware.GetAgentID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("failed to resolve current user"))
		return
	}

	page, err := parsePositiveInt(c.Query("page"), 1)
	if err != nil {
		utils.Error(c, utils.BadRequestError("invalid page"))
		return
	}
	pageSize, err := parsePositiveInt(c.Query("pageSize"), 20)
	if err != nil {
		utils.Error(c, utils.BadRequestError("invalid pageSize"))
		return
	}

	var status *int
	if rawStatus := c.Query("status"); rawStatus != "" {
		value, err := strconv.Atoi(rawStatus)
		if err != nil {
			utils.Error(c, utils.BadRequestError("invalid status"))
			return
		}
		status = &value
	}

	result, err := h.cardService.List(c.Request.Context(), service.CardListInput{
		AgentID:  agentID,
		Page:     page,
		PageSize: pageSize,
		Status:   status,
		Keyword:  c.Query("keyword"),
	})
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, "cards fetched", result)
}

func (h *CardHandler) Detail(c *gin.Context) {
	agentID, ok := middleware.GetAgentID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("failed to resolve current user"))
		return
	}

	cardID, err := parseUintParam(c.Param("id"))
	if err != nil {
		utils.Error(c, utils.BadRequestError("invalid card id"))
		return
	}

	card, err := h.cardService.Detail(c.Request.Context(), agentID, cardID)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, "card fetched", card)
}

func (h *CardHandler) Destroy(c *gin.Context) {
	agentID, ok := middleware.GetAgentID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("failed to resolve current user"))
		return
	}

	cardID, err := parseUintParam(c.Param("id"))
	if err != nil {
		utils.Error(c, utils.BadRequestError("invalid card id"))
		return
	}

	result, err := h.cardService.Destroy(c.Request.Context(), agentID, cardID)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, "card destroyed", result)
}

func (h *CardHandler) Stats(c *gin.Context) {
	agentID, ok := middleware.GetAgentID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("failed to resolve current user"))
		return
	}

	stats, err := h.cardService.Stats(c.Request.Context(), agentID)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, "card stats fetched", stats)
}

func parsePositiveInt(value string, defaultValue int) (int, error) {
	if value == "" {
		return defaultValue, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	if parsed <= 0 {
		return 0, strconv.ErrSyntax
	}
	return parsed, nil
}

func parseUintParam(value string) (uint, error) {
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(parsed), nil
}
