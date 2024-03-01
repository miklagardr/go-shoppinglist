package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"go.mod/modals"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderListController struct {
	client *mongo.Client
}

func NewOrderListController(c *mongo.Client) *OrderListController {
	return &OrderListController{c}
}
func (olc OrderListController) CreateOrderList(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	if req.Method == http.MethodPost {
		var orderList modals.OrderList
		err := json.NewDecoder(req.Body).Decode(&orderList)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		collection := olc.client.Database("shoppinglist").Collection("orderlist")
		_, err = collection.InsertOne(context.Background(), orderList)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		jsonOrderList, err := json.Marshal(orderList)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonOrderList)

	} else {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
func (olc OrderListController) UpdateOrderList(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	if req.Method == http.MethodPut {
		var orderList modals.OrderList
		err := json.NewDecoder(req.Body).Decode(&orderList)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		collection := olc.client.Database("shoppinglist").Collection("orderlist")
		filter := bson.M{"username": orderList.Username}
		update := bson.M{"$set": bson.M{
			"products":        orderList.Products,
			"ordertotalprice": orderList.OrderTotalPrice,
		}}

		if len(orderList.Products) == 0 {

			_, err := collection.DeleteOne(context.Background(), filter)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			return
		}

		result, err := collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		jsonResult, err := json.Marshal(result)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResult)

	} else {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
func (olc OrderListController) GetOrderList(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
	if req.Method == http.MethodGet {
		username := p.ByName("username")
		filter := bson.M{"username": username}
		var orderlist modals.OrderList
		collection := olc.client.Database("shoppinglist").Collection("orderlist")
		err := collection.FindOne(context.Background(), filter).Decode(&orderlist)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				http.Error(w, "No order list or login", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
		jsonOrder, _ := json.Marshal(orderlist)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonOrder)

	} else {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
