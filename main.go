package main

import (
	"context"
	"encoding/gob"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"go.mod/controllers"
	"go.mod/modals"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	gob.Register(modals.User{})
}

func main() {
	r := httprouter.New()
	pc := controllers.NewProductController(getClient())
	uc := controllers.NewUserController(getClient())
	olc := controllers.NewOrderListController(getClient())
	corsHandler := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			h.ServeHTTP(w, r)
		})
	}

	r.POST("/user/login", uc.LogInUser)
	r.POST("/user/createUser", uc.CreateUser)
	r.POST("/user/logout", uc.LogOutUser)
	r.GET("/user/getUser", uc.GetUser)

	r.GET("/products", pc.GetAllProduct)
	r.GET("/products/:id", pc.GetProduct)

	r.POST("/orderlist/createorderlist", olc.CreateOrderList)
	r.PUT("/orderlist/updateorderlist", olc.UpdateOrderList)
	r.GET("/orderlist/fetchorderlist/:username", olc.GetOrderList)

	http.ListenAndServe(":8080", corsHandler(r))
}
func getClient() *mongo.Client {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	return client
}
