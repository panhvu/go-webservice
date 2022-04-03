package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
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
		},
		{
			"productId": 2,
			"manufacturer": "Jimmy-Jones",
			"sku": "2915559",
			"upc": "95846930",
			"pricePerUnit": "993.98",
			"quantityOnHand": 3204,
			"productName": "Hammer"
		}
	]`
	err := json.Unmarshal([]byte(productsJson), &productList)
	if err != nil {
		log.Fatal(err)
	}
}

func productHandler(w http.ResponseWriter, r *http.Request) {
	// getting ID by simple string parsing
	urlPathSegments := strings.Split(r.URL.Path, "products/")
	productID, err := strconv.Atoi(urlPathSegments[len(urlPathSegments)-1]) //convert string to int
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	product, listItemIndex := findProductByID(productID)
	if product == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// return a single product
		productJSON, err := json.Marshal(product)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(productJSON)
	case http.MethodPut:
		// update existing product in list
		var updatedProduct Product
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(bodyBytes, &updatedProduct)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// sanity check if IDs in body and url match
		if updatedProduct.ProductID != productID {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		product = &updatedProduct
		productList[listItemIndex] = *product
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
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

func findProductByID(productID int) (*Product, int) {
	for i, product := range productList {
		if product.ProductID == productID {
			return &product, i
		}
	}
	return nil, 0
}

func main() {
	http.HandleFunc("/products", productsHandler)
	http.HandleFunc("/products/", productHandler)
	http.ListenAndServe(":5000", nil)
}
