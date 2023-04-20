package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Product struct represents a product in the catalog
type Product struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Category     string  `json:"category"`
	Price        float64 `json:"price"`
	Availability int     `json:"availability"`
	Quantity     int     `json:"quantity"`
}

// Order struct represents an order
type Order struct {
	ID           int       `json:"order_id"`
	Products     []Product `json:"products"`
	OrderValue   float64   `json:"order_value"`
	DispatchDate time.Time `json:"dispatch_date,omitempty"`
	OrderStatus  string    `json:"order_status"`
}

// ProductCatalog is an in-memory map that stores the products in the catalog
var ProductCatalog = map[int]Product{
	1: {ID: 1, Name: "Product1", Category: "Premium", Price: 100.0, Availability: 10},
	2: {ID: 2, Name: "Product2", Category: "Regular", Price: 150.0, Availability: 20},
	3: {ID: 3, Name: "Product3", Category: "Budget", Price: 200.0, Availability: 30},
	4: {ID: 4, Name: "Product4", Category: "Premium", Price: 100.0, Availability: 50},
	5: {ID: 5, Name: "Product5", Category: "Premium", Price: 90.0, Availability: 25},
	6: {ID: 6, Name: "Product6", Category: "Budget", Price: 200.0, Availability: 15},
}

// OrderMap is an in-memory map that stores the orders
var OrderMap = make(map[int]Order)

// Mutex to lock and unlock the OrderMap during concurrent access
var mutex = &sync.Mutex{}

// getProductCatalogHandler is a http handler function that returns the product catalog
func getProductCatalogHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Convert the ProductCatalog map to a slice of Products
		var products []Product
		for _, v := range ProductCatalog {
			products = append(products, v)
		}

		// Encode the slice of Products as JSON and write to the http response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(products)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getOrderCatalogHandler is a http handler function that returns the order catalog
func getOrderCatalogHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Convert the OrderMap map to a slice of Products
		var orders []Order
		for _, v := range OrderMap {
			orders = append(orders, v)
		}

		// Encode the slice of Products as JSON and write to the http response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(orders)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// placeOrderHandler is a http handler function that places an order and updates the product catalog accordingly
func placeOrderHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var premiumCount int
		products := make([]Product, 0)
		var orderValue float64
		// Parse the request body into an Order struct
		var order Order
		err := json.NewDecoder(r.Body).Decode(&order)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		for _, p := range order.Products {
			// Check if the product is available in the catalog
			product, ok := ProductCatalog[p.ID]
			if !ok {
				http.Error(w, "Product not found in catalog", http.StatusBadRequest)
				return
			}

			// Check if the requested quantity is available for the product
			if p.Quantity > product.Availability || p.Quantity > 10 {
				http.Error(w, "Requested quantity not available for product", http.StatusBadRequest)
				return
			}

			// Calculate the order value
			orderValue = orderValue + float64(p.Quantity)*product.Price
			// Check if the order contains 3 premium different products to apply discount
			if product.Category == "Premium" {
				premiumCount++
				if premiumCount >= 3 {
					orderValue *= 0.9
				}
			}

			// Update the availability of the product in the catalog
			ProductCatalog[p.ID] = Product{
				ID:           product.ID,
				Name:         product.Name,
				Category:     product.Category,
				Price:        product.Price,
				Availability: product.Availability - p.Quantity,
			}
			products = append(products, ProductCatalog[p.ID])
		}
		// Add the order to the OrderMap
		mutex.Lock()
		id := len(OrderMap) + 1
		OrderMap[id] = Order{
			ID:          id,
			Products:    products,
			OrderValue:  orderValue,
			OrderStatus: "Placed",
		}
		order.ID = id
		order.OrderStatus = "Placed"
		order.Products = products
		order.OrderValue = orderValue
		mutex.Unlock()

		// Encode the Order struct as JSON and write to the http response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(order)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func updateOrderStatusHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		// Parse the request body into an Order struct
		var o Order
		err := json.NewDecoder(r.Body).Decode(&o)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Check if the order exists in the OrderMap
		mutex.Lock()
		order, ok := OrderMap[o.ID]
		if !ok {
			mutex.Unlock()
			http.Error(w, "Order not found", http.StatusBadRequest)
			return
		}

		// Update the order status and dispatch date if necessary
		if o.OrderStatus == "Dispatched" {
			order.DispatchDate = time.Now() // Set the dispatch date to today's date
		}
		order.OrderStatus = o.OrderStatus
		OrderMap[order.ID] = order
		mutex.Unlock()

		// Encode the Order struct as JSON and write to the http response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(order)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	// Initialize the http router
	router := http.NewServeMux()

	// Register the http handlers
	router.HandleFunc("/productCatalog", getProductCatalogHandler)
	router.HandleFunc("/orders", getOrderCatalogHandler)
	router.HandleFunc("/placeOrder", placeOrderHandler)
	router.HandleFunc("/updateOrderStatus", updateOrderStatusHandler)

	// Start the http server
	fmt.Println("Server listening on port 8080...")
	http.ListenAndServe(":8080", router)
}
