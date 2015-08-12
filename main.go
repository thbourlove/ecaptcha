package main

import (
	"crypto/md5"
	"fmt"
	"github.com/dchest/captcha"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/thbourlove/eCaptcha/store"
	"io"
	"log"
	"net/http"
	"strconv"
)

var userStore *store.StringStore
var validationStore *store.StringStore

func newCaptcha(writer http.ResponseWriter, request *http.Request) {
	clientId := request.URL.Query().Get("clientId")
	privateKey := userStore.Get(clientId, false)
	if privateKey == "" {
		writer.WriteHeader(401)
		return
	}
	captchaId := captcha.New()
	validationStore.Set(captchaId, clientId)
	io.WriteString(writer, captchaId)
}

func reloadCaptcha(writer http.ResponseWriter, request *http.Request) {
	captchaId := request.URL.Query().Get("captchaId")
	result := captcha.Reload(captchaId)
	if result {
		io.WriteString(writer, "success")
	} else {
		io.WriteString(writer, "fail")
	}
}

func verifyCaptcha(writer http.ResponseWriter, request *http.Request) {
	captchaId := request.URL.Query().Get("captchaId")
	verifyString := request.URL.Query().Get("verifyString")

	clientId := validationStore.Get(captchaId, true)

	result := captcha.VerifyString(captchaId, verifyString)
	if result {
		privateKey := userStore.Get(clientId, false)
		md5String := fmt.Sprintf("%x", md5.Sum([]byte(privateKey+captchaId)))
		validationStore.Set(md5String, clientId)
		io.WriteString(writer, md5String)
	} else {
		io.WriteString(writer, "fail")
	}
}

func validateCaptcha(writer http.ResponseWriter, request *http.Request) {
	secret := request.URL.Query().Get("secret")
	clientId := validationStore.Get(secret, true)
	if clientId != "" {
		privateKey := userStore.Get(clientId, false)
		md5String := fmt.Sprintf("%x", md5.Sum([]byte(privateKey+secret)))
		io.WriteString(writer, md5String)
	} else {
		io.WriteString(writer, "fail")
	}
}

func imageCaptcha(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	captchaId := vars["captchaId"]
	width, err := strconv.Atoi(request.URL.Query().Get("w"))
	if err != nil {
		width = 200
	}
	height, err := strconv.Atoi(request.URL.Query().Get("h"))
	if err != nil {
		height = 100
	}
	writer.Header().Set("Content-Type", "image/png")
	captcha.WriteImage(writer, captchaId, width, height)
}

func createRedisConn(address string, db int) (c redis.Conn) {
	c, err := redis.Dial("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	c.Do("SELECT", db)
	return c
}

func initRedisCaptchaStore() {
	c := createRedisConn(":6379", 4)
	rs := store.NewBytesStore(c)
	captcha.SetCustomStore(rs)
}

func initRedisUserStore() {
	c := createRedisConn(":6379", 5)
	userStore = store.NewStringStore(c)
}

func initRedisValidationStore() {
	c := createRedisConn(":6379", 6)
	validationStore = store.NewStringStore(c)
}

func main() {
	initRedisCaptchaStore()
	initRedisUserStore()
	initRedisValidationStore()

	router := mux.NewRouter()

	router.HandleFunc("/captcha/new", newCaptcha)
	router.HandleFunc("/captcha/reload", reloadCaptcha)
	router.HandleFunc("/captcha/verify", verifyCaptcha)
	router.HandleFunc("/captcha/validate", validateCaptcha)
	router.HandleFunc("/captcha/{captchaId}.png", imageCaptcha)

	handler := cors.Default().Handler(router)
	err := http.ListenAndServe(":8888", handler)
	if err != nil {
		log.Fatal(err)
	}
}
