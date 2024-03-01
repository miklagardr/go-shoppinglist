package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/julienschmidt/httprouter"
	"go.mod/modals"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserController struct {
	client *mongo.Client
}

func NewUserController(c *mongo.Client) *UserController {
	return &UserController{c}
}

func sessionKey() []byte {
	var sessionKey = []byte("so-secret---key")
	return sessionKey
}

var store = sessions.NewCookieStore(sessionKey())

func (uc UserController) GetUser(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	session, err := store.Get(req, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if authenticated, ok := session.Values["authenticated"].(bool); ok && authenticated {
		user, ok := session.Values["user"].(modals.User)
		if !ok {
			http.Error(w, "Invalid user data in the session", http.StatusInternalServerError)
			return
		}
		jsonUser, _ := json.Marshal(user)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonUser)

	} else {
		http.Error(w, "User is not logged in", http.StatusUnauthorized)
		return
	}
}

func (uc UserController) LogInUser(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	if req.Method == http.MethodPost {

		session, err := store.Get(req, "user-session")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if authenticated, ok := session.Values["authenticated"].(bool); ok && authenticated {
			http.Error(w, "User is  logged in", http.StatusUnauthorized)
			return
		}

		var user modals.User
		err = json.NewDecoder(req.Body).Decode(&user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var existingUser modals.User
		collection := uc.client.Database("shoppinglist").Collection("users")
		err = collection.FindOne(context.TODO(), bson.M{"username": user.Username}).Decode(&existingUser)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				http.Error(w, "There is no such a user", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
		if user.Password != existingUser.Password {
			http.Error(w, "Invalid Password", http.StatusUnauthorized)
			return
		}

		session.Values["user"] = user
		session.Values["authenticated"] = true

		err = session.Save(req, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jsonResponse := modals.User{
			Username:   user.Username,
			Email:      user.Email,
			Membership: user.Membership,
		}

		w.Header().Set("Content-Type", "application/json")
		jsonUser, _ := json.Marshal(jsonResponse)
		w.Write(jsonUser)

	} else {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}

}

func (uc UserController) LogOutUser(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	if req.Method == http.MethodPost {

		session, err := store.Get(req, "user-session")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if authenticated, ok := session.Values["authenticated"].(bool); ok && authenticated {

			for key := range session.Values {
				delete(session.Values, key)
			}

			session.Values["authenticated"] = false
			err = session.Save(req, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("User logged out successfully"))
		} else {
			http.Error(w, "User is not logged in", http.StatusUnauthorized)
			return
		}
	} else {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
}

func (uc UserController) CreateUser(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	if req.Method == http.MethodPost {
		var user modals.User

		err := json.NewDecoder(req.Body).Decode(&user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		collection := uc.client.Database("shoppinglist").Collection("users")
		var existingUser modals.User
		err = collection.FindOne(context.TODO(), bson.M{"$or": []bson.M{
			{"username": user.Username},
			{"email": user.Email},
		}}).Decode(&existingUser)

		if err == nil {
			http.Error(w, "Username or Email already exists", http.StatusConflict)
			return
		} else if err != mongo.ErrNoDocuments {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = collection.InsertOne(context.Background(), user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("User created successfully.Redirect to login page.."))

	} else {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (uc UserController) DeleteUser(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	if req.Method == http.MethodDelete {

		var user modals.User
		err := json.NewDecoder(req.Body).Decode(&user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		session, err := store.Get(req, "user-session")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for key := range session.Values {
			delete(session.Values, key)
		}

		session.Values["authenticated"] = false
		err = session.Save(req, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		collection := uc.client.Database("shoppinglist").Collection("users")
		_, err = collection.DeleteOne(context.Background(), bson.M{"username": user.Username})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte("User deleted successfully"))

	} else {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
}
