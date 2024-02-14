package main

import (
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "log"
    "math"
    "net/http"
    "strconv"
    "strings"
    "time"
    "unicode"
)

var receipts = make(map[string]receipt)

type item struct {
    ShortDescription    string      `json:"shortDescription"`
    Price               string     `json:"price"`
}

type receipt struct {
    Retailer        string  `json:"retailer"`
    PurchaseDate    string  `json:"purchaseDate"`
    PurchaseTime    string  `json:"purchaseTime"`
    Items           []item  `json:"items"`
    Total           string `json:"total"`
    Points          int64
}


func main() {
    router := gin.Default()
   router.GET("/receipts/:id/points", getPoints)
    router.POST("/receipts/process", processReceipts)

    router.Run("localhost:8080")
}


func getPoints(c *gin.Context) {
    var id = c.Param("id")

    log.Printf("id %s", id)

    receipt, ok := receipts[id]

    if !ok {
        c.IndentedJSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
    } else {
        var response struct {
            Points      int64 `json:"points"`
        }
        response.Points = receipt.Points
        c.IndentedJSON(http.StatusCreated, response)
    }

}

func processReceipts(c *gin.Context) {
    var newReceipt receipt

    if err := c.Bind(&newReceipt); err != nil {
        log.Panic(err)
        return
    }

    id := uuid.New().String()
    newReceipt.Points = populatePoints(newReceipt)
    receipts[id] = newReceipt

    var response struct {
        Id      string `json:"id"`
    }
    response.Id = id
    c.IndentedJSON(http.StatusCreated, response)
}

func populatePoints(newReceipt receipt) int64 {
    log.Println("******************** CALCULATING POINTS *****************");
    var points int64 = 0

    points += calculatePointsInRetailerName(newReceipt);
    points += calculatePointsForRoundTotal(newReceipt);
    points += calculatePointsForQuarterTotal(newReceipt);
    points += calculatePointsForItemCount(newReceipt);
    points += calculatePointsForItemDescription(newReceipt);
    points += calculatePointsForPurchaseDay(newReceipt);
    points += calculatePurchaseTime(newReceipt);

    newReceipt.Points = points

    log.Printf("Number of points %d", points)

    return points
}

func calculatePurchaseTime(newReceipt receipt) int64 {
    var points int64 = 0

    purchaseTime, _ := time.Parse("15:04", newReceipt.PurchaseTime)
    twoPm, _ := time.Parse("15:04", "14:00")
    fourPm, _ := time.Parse("15:04", "16:00")

    if purchaseTime.After(twoPm) && purchaseTime.Before(fourPm) {
        points = 10;
        log.Printf("10 points - %s is between 2:00pm and 4:00pm", purchaseTime)
    }

    return points
}

func calculatePointsForPurchaseDay(newReceipt receipt) int64 {
    var points int64 = 0

    myDate, _ := time.Parse("2006-01-02", newReceipt.PurchaseDate)

    dayOfMonth := myDate.Day()

    if dayOfMonth % 2 == 1 {
        points = 6;
        log.Printf("6 Points - day in the purchase date is odd");
    }

    return points
}

func calculatePointsForItemDescription(newReceipt receipt) int64 {
    var points int64 = 0

    for _, item := range newReceipt.Items {
        if len(strings.Trim(item.ShortDescription, " ")) % 3 == 0 {
            priceFloat, _ := strconv.ParseFloat(item.Price, 64)
            pointsForItem := int64(math.Ceil(priceFloat * 0.2))
            points += pointsForItem
            log.Printf("3 Points - \"%s\" is %d characters (a multiple of 3)",
                strings.Trim(item.ShortDescription, " "),
                len(strings.Trim(item.ShortDescription, " ")),
            )
            log.Printf("           item price of %s * 0.2 = %f, rounded up is %d points",
                item.Price,
                priceFloat*0.2,
                pointsForItem,
            )

        }
    }

    return points
}

func calculatePointsForItemCount(newReceipt receipt) int64 {
    var points = int64(len(newReceipt.Items) / 2 * 5)

    log.Printf("%d points - %d items (X pairs @ 5 points each)", points, len(newReceipt.Items))

    return points
}

func calculatePointsForQuarterTotal(newReceipt receipt) int64 {
    var points int64 = 0

    totalFloat, _ := strconv.ParseFloat(newReceipt.Total, 64)
    totalInt := int64(totalFloat * 100)
    if totalInt % 25 == 0 {
        points = 25
        log.Printf("25 points - total is a multiple of 0.25");
    }

    return points
}


func calculatePointsForRoundTotal(newReceipt receipt) int64 {
    var points int64 = 0

    totalFloat, _ := strconv.ParseFloat(newReceipt.Total, 64)

    var totalRounded = float64(int64(totalFloat))
    if totalRounded == totalFloat {
        points = 50
        log.Printf("50 points - total is a round dollar amount");
    }
    return points
}

func calculatePointsInRetailerName(newReceipt receipt) int64 {
    var points int64 = 0

    for i := 0; i < len(newReceipt.Retailer); i++ {   //run a loop and iterate through each character
        if unicode.IsLetter(rune(newReceipt.Retailer[i])) || unicode.IsNumber(rune(newReceipt.Retailer[i])) {
            points++
        }
    }

    log.Printf("%d points - retailer name has %d characters", points, points);

    return points
}
