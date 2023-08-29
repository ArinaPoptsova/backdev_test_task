package controllers

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	models "backdev_test_task/models"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var c = context.TODO()

func init_db() (collection *mongo.Collection) {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGODB"))
	client, err := mongo.Connect(c, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(c, nil)
	if err != nil {
		log.Fatal(err)
	}
	collection = client.Database("test_task").Collection("sessions")
	return collection
}

var collection *mongo.Collection = init_db()
var validate = validator.New()

var SECRET_KEY string = os.Getenv("SECRET_KEY")

type Details struct {
	User_id uuid.UUID
	jwt.StandardClaims
}

type FullToken struct {
	Refresh_token string `json:"refresh_token"`
	Token         string `json:"access_token"`
}

func NewToken(user_id uuid.UUID) (FullToken, error) {
	claims := &Details{
		User_id: user_id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS512, claims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Panic(err)
		fmt.Println(err)
	}
	refreshToken := base64.StdEncoding.EncodeToString([]byte(uuid.New().String()))
	var session models.Session
	session.ID = primitive.NewObjectID()
	session.User_id = user_id
	session.Refresh_token = &refreshToken
	session.Is_active = true
	_, err = collection.InsertOne(c, session)
	if err != nil {
		log.Fatal(err)
	}
	return FullToken{
		Refresh_token: *session.Refresh_token,
		Token:         token,
	}, err
}

func CreateToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user_id uuid.UUID = uuid.MustParse(c.Param("user_id"))
		token, _ := NewToken(user_id)
		fmt.Println(SECRET_KEY)
		c.JSON(http.StatusOK, token)
	}
}

func RefreshToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		refreshToken := c.Param("refresh_token")
		var session models.Session
		err := collection.FindOne(context.TODO(), bson.D{{"refresh_token", refreshToken}}).Decode(&session)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		token, _ := NewToken(session.User_id)
		if session.Is_active == false {
			c.JSON(http.StatusForbidden, gin.H{"error": "This refresh token is unactive"})
			return
		}
		res, err := collection.UpdateOne(
			context.TODO(),
			bson.D{{"_id", session.ID}},
			bson.D{{"$set", bson.D{{"is_active", false}}}},
		)
		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{"error": err})
		}
		fmt.Println(res)
		c.JSON(http.StatusOK, token)
	}
}
