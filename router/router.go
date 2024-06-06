package router

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Group(func(public chi.Router) {
		public.Get("/", func(w http.ResponseWriter, r *http.Request) {
			res := map[string]interface{}{
				"data": "done",
			}

			render.JSON(w, r, res)
		})

		public.Get("/payment", func(w http.ResponseWriter, r *http.Request) {
			ipAddr := strings.Join([]string{
				r.Header.Get("x-forwarded-for"),
				r.RemoteAddr,
			}, ",")

			tmnCode := "EILHM5O9"
			secretKey := "SOHQLWQGJPEPEFHZAQUWXGSELDGUPADR"
			vnpUrl := "https://sandbox.vnpayment.vn/paymentv2/vpcpay.html"
			returnUrl := "http://localhost:3000/result"

			date := time.Now()
			createDate := date.Format("20060102150405")
			expireDate := date.Add(5 * time.Minute).Format("20060102150405")
			orderId := date.Format("174354")
			amount := "100000"
			bankCode := "NCB"
			orderInfo := "HEHE"
			orderType := "other"
			locale := "vn"
			if locale == "" {
				locale = "vn"
			}
			currCode := "VND"

			vnpParams := map[string]string{
				"vnp_Version":    "2.1.0",
				"vnp_Command":    "pay",
				"vnp_TmnCode":    tmnCode,
				"vnp_Locale":     locale,
				"vnp_CurrCode":   currCode,
				"vnp_TxnRef":     orderId,
				"vnp_OrderInfo":  orderInfo,
				"vnp_OrderType":  orderType,
				"vnp_Amount":     amount + "00",
				"vnp_ReturnUrl":  returnUrl,
				"vnp_IpAddr":     ipAddr,
				"vnp_CreateDate": createDate,
				"vnp_ExpireDate": expireDate,
			}

			if bankCode != "" {
				vnpParams["vnp_BankCode"] = bankCode
			}

			vnpParams = sortMap(vnpParams)

			signData := url.Values{}
			for key, value := range vnpParams {
				signData.Add(key, value)
			}

			signature := generateSignature(signData.Encode(), secretKey)
			signData.Add("vnp_SecureHash", signature)

			redirectURL := vnpUrl + "?" + signData.Encode()
			http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		})

		public.Get("/result", func(w http.ResponseWriter, r *http.Request) {
			log.Println(r)
		})
	})

	http.ListenAndServe(":3000", r)
	return r
}

func sortMap(inputMap map[string]string) map[string]string {
	keys := make([]string, 0, len(inputMap))
	for key := range inputMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	sortedMap := make(map[string]string)
	for _, key := range keys {
		sortedMap[key] = inputMap[key]
	}
	return sortedMap
}

func generateSignature(data, secretKey string) string {
	hash := hmac.New(sha512.New, []byte(secretKey))
	hash.Write([]byte(data))
	return hex.EncodeToString(hash.Sum(nil))
}
