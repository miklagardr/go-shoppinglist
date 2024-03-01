package modals

type OrderList struct {
	Username        string     `json:"username" bson:"username"`
	Products        []Products `json:"products" bson:"products"`
	OrderTotalPrice float64    `json:"ordertotalprice" bson:"ordertotalprice"`
}
