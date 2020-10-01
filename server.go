package main

import (
	"./config"
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

var conf = config.GetConfigValues()
var httpPort, httpsPort = conf.PortHTTP, conf.PortHTTPS

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Привет!")
}

func redirectToHttps(w http.ResponseWriter, r *http.Request) {
	addr, err := url.Parse("http://" + r.Host)
	if err != nil {
		log.Fatal(err)
	}
	var url string
	if os.Getenv("NODE_ENV") == "production" {
		url = fmt.Sprintf("https://%s:%d%s", addr.Hostname(), httpsPort, r.RequestURI)
	} else {
		url = fmt.Sprintf("http://%s:%d%s", addr.Hostname(), httpsPort, r.RequestURI)
	}
	http.Redirect(w, r, url, http.StatusMovedPermanently)
}

type Post struct {
	Id string
	Title string
}

var postData = make(map[string]*Post)
var client *mongo.Client
var ctx context.Context
func InsertPost(title string) {

	post := Post{
		Id: primitive.NewObjectID().String(),
		Title: title,
	}
	collection := client.Database("todolist").Collection("todo")
	insertResult, err := collection.InsertOne(context.TODO(), post)



	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Inserted post with ID:", insertResult.InsertedID)
}

func getPost(id primitive.ObjectID) {

	collection := client.Database("todolist").Collection("todo")

	filter := bson.M{"id": id}

	var post Post

	err := collection.FindOne(context.TODO(), filter).Decode(&post)

	if err != nil {

		log.Fatal(err)

	}

	fmt.Println("Found post with title ", post.Title)
}

func mongoDB() {
	var err error
	client, err = mongo.NewClient(options.Client().ApplyURI("mongodb+srv://todolistapp:DuJwxGQpdQHKUXv3@cluster0.gk5ib.mongodb.net/todolist?retryWrites=true&w=majority"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
}


func info() {
	var res []string
	arr, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}
	for _, elem := range arr {
		addr, err := elem.Addrs()
		if err != nil {
			log.Fatal(err)
		}
		for _, addr := range addr {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip.To4() != nil {
				res = append(res, ip.To4().String())
			}
		}
	}
	fmt.Println("На комьютере: ", fmt.Sprintf("http://%s:%d", res[0], httpsPort))
	fmt.Println("В локальной сети: ", fmt.Sprintf("http://%s:%d", res[1], httpsPort))
}

func getAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	var res []bson.M
	collection := client.Database("todolist").Collection("todo")
	cursor, _ := collection.Find(context.TODO(), bson.M{})
	cursor.All(context.TODO(), &res)
	body, _ := json.Marshal(res)
	w.Write(body)
}

func addElem(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var elem Post
	err := decoder.Decode(&elem)
	elem.Id = primitive.NewObjectID().String()
	if err != nil {
		fmt.Println(err)
	}
	collection := client.Database("todolist").Collection("todo")
	insertResult, err := collection.InsertOne(context.TODO(), elem)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Inserted post with ID:", insertResult.InsertedID)
	w.WriteHeader(http.StatusOK)
}

func deleteElem(w http.ResponseWriter, r *http.Request) {
	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	decoder := json.NewDecoder(r.Body)
	var elem bson.M
	err := decoder.Decode(&elem)
	fmt.Println(elem["Id"])
	collection := client.Database("todolist").Collection("todo")
	result, err := collection.DeleteOne(ctx, bson.M{"id": elem["Id"]})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(result.DeletedCount)
	}
}

func main() {
	httpURL, httpsURL := fmt.Sprintf(":%d", httpPort), fmt.Sprintf(":%d", httpsPort)
	http.HandleFunc("/", handler)
	http.HandleFunc("/api/getAll", getAll)
	http.HandleFunc("/api/addElem", addElem)
	http.HandleFunc("/api/deleteElem", deleteElem)
	go func() {
		var err error
		if os.Getenv("NODE_ENV") == "production" {
			err = http.ListenAndServeTLS(httpsURL, "./keys/cert.pem", "./keys/key.pem", nil)
		} else {
			go info()
			go mongoDB()
			err = http.ListenAndServe(httpsURL, nil)
		}
		if err != nil {
			log.Fatal(err)
		}
	}()
	err := http.ListenAndServe(httpURL, http.HandlerFunc(redirectToHttps))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
}
