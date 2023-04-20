#endpoint details
/productCatalog - will give all the product details

/orders - lists all the orders placed 

/placeOrder - endpoint used to place a new order 
   example request: 
   {
    "products": [
        {
            "id":1,
            "quantity": 1
        },
        {
            "id":4,
            "quantity": 2
        },
        {
            "id":5,
            "quantity": 3
        },
        {
            "id":1,
            "quantity": 3
        }
    ]
}

/updateOrderStatus - to update order status to Placed/Dispatched/Completed/Cancelled
example request:
{
    "order_id": 1,
    "order_status": "Dispatched"
}
