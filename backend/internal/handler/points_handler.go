package handler

import (
	"strconv"

	"backend/internal/middleware"
	"backend/internal/service"
	"backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type PointsHandler struct {
	pointsService service.PointsService
}

type rechargeApplyRequest struct {
	Amount        decimal.Decimal `json:"amount"`
	PaymentMethod string          `json:"paymentMethod" binding:"max=50"`
	PaymentProof  string          `json:"paymentProof" binding:"max=500"`
	Remark        string          `json:"remark" binding:"max=500"`
}

func NewPointsHandler(pointsService service.PointsService) *PointsHandler {
	return &PointsHandler{pointsService: pointsService}
}

func (h *PointsHandler) Balance(c *gin.Context) {
	agentID, ok := middleware.GetAgentID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("failed to resolve current user"))
		return
	}

	result, err := h.pointsService.Balance(c.Request.Context(), agentID)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, "points balance fetched", result)
}

func (h *PointsHandler) ApplyRecharge(c *gin.Context) {
	agentID, ok := middleware.GetAgentID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("failed to resolve current user"))
		return
	}

	var request rechargeApplyRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, utils.BadRequestError(err.Error()))
		return
	}
	if request.Amount.LessThanOrEqual(decimal.Zero) {
		utils.Error(c, utils.BadRequestError("amount must be greater than 0"))
		return
	}

	result, err := h.pointsService.ApplyRecharge(c.Request.Context(), service.RechargeApplyInput{
		AgentID:       agentID,
		Amount:        request.Amount,
		PaymentMethod: request.PaymentMethod,
		PaymentProof:  request.PaymentProof,
		Remark:        request.Remark,
	})
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Created(c, "recharge request submitted", result)
}

func (h *PointsHandler) PendingRechargeRequests(c *gin.Context) {
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

	result, err := h.pointsService.PendingRechargeRequests(c.Request.Context(), service.PendingRechargeInput{
		AgentID:  agentID,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, "pending recharge requests fetched", result)
}

func (h *PointsHandler) ApproveRecharge(c *gin.Context) {
	agentID, ok := middleware.GetAgentID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("failed to resolve current user"))
		return
	}

	requestID, err := parseUintParam(c.Param("id"))
	if err != nil {
		utils.Error(c, utils.BadRequestError("invalid recharge request id"))
		return
	}

	result, err := h.pointsService.ApproveRecharge(c.Request.Context(), service.RechargeApproveInput{
		ApproverID: agentID,
		RequestID:  requestID,
	})
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, "recharge request approved", result)
}

func (h *PointsHandler) RejectRecharge(c *gin.Context) {
	agentID, ok := middleware.GetAgentID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("failed to resolve current user"))
		return
	}

	requestID, err := parseUintParam(c.Param("id"))
	if err != nil {
		utils.Error(c, utils.BadRequestError("invalid recharge request id"))
		return
	}

	result, err := h.pointsService.RejectRecharge(c.Request.Context(), service.RechargeRejectInput{
		ApproverID: agentID,
		RequestID:  requestID,
	})
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, "recharge request rejected", result)
}

func (h *PointsHandler) RechargeHistory(c *gin.Context) {
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

	result, err := h.pointsService.RechargeHistory(c.Request.Context(), service.RechargeHistoryInput{
		AgentID:  agentID,
		Page:     page,
		PageSize: pageSize,
		Status:   status,
	})
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, "recharge history fetched", result)
}

func (h *PointsHandler) Records(c *gin.Context) {
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

	var recordType *int
	if rawType := c.Query("type"); rawType != "" {
		value, err := strconv.Atoi(rawType)
		if err != nil {
			utils.Error(c, utils.BadRequestError("invalid type"))
			return
		}
		recordType = &value
	}

	result, err := h.pointsService.Records(c.Request.Context(), service.PointsRecordsInput{
		AgentID:  agentID,
		Page:     page,
		PageSize: pageSize,
		Type:     recordType,
	})
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, "points records fetched", result)
}

func (h *PointsHandler) Stats(c *gin.Context) {
	agentID, ok := middleware.GetAgentID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("failed to resolve current user"))
		return
	}

	result, err := h.pointsService.Stats(c.Request.Context(), agentID)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, "points stats fetched", result)
}
