-- +goose Up
-- +goose StatementBegin
-- Prevent double-settlement for a given (wallet, order).
CREATE UNIQUE INDEX IF NOT EXISTS idx_seller_wallet_tx_unique_order
  ON seller_wallet_transactions(seller_wallet_id, reference_type, reference_id)
  WHERE reference_type = 'order' AND reference_id IS NOT NULL;

-- Prevent double-release of the same order hold.
CREATE UNIQUE INDEX IF NOT EXISTS idx_seller_wallet_tx_unique_hold_release
  ON seller_wallet_transactions(seller_wallet_id, reference_type, reference_id)
  WHERE reference_type = 'hold_release' AND reference_id IS NOT NULL;

-- Prevent duplicate payout requests keyed by provider ref when present.
CREATE UNIQUE INDEX IF NOT EXISTS idx_seller_payouts_provider_ref
  ON seller_payouts(provider, provider_payout_id)
  WHERE provider IS NOT NULL AND provider_payout_id IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_seller_payouts_provider_ref;
DROP INDEX IF EXISTS idx_seller_wallet_tx_unique_hold_release;
DROP INDEX IF EXISTS idx_seller_wallet_tx_unique_order;
-- +goose StatementEnd
