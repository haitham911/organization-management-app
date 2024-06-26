package controllers

import (
	"net/http"
	"organization-management-app/config"
	"organization-management-app/models"
	"os"

	"github.com/gin-gonic/gin"
	stripe "github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/price"
	"github.com/stripe/stripe-go/v72/product"
)

type ProductWithPrices struct {
	Product stripe.Product `json:"product"`
	Prices  []stripe.Price `json:"prices"`
}

func CreateProduct(c *gin.Context) {
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	config.DB.Create(&product)
	c.JSON(http.StatusOK, product)
}

func GetProductsDB(c *gin.Context) {
	var products []models.Product
	config.DB.Find(&products)
	c.JSON(http.StatusOK, products)
}
func ListProducts(c *gin.Context) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	params := &stripe.ProductListParams{}
	i := product.List(params)

	var products []stripe.Product
	for i.Next() {
		products = append(products, *i.Product())
	}

	if err := i.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, products)
}
func ListProductsWithPrices(c *gin.Context) {
	params := &stripe.ProductListParams{}
	params.Filters.AddFilter("active", "", "true")
	i := product.List(params)

	var productsWithPrices []ProductWithPrices

	for i.Next() {
		p := i.Product()

		priceParams := &stripe.PriceListParams{
			Product: stripe.String(p.ID),
		}
		priceParams.Filters.AddFilter("active", "", "true")
		priceIter := price.List(priceParams)

		var prices []stripe.Price
		for priceIter.Next() {
			prices = append(prices, *priceIter.Price())
		}

		productsWithPrices = append(productsWithPrices, ProductWithPrices{
			Product: *p,
			Prices:  prices,
		})
	}

	if err := i.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, productsWithPrices)
}
