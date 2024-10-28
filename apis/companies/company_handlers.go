package companies

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

type CompanyHandler struct {
	repo CompanyRepository
}

func NewCompanyHandler(repo CompanyRepository) *CompanyHandler {
	return &CompanyHandler{repo: repo}
}

func (h *CompanyHandler) GetCompanyByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	company, err := h.repo.GetByID(context.Background(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(company)
}

func (h *CompanyHandler) CreateCompany(w http.ResponseWriter, r *http.Request) {
	var companyReq CompanyRequest
	err := json.NewDecoder(r.Body).Decode(&companyReq)
	if err != nil {
		http.Error(w, "not json", http.StatusBadRequest)
		return
	}

	validate := validator.New()
	if err := validate.Struct(companyReq); err != nil {
		http.Error(w, "request not in valid format", http.StatusBadRequest)
		log.Error().Err(err).Msg("request not in valid format")
		return
	}

	company := Company{
		Name:        companyReq.Name,
		YearFounded: companyReq.YearFounded,
	}

	err = h.repo.Create(context.Background(), &company)
	if err != nil {
		log.Error().Err(err)
		http.Error(w, "Failed to create company", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(company)

}

func (h *CompanyHandler) GetCompanies(w http.ResponseWriter, r *http.Request) {
	companies, err := h.repo.GetAll(context.Background())
	if err != nil {
		http.Error(w, "Failed to retrieve companies", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(companies); err != nil {
		http.Error(w, "Failed to encode companies to JSON", http.StatusInternalServerError)
		return
	}
}
