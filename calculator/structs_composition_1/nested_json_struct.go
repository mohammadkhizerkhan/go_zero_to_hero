package structscomposition1

import (
	"fmt"
	"time"
)

type BaseEntity struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	BaseEntity
	Name string `json:"name"`
}

type UserResponse struct {
	BaseEntity
	Name    string  `json:"name"`
	Age     int16   `json:"age"`
	Address Address `json:"address"`
	Orders  []Order `json:"orders"`
}

type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	Pincode int32  `json:"pincode"`
}

type Order struct {
	BaseEntity
	Name     string `json:"name"`
	Quantity int16  `json:"quantity"`
	Rating   int16  `json:"rating,omitempty"`
}

func TestNestedStructs() {
	user := UserResponse{
		BaseEntity: BaseEntity{ID: "u-100"},
		Name:       "Khizer",
		Age:        30,
		Address: Address{
			Street:  "123 Main St",
			City:    "Lahore",
			Pincode: 54000,
		},
		Orders: []Order{
			{BaseEntity: BaseEntity{ID: "o-1"}, Name: "Laptop", Quantity: 1, Rating: 5},
			{BaseEntity: BaseEntity{ID: "o-2"}, Name: "Mouse", Quantity: 2},
		},
	}

	fmt.Println("the normal id is ", user.ID, " from the base Entity is ", user.BaseEntity.ID)
}
