package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type Product struct {
	ProductID      int    `json:"productId"`
	Manufacturer   string `json:"manufacturer"`
	Sku            string `json:"sku"`
	Upc            string `json:"upc"`
	PricePerUnit   string `json:"pricePerUnit"`
	QuantityOnHand int    `json:"quantityOnHand"`
	ProductName    string `json:"productName"`
}

var productList []Product

func init() {
	// local data from memory
	productsJson := `[
		{
			"productId": 1,
			"manufacturer": "Johns-Jenkins",
			"sku": "2924d9",
			"upc": "39449393",
			"pricePerUnit": "43.98",
			"quantityOnHand": 2040,
			"productName": "Product"
		}
	]`
	err := json.Unmarshal([]byte(productsJson), &productList)
	if err != nil {
		log.Fatal(err)
	}
}

func productsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGetRequest(w)

	case http.MethodPost:
		handlePostRequest(w, r)
	}
}

func handleGetRequest(w http.ResponseWriter) {
	productsJson, err := json.Marshal(productList)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(productsJson)
}

func handlePostRequest(w http.ResponseWriter, r *http.Request) {
	// add a new product to the list
	var newProduct Product
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(bodyBytes, &newProduct)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// make sure client didn't send productID, which shall be generated from server
	if newProduct.ProductID != 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	newProduct.ProductID = getNextID()
	productList = append(productList, newProduct)
	w.WriteHeader(http.StatusCreated)
	return
}

func getNextID() int {
	highestID := -1
	for _, product := range productList {
		if highestID < product.ProductID {
			highestID = product.ProductID
		}
	}
	return highestID + 1
}

func main() {
	http.HandleFunc("/products", productsHandler)
	http.ListenAndServe(":5000", nil)
}
