package rest

import (
	"encoding/json"
	"fmt"
	"path"

	"github.com/aopoltorzhicky/bitfinex-api-go/pkg/models/common"
	"github.com/aopoltorzhicky/bitfinex-api-go/pkg/models/notification"
	"github.com/aopoltorzhicky/bitfinex-api-go/pkg/models/order"
	"github.com/aopoltorzhicky/bitfinex-api-go/pkg/models/tradeexecutionupdate"
)

// OrderService manages data flow for the Order API endpoint
type OrderService struct {
	requestFactory
	Synchronous
}

type OrderIDs []int
type GroupOrderIDs []int
type ClientOrderIDs [][]interface{}
type OrderOps [][]interface{}

// OrderMultiOpsRequest - data structure for constructing order multi ops request payload
type OrderMultiOpsRequest struct {
	Ops OrderOps `json:"ops"`
}

// CancelOrderMultiRequest - data structure for constructing cancel order multi request payload
type CancelOrderMultiRequest struct {
	OrderIDs       OrderIDs       `json:"id,omitempty"`
	GroupOrderIDs  GroupOrderIDs  `json:"gid,omitempty"`
	ClientOrderIDs ClientOrderIDs `json:"cid,omitempty"`
	All            int            `json:"all,omitempty"`
}

// Retrieves all of the active orders
// See https://docs.bitfinex.com/reference#rest-auth-orders for more info
func (s *OrderService) All() (*order.Snapshot, error) {
	// use no symbol, this will get all orders
	return s.getActiveOrders("")
}

// Retrieves all of the active orders with for the given symbol
// See https://docs.bitfinex.com/reference#rest-auth-orders for more info
func (s *OrderService) GetBySymbol(symbol string) (*order.Snapshot, error) {
	// use no symbol, this will get all orders
	return s.getActiveOrders(symbol)
}

// Retrieve an active order by the given ID
// See https://docs.bitfinex.com/reference#rest-auth-orders for more info
func (s *OrderService) GetByOrderId(orderID int64) (o *order.Order, err error) {
	os, err := s.All()
	if err != nil {
		return nil, err
	}
	for _, order := range os.Snapshot {
		if order.ID == orderID {
			return order, nil
		}
	}
	return nil, common.ErrNotFound
}

// Retrieves all past orders
// See https://docs.bitfinex.com/reference#orders-history for more info
func (s *OrderService) AllHistory() (*order.Snapshot, error) {
	// use no symbol, this will get all orders
	return s.getHistoricalOrders("")
}

// Retrieves all past orders with the given symbol
// See https://docs.bitfinex.com/reference#orders-history for more info
func (s *OrderService) GetHistoryBySymbol(symbol string) (*order.Snapshot, error) {
	// use no symbol, this will get all orders
	return s.getHistoricalOrders(symbol)
}

// Retrieve a single order in history with the given id
// See https://docs.bitfinex.com/reference#orders-history for more info
func (s *OrderService) GetHistoryByOrderId(orderID int64) (o *order.Order, err error) {
	os, err := s.AllHistory()
	if err != nil {
		return nil, err
	}
	for _, order := range os.Snapshot {
		if order.ID == orderID {
			return order, nil
		}
	}
	return nil, common.ErrNotFound
}

// Retrieves the trades generated by an order
// See https://docs.bitfinex.com/reference#orders-history for more info
func (s *OrderService) OrderTrades(symbol string, orderID int64) (*tradeexecutionupdate.Snapshot, error) {
	key := fmt.Sprintf("%s:%d", symbol, orderID)
	req, err := s.requestFactory.NewAuthenticatedRequest(common.PermissionRead, path.Join("order", key, "trades"))
	if err != nil {
		return nil, err
	}
	raw, err := s.Request(req)
	if err != nil {
		return nil, err
	}
	return tradeexecutionupdate.SnapshotFromRaw(raw)
}

func (s *OrderService) getActiveOrders(symbol string) (*order.Snapshot, error) {
	req, err := s.requestFactory.NewAuthenticatedRequest(common.PermissionRead, path.Join("orders", symbol))
	if err != nil {
		return nil, err
	}
	raw, err := s.Request(req)
	if err != nil {
		return nil, err
	}
	os, err := order.SnapshotFromRaw(raw)
	if err != nil {
		return nil, err
	}
	if os == nil {
		return &order.Snapshot{}, nil
	}
	return os, nil
}

func (s *OrderService) getHistoricalOrders(symbol string) (*order.Snapshot, error) {
	req, err := s.requestFactory.NewAuthenticatedRequest(common.PermissionRead, path.Join("orders", symbol, "hist"))
	if err != nil {
		return nil, err
	}
	raw, err := s.Request(req)
	if err != nil {
		return nil, err
	}
	os, err := order.SnapshotFromRaw(raw)
	if err != nil {
		return nil, err
	}
	if os == nil {
		return &order.Snapshot{}, nil
	}
	return os, nil
}

// Submit a request to create a new order
// see https://docs.bitfinex.com/reference#submit-order for more info
func (s *OrderService) SubmitOrder(onr *order.NewRequest) (*notification.Notification, error) {
	bytes, err := onr.ToJSON()
	if err != nil {
		return nil, err
	}
	req, err := s.requestFactory.NewAuthenticatedRequestWithBytes(common.PermissionWrite, path.Join("order", "submit"), bytes)
	if err != nil {
		return nil, err
	}
	raw, err := s.Request(req)
	if err != nil {
		return nil, err
	}
	return notification.FromRaw(raw)
}

// Submit a request to update an order with the given id with the given changes
// see https://docs.bitfinex.com/reference#order-update for more info
func (s *OrderService) SubmitUpdateOrder(our *order.UpdateRequest) (*notification.Notification, error) {
	bytes, err := our.ToJSON()
	if err != nil {
		return nil, err
	}
	req, err := s.requestFactory.NewAuthenticatedRequestWithBytes(common.PermissionWrite, path.Join("order", "update"), bytes)
	if err != nil {
		return nil, err
	}
	raw, err := s.Request(req)
	if err != nil {
		return nil, err
	}
	return notification.FromRaw(raw)
}

// Submit a request to cancel an order with the given Id
// see https://docs.bitfinex.com/reference#cancel-order for more info
func (s *OrderService) SubmitCancelOrder(oc *order.CancelRequest) error {
	bytes, err := oc.ToJSON()
	if err != nil {
		return err
	}
	req, err := s.requestFactory.NewAuthenticatedRequestWithBytes(common.PermissionWrite, path.Join("order", "cancel"), bytes)
	if err != nil {
		return err
	}
	_, err = s.Request(req)
	if err != nil {
		return err
	}
	return nil
}

// CancelOrderMulti cancels multiple orders simultaneously. Orders can be canceled based on the Order ID,
// the combination of Client Order ID and Client Order Date, or the Group Order ID. Alternatively, the body
// param 'all' can be used with a value of 1 to cancel all orders.
// see https://docs.bitfinex.com/reference#rest-auth-order-cancel-multi for more info
func (s *OrderService) CancelOrderMulti(args CancelOrderMultiRequest) (*notification.Notification, error) {
	bytes, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}

	req, err := s.requestFactory.NewAuthenticatedRequestWithBytes(
		common.PermissionWrite,
		path.Join("order", "cancel", "multi"),
		bytes,
	)
	if err != nil {
		return nil, err
	}

	raw, err := s.Request(req)
	if err != nil {
		return nil, err
	}

	return notification.FromRaw(raw)
}

// CancelOrdersMultiOp cancels multiple orders simultaneously. Accepts a slice of order ID's to be canceled.
// see https://docs.bitfinex.com/reference#rest-auth-order-multi for more info
func (s *OrderService) CancelOrdersMultiOp(ids OrderIDs) (*notification.Notification, error) {
	pld := OrderMultiOpsRequest{
		Ops: OrderOps{
			{
				"oc_multi",
				map[string][]int{"id": ids},
			},
		},
	}

	bytes, err := json.Marshal(pld)
	if err != nil {
		return nil, err
	}

	req, err := s.requestFactory.NewAuthenticatedRequestWithBytes(
		common.PermissionWrite,
		path.Join("order", "multi"),
		bytes,
	)
	if err != nil {
		return nil, err
	}

	raw, err := s.Request(req)
	if err != nil {
		return nil, err
	}

	return notification.FromRaw(raw)
}

// CancelOrderMultiOp cancels order. Accepts orderID to be canceled.
// see https://docs.bitfinex.com/reference#rest-auth-order-multi for more info
func (s *OrderService) CancelOrderMultiOp(orderID int) (*notification.Notification, error) {
	pld := OrderMultiOpsRequest{
		Ops: OrderOps{
			{
				"oc",
				map[string]int{"id": orderID},
			},
		},
	}

	bytes, err := json.Marshal(pld)
	if err != nil {
		return nil, err
	}

	req, err := s.requestFactory.NewAuthenticatedRequestWithBytes(
		common.PermissionWrite,
		path.Join("order", "multi"),
		bytes,
	)
	if err != nil {
		return nil, err
	}

	raw, err := s.Request(req)
	if err != nil {
		return nil, err
	}

	return notification.FromRaw(raw)
}

// OrderNewMultiOp creates new order. Accepts instance of order.NewRequest
// see https://docs.bitfinex.com/reference#rest-auth-order-multi for more info
func (s *OrderService) OrderNewMultiOp(onr order.NewRequest) (*notification.Notification, error) {
	pld := OrderMultiOpsRequest{
		Ops: OrderOps{
			{
				"on",
				onr.EnrichedPayload(),
			},
		},
	}

	bytes, err := json.Marshal(pld)
	if err != nil {
		return nil, err
	}

	req, err := s.requestFactory.NewAuthenticatedRequestWithBytes(
		common.PermissionWrite,
		path.Join("order", "multi"),
		bytes,
	)
	if err != nil {
		return nil, err
	}

	raw, err := s.Request(req)
	if err != nil {
		return nil, err
	}

	return notification.FromRaw(raw)
}

// OrderUpdateMultiOp updates order. Accepts instance of order.UpdateRequest
// see https://docs.bitfinex.com/reference#rest-auth-order-multi for more info
func (s *OrderService) OrderUpdateMultiOp(our order.UpdateRequest) (*notification.Notification, error) {
	pld := OrderMultiOpsRequest{
		Ops: OrderOps{
			{
				"ou",
				our.EnrichedPayload(),
			},
		},
	}

	bytes, err := json.Marshal(pld)
	if err != nil {
		return nil, err
	}

	req, err := s.requestFactory.NewAuthenticatedRequestWithBytes(
		common.PermissionWrite,
		path.Join("order", "multi"),
		bytes,
	)
	if err != nil {
		return nil, err
	}

	raw, err := s.Request(req)
	if err != nil {
		return nil, err
	}

	return notification.FromRaw(raw)
}

// OrderMultiOp - send Multiple order-related operations. Please note the sent object has
// only one property with a value of a slice of slices detailing each order operation.
// see https://docs.bitfinex.com/reference#rest-auth-order-multi for more info
func (s *OrderService) OrderMultiOp(ops OrderOps) (*notification.Notification, error) {
	enrichedOrderOps := OrderOps{}

	for _, v := range ops {
		if v[0] == "on" {
			o, ok := v[1].(order.NewRequest)
			if !ok {
				return nil, fmt.Errorf("Invalid type for `on` operation. Expected: order.NewRequest")
			}
			v[1] = o.EnrichedPayload()
		}

		if v[0] == "ou" {
			o, ok := v[1].(order.UpdateRequest)
			if !ok {
				return nil, fmt.Errorf("Invalid type for `ou` operation. Expected: order.UpdateRequest")
			}
			v[1] = o.EnrichedPayload()
		}

		enrichedOrderOps = append(enrichedOrderOps, v)
	}

	pld := OrderMultiOpsRequest{Ops: enrichedOrderOps}
	bytes, err := json.Marshal(pld)
	if err != nil {
		return nil, err
	}

	req, err := s.requestFactory.NewAuthenticatedRequestWithBytes(
		common.PermissionWrite,
		path.Join("order", "multi"),
		bytes,
	)
	if err != nil {
		return nil, err
	}

	raw, err := s.Request(req)
	if err != nil {
		return nil, err
	}

	return notification.FromRaw(raw)
}
