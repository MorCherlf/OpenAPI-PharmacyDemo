package main

import (
	_ "Pharmacy/docs"
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"     // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// @title Pharmacy API
// @version 1.0
// @description This Server API is a simulator pharmacy.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

var requestCount int32 // Counting
var db *gorm.DB        // 数据库连接

// 初始化数据库连接和迁移
func init() {
	var err error
	db, err = gorm.Open(sqlite.Open("metrics.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 自动迁移创建 Metric 表
	db.AutoMigrate(&EndpointMetric{})
}

// Metric 数据库模型，用于存储指标
type EndpointMetric struct {
	ID           uint      `gorm:"primaryKey"`
	Endpoint     string    `json:"endpoint"`
	Timestamp    time.Time `json:"timestamp"`
	RequestCount int       `json:"request_count"`
}

// @host localhost:8080
// @BasePath /
func main() {
	r := gin.Default()

	// 中间件用于收集请求计数
	r.Use(func(c *gin.Context) {
		atomic.AddInt32(&requestCount, 1)
		recordEndpointMetric(c.FullPath()) // 记录完整路径作为 Endpoint
		c.Next()
	})

	r.GET("/metrics", getMetrics) // 添加获取指标的路由

	r.GET("/medicines", getMedicines)
	r.GET("/medicines/:id", getMedicineByID)
	r.POST("/medicines", createMedicine)
	r.PUT("/medicines/:id", updateMedicine)
	r.DELETE("/medicines/:id", deleteMedicine)

	// Use ginSwagger middleware
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.Run(":8080")
}

// 获取请求计数指标
func getMetrics(c *gin.Context) {
	var metrics []EndpointMetric
	db.Find(&metrics)
	c.JSON(http.StatusOK, metrics)
}

// 将当前请求计数记录到数据库中
func recordEndpointMetric(endpoint string) {
	metric := EndpointMetric{
		Endpoint:     endpoint,
		Timestamp:    time.Now().UTC(),
		RequestCount: 1,
	}
	result := db.Create(&metric)
	if result.Error != nil {
		fmt.Println("Failed to write metric to database:", result.Error)
	} else {
		fmt.Println("Metric written to database successfully")
	}
}

// Medicine
type Medicine struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Manufacturer string  `json:"manufacturer"`
	Price        float64 `json:"price"`
	Stock        int     `json:"stock"`
}

// Medicine datas
var medicines = []Medicine{
	{ID: 1, Name: "ABC", Manufacturer: "1234", Price: 15.5, Stock: 100},
	{ID: 2, Name: "EFG", Manufacturer: "5678", Price: 12.0, Stock: 50},
	{ID: 3, Name: "XYZ", Manufacturer: "9999", Price: 5.8, Stock: 200},
}

// @Summary Get Medicine
// @Description Get All Medicine's Data
// @ID get-medicines
// @Produce  json
// @Success 200 {array} Medicine
// @Router /medicines [get]
func getMedicines(c *gin.Context) {
	c.JSON(http.StatusOK, medicines)
}

// @Summary Get Medicine By ID
// @Description Get Medicine Data By ID
// @ID get-medicine-by-id
// @Produce  json
// @Param id path int true "Medicine ID"
// @Success 200 {object} Medicine
// @Failure 404 {object} string "Medicine is not exist"
// @Router /medicines/{id} [get]
func getMedicineByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unavailable ID"})
		return
	}

	for _, m := range medicines {
		if m.ID == id {
			c.JSON(http.StatusOK, m)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Medicine is not exist."})
}

// @Summary Create New Medicine
// @Description Create New Medicine Data
// @ID create-medicine
// @Accept  json
// @Produce  json
// @Param medicine body Medicine true "Medicine Data"
// @Success 201 {object} Medicine
// @Router /medicines [post]
func createMedicine(c *gin.Context) {
	var newMedicine Medicine
	if err := c.ShouldBindJSON(&newMedicine); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newMedicine.ID = len(medicines) + 1
	medicines = append(medicines, newMedicine)
	c.JSON(http.StatusCreated, newMedicine)
}

// @Summary Update Medicine
// @Description Update Medicine By ID
// @ID update-medicine
// @Accept  json
// @Produce  json
// @Param id path int true "Medicine ID"
// @Param medicine body Medicine true "Medicine Info"
// @Success 200 {object} Medicine
// @Failure 404 {object} string "Medicine is not exist"
// @Router /medicines/{id} [put]
func updateMedicine(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unavailable ID"})
		return
	}

	var updatedMedicine Medicine
	if err := c.ShouldBindJSON(&updatedMedicine); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for i, m := range medicines {
		if m.ID == id {
			medicines[i] = updatedMedicine
			medicines[i].ID = id
			c.JSON(http.StatusOK, medicines[i])
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Medicine is not exist"})
}

// @Summary Delete Medicine
// @Description Delete Medicine By ID
// @ID delete-medicine
// @Produce  json
// @Param id path int true "Medicine ID"
// @Success 204 "No Content"
// @Failure 404 {object} string "Medicine is not exist"
// @Router /medicines/{id} [delete]
func deleteMedicine(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unavailable ID"})
		return
	}

	for i, m := range medicines {
		if m.ID == id {
			medicines = append(medicines[:i], medicines[i+1:]...)
			c.JSON(http.StatusNoContent, gin.H{})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Medicine is not exist"})
}
