package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	common "github.com/jaydeepjr28/commons"
	pb "github.com/jaydeepjr28/commons/api"
	"github.com/jaydeepjr28/omsv2-gateway/gateway"
	"go.opentelemetry.io/otel"
	otelCodes "go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type handler struct {
	gateway gateway.OrdersGateway
}

func NewHandler(gateway gateway.OrdersGateway) *handler {
	return &handler{gateway}
}

func (h *handler) registerRoutes(router *gin.Engine) {
	router.POST("/api/customers/:customerID/orders", h.handleCreateOrder)
	router.GET("/api/customers/:customerID/orders/:orderID", h.handleGetOrder)

	// Serve the existing static pages for any route not handled by the API.
	router.NoRoute(gin.WrapH(http.FileServer(http.Dir("public"))))
}

func (h *handler) handleGetOrder(c *gin.Context) {
	customerID := c.Param("customerID")
	orderID := c.Param("orderID")

	// Create a tracer span
	tr := otel.Tracer("http")
	ctx, span := tr.Start(c.Request.Context(), fmt.Sprintf("%s %s", c.Request.Method, c.Request.RequestURI))
	defer span.End()

	o, err := h.gateway.GetOrder(ctx, orderID, customerID)
	rStatus := status.Convert(err)
	if rStatus != nil {
		span.SetStatus(otelCodes.Error, err.Error())

		if rStatus.Code() != codes.InvalidArgument {
			common.WriteError(c.Writer, http.StatusBadRequest, rStatus.Message())
			return
		}

		common.WriteError(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	common.WriteJSON(c.Writer, http.StatusOK, o)
}

func (h *handler) handleCreateOrder(c *gin.Context) {
	customerID := c.Param("customerID")

	var items []*pb.ItemsWithQuantity
	if err := common.ReadJSON(c.Request, &items); err != nil {
		common.WriteError(c.Writer, http.StatusBadRequest, err.Error())
		return
	}

	if err := validateItems(items); err != nil {
		common.WriteError(c.Writer, http.StatusBadRequest, err.Error())
		return
	}

	// Create a tracer span
	tr := otel.Tracer("http")
	ctx, span := tr.Start(c.Request.Context(), fmt.Sprintf("%s %s", c.Request.Method, c.Request.RequestURI))
	defer span.End()

	o, err := h.gateway.CreateOrder(ctx, &pb.CreateOrderRequest{
		CustomerID: customerID,
		Items:      items,
	})
	rStatus := status.Convert(err)
	if rStatus != nil {
		span.SetStatus(otelCodes.Error, err.Error())

		if rStatus.Code() != codes.InvalidArgument {
			common.WriteError(c.Writer, http.StatusBadRequest, rStatus.Message())
			return
		}

		common.WriteError(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	res := &CreateOrderRequest{
		Order:         o,
		RedirectToURL: fmt.Sprintf("http://localhost:8080/success.html?customerID=%s&orderID=%s", o.CustomerID, o.ID),
	}

	common.WriteJSON(c.Writer, http.StatusOK, res)
}

func validateItems(items []*pb.ItemsWithQuantity) error {
	if len(items) == 0 {
		return common.ErrNoItems
	}

	for _, i := range items {
		if i.ID == "" {
			return errors.New("item ID is required")
		}

		if i.Quantity <= 0 {
			return errors.New("items must have a valid quantity")
		}
	}

	return nil
}
