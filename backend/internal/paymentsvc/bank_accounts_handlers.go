package paymentsvc

import (
	"net/http"
	"strings"
	"time"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type sellerBankAccountPayload struct {
	ID                  string    `json:"id"`
	SellerID            string    `json:"seller_id"`
	BankName            string    `json:"bank_name"`
	AccountName         string    `json:"account_name"`
	AccountNumberMasked string    `json:"account_number_masked"`
	Country             string    `json:"country,omitempty"`
	Currency            string    `json:"currency"`
	IsDefault           bool      `json:"is_default"`
	CreatedAt           time.Time `json:"created_at"`
}

func (s *Service) HandleListSellerBankAccounts(c *gin.Context) {
	sellerID, ok := s.sellerIDForUser(c)
	if !ok {
		apiresponse.Error(c, http.StatusForbidden, "forbidden", "Seller not found.")
		return
	}
	rows, err := s.pool.Query(c.Request.Context(), `
		SELECT id, seller_id, bank_name, account_name, account_number_masked, COALESCE(country, ''), currency, is_default, created_at
		FROM seller_bank_accounts
		WHERE seller_id = $1
		ORDER BY is_default DESC, created_at DESC
	`, sellerID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list bank accounts.")
		return
	}
	defer rows.Close()
	out := make([]sellerBankAccountPayload, 0)
	for rows.Next() {
		var item sellerBankAccountPayload
		if err := rows.Scan(
			&item.ID,
			&item.SellerID,
			&item.BankName,
			&item.AccountName,
			&item.AccountNumberMasked,
			&item.Country,
			&item.Currency,
			&item.IsDefault,
			&item.CreatedAt,
		); err != nil {
			apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not parse bank account.")
			return
		}
		out = append(out, item)
	}
	apiresponse.OK(c, gin.H{"items": out})
}

func maskAccountNumber(v string) string {
	v = strings.TrimSpace(v)
	if len(v) <= 4 {
		return "****"
	}
	return "****" + v[len(v)-4:]
}

type createSellerBankAccountRequest struct {
	BankName      string `json:"bank_name" binding:"required"`
	AccountName   string `json:"account_name" binding:"required"`
	AccountNumber string `json:"account_number" binding:"required"`
	Country       string `json:"country"`
	Currency      string `json:"currency"`
	IsDefault     bool   `json:"is_default"`
}

func (s *Service) HandleCreateSellerBankAccount(c *gin.Context) {
	sellerID, ok := s.sellerIDForUser(c)
	if !ok {
		apiresponse.Error(c, http.StatusForbidden, "forbidden", "Seller not found.")
		return
	}
	var req createSellerBankAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid bank account payload.")
		return
	}
	if strings.TrimSpace(req.Currency) == "" {
		req.Currency = "USD"
	}
	tx, err := s.pool.Begin(c.Request.Context())
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not create bank account.")
		return
	}
	defer tx.Rollback(c.Request.Context())
	if req.IsDefault {
		if _, err := tx.Exec(c.Request.Context(), `UPDATE seller_bank_accounts SET is_default = FALSE WHERE seller_id = $1`, sellerID); err != nil {
			apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not create bank account.")
			return
		}
	}
	var item sellerBankAccountPayload
	err = tx.QueryRow(c.Request.Context(), `
		INSERT INTO seller_bank_accounts (seller_id, bank_name, account_name, account_number_masked, country, currency, is_default)
		VALUES ($1, $2, $3, $4, NULLIF($5,''), $6, $7)
		RETURNING id, seller_id, bank_name, account_name, account_number_masked, COALESCE(country, ''), currency, is_default, created_at
	`, sellerID, strings.TrimSpace(req.BankName), strings.TrimSpace(req.AccountName), maskAccountNumber(req.AccountNumber), strings.TrimSpace(req.Country), strings.TrimSpace(req.Currency), req.IsDefault).Scan(
		&item.ID, &item.SellerID, &item.BankName, &item.AccountName, &item.AccountNumberMasked, &item.Country, &item.Currency, &item.IsDefault, &item.CreatedAt,
	)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not create bank account.")
		return
	}
	if err := tx.Commit(c.Request.Context()); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not create bank account.")
		return
	}
	apiresponse.Respond(c, http.StatusCreated, item, nil)
}

func (s *Service) HandleUpdateSellerBankAccount(c *gin.Context) {
	sellerID, ok := s.sellerIDForUser(c)
	if !ok {
		apiresponse.Error(c, http.StatusForbidden, "forbidden", "Seller not found.")
		return
	}
	id, err := uuid.Parse(c.Param("bank_account_id"))
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid bank account id.")
		return
	}
	var req struct {
		BankName      string `json:"bank_name"`
		AccountName   string `json:"account_name"`
		AccountNumber string `json:"account_number"`
		Country       string `json:"country"`
		Currency      string `json:"currency"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid bank account payload.")
		return
	}
	cmd, err := s.pool.Exec(c.Request.Context(), `
		UPDATE seller_bank_accounts
		SET
		  bank_name = COALESCE(NULLIF($3, ''), bank_name),
		  account_name = COALESCE(NULLIF($4, ''), account_name),
		  account_number_masked = COALESCE(NULLIF($5, ''), account_number_masked),
		  country = COALESCE(NULLIF($6, ''), country),
		  currency = COALESCE(NULLIF($7, ''), currency)
		WHERE id = $1 AND seller_id = $2
	`, id, sellerID, strings.TrimSpace(req.BankName), strings.TrimSpace(req.AccountName), maskOrEmpty(req.AccountNumber), strings.TrimSpace(req.Country), strings.TrimSpace(req.Currency))
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not update bank account.")
		return
	}
	if cmd.RowsAffected() == 0 {
		apiresponse.Error(c, http.StatusNotFound, "not_found", "Bank account not found.")
		return
	}
	apiresponse.OK(c, gin.H{"id": id.String(), "updated": true})
}

func maskOrEmpty(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	return maskAccountNumber(v)
}

func (s *Service) HandleDeleteSellerBankAccount(c *gin.Context) {
	sellerID, ok := s.sellerIDForUser(c)
	if !ok {
		apiresponse.Error(c, http.StatusForbidden, "forbidden", "Seller not found.")
		return
	}
	id, err := uuid.Parse(c.Param("bank_account_id"))
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid bank account id.")
		return
	}
	cmd, err := s.pool.Exec(c.Request.Context(), `DELETE FROM seller_bank_accounts WHERE id = $1 AND seller_id = $2`, id, sellerID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not delete bank account.")
		return
	}
	if cmd.RowsAffected() == 0 {
		apiresponse.Error(c, http.StatusNotFound, "not_found", "Bank account not found.")
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Service) HandleSetDefaultSellerBankAccount(c *gin.Context) {
	sellerID, ok := s.sellerIDForUser(c)
	if !ok {
		apiresponse.Error(c, http.StatusForbidden, "forbidden", "Seller not found.")
		return
	}
	id, err := uuid.Parse(c.Param("bank_account_id"))
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid bank account id.")
		return
	}
	tx, err := s.pool.Begin(c.Request.Context())
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not update default account.")
		return
	}
	defer tx.Rollback(c.Request.Context())
	if _, err := tx.Exec(c.Request.Context(), `UPDATE seller_bank_accounts SET is_default = FALSE WHERE seller_id = $1`, sellerID); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not update default account.")
		return
	}
	cmd, err := tx.Exec(c.Request.Context(), `UPDATE seller_bank_accounts SET is_default = TRUE WHERE id = $1 AND seller_id = $2`, id, sellerID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not update default account.")
		return
	}
	if cmd.RowsAffected() == 0 {
		apiresponse.Error(c, http.StatusNotFound, "not_found", "Bank account not found.")
		return
	}
	if err := tx.Commit(c.Request.Context()); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not update default account.")
		return
	}
	apiresponse.OK(c, gin.H{"id": id.String(), "is_default": true})
}

