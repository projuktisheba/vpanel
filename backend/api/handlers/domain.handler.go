package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/projuktisheba/vpanel/backend/internal/dbrepo"
	"github.com/projuktisheba/vpanel/backend/internal/models"
	"github.com/projuktisheba/vpanel/backend/utils"
)

type DomainHandler struct {
	DB       *dbrepo.DBRepository
	infoLog  *log.Logger
	errorLog *log.Logger
}

func newDomainHandler(db *dbrepo.DBRepository, infoLog, errorLog *log.Logger) DomainHandler {
	return DomainHandler{
		DB:       db,
		infoLog:  infoLog,
		errorLog: errorLog,
	}
}

// ==================== Create Domain ====================
func (h *DomainHandler) CreateDomain(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Domain        string  `json:"domain"`
		SSLUpdateDate *string `json:"ssl_update_date,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorLog.Println("ERROR_CreateDomain: failed to decode request:", err)
		utils.BadRequest(w, errors.New("invalid JSON body"))
		return
	}

	if err := ValidateDomain(req.Domain); err != nil {
		utils.BadRequest(w, err)
		return
	}

	d := &models.Domain{
		Domain: req.Domain,
	}

	err := h.DB.Domain.CreateDomain(r.Context(), d)
	if err != nil {
		h.errorLog.Println("ERROR_CreateDomain:", err)
		utils.ServerError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.Response{
		Error:   false,
		Message: "Domain created successfully",
	})
}

// ==================== Update Domain (name + SSL) ====================
func (h *DomainHandler) UpdateDomain(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(w, errors.New("invalid domain ID"))
		return
	}

	var req struct {
		Domain        string  `json:"domain"`
		SSLUpdateDate *string `json:"ssl_update_date,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, errors.New("invalid JSON body"))
		return
	}

	if err := ValidateDomain(req.Domain); err != nil {
		utils.BadRequest(w, err)
		return
	}

	d := &models.Domain{
		ID:     id,
		Domain: req.Domain,
	}

	err = h.DB.Domain.UpdateDomain(r.Context(), d)
	if err != nil {
		h.errorLog.Println("ERROR_UpdateDomain:", err)
		utils.ServerError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.Response{
		Error:   false,
		Message: "Domain updated successfully",
	})
}

// ==================== Update Domain Name Only ====================
func (h *DomainHandler) UpdateDomainName(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("domain_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(w, errors.New("invalid domain ID"))
		return
	}

	var req struct {
		Domain string `json:"domain"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, errors.New("invalid JSON body"))
		return
	}

	if err := ValidateDomain(req.Domain); err != nil {
		utils.BadRequest(w, err)
		return
	}

	_, err = h.DB.Domain.UpdateDomainName(r.Context(), id, req.Domain)
	if err != nil {
		h.errorLog.Println("ERROR_UpdateDomainName:", err)
		utils.ServerError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.Response{
		Error:   false,
		Message: "Domain name updated successfully",
	})
}

// ==================== Delete Domain ====================
func (h *DomainHandler) DeleteDomain(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(w, errors.New("invalid domain ID"))
		return
	}

	if err := h.DB.Domain.DeleteDomain(r.Context(), id); err != nil {
		h.errorLog.Println("ERROR_DeleteDomain:", err)
		utils.ServerError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.Response{
		Error:   false,
		Message: "Domain deleted successfully",
	})
}

// ==================== List Domains ====================
func (h *DomainHandler) ListDomains(w http.ResponseWriter, r *http.Request) {
	domains, err := h.DB.Domain.ListDomains(r.Context())
	if err != nil {
		h.errorLog.Println("ERROR_ListDomains:", err)
		utils.ServerError(w, err)
		return
	}
	var resp struct {
		Error   bool             `json:"error"`
		Message string           `json:"message"`
		Domains []*models.Domain `json:"domains"`
	}
	resp.Error = false
	resp.Message = "Domains fetched successfully"
	resp.Domains = domains
	utils.WriteJSON(w, http.StatusOK, resp)
}

var domainRegex = regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[A-Za-z]{2,63}$`)

// ValidateDomain validates a domain name
func ValidateDomain(domain string) error {
	domain = strings.TrimSpace(domain)

	if domain == "" {
		return errors.New("domain cannot be empty")
	}

	if !domainRegex.MatchString(domain) {
		return errors.New("invalid domain format")
	}

	return nil
}
