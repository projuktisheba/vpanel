package handlers

import (
	"log"
	"net/http"

	"github.com/projuktisheba/vpanel/backend/internal/config"
	"github.com/projuktisheba/vpanel/backend/internal/pkg/ssl"
	"github.com/projuktisheba/vpanel/backend/internal/utils"
)

type SSLHandler struct {
	infoLog  *log.Logger
	errorLog *log.Logger
}

func newSSLHandler(infoLog, errorLog *log.Logger) SSLHandler {
	return SSLHandler{
		infoLog:  infoLog,
		errorLog: errorLog,
	}
}

func (h *SSLHandler) CheckSSL(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")

	// checking the ssl certificate
	hasSSL := ssl.CheckSSL(domain)

	// ======== Build Response ========
	var response struct {
		Error     bool   `json:"error"`
		Message   string `json:"string"`
		SSLStatus bool   `json:"ssl_status"`
	}
	response.Error = false
	response.Message = "SSL Certificated checked"
	response.SSLStatus = hasSSL
	utils.WriteJSON(w, http.StatusOK, response)
}

func (h *SSLHandler) CheckAndIssueSSL(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")

	// checking the ssl certificate
	hasSSL := ssl.CheckSSL(domain)

	var err error
	if !hasSSL {
		// setup ssl certificate
		err = ssl.SetupSSL(r.Context(), domain, config.Email, true)
	}

	// ======== Build Response ========
	var response struct {
		Error     bool   `json:"error"`
		Message   string `json:"string"`
		SSLStatus bool   `json:"ssl_status"`
	}
	if err != nil {
		response.Error = true
		response.Message = err.Error()
		response.SSLStatus = false
	} else {
		response.Error = false
		response.Message = "SSL Certificated checked"
		response.SSLStatus = hasSSL
	}
	utils.WriteJSON(w, http.StatusOK, response)
}

func (h *SSLHandler) IssueSSL(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")

	// setup ssl certificate
	err := ssl.SetupSSL(r.Context(), domain, config.Email, true)

	// ======== Build Response ========
	var response struct {
		Error     bool   `json:"error"`
		Message   string `json:"string"`
		SSLStatus bool   `json:"ssl_status"`
	}
	if err != nil {
		response.Error = true
		response.Message = err.Error()
		response.SSLStatus = false
	} else {
		response.Error = false
		response.Message = "SSL Certificated checked"
		response.SSLStatus = true
	}
	utils.WriteJSON(w, http.StatusOK, response)
}
