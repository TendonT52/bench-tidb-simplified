CREATE TABLE transfer(
  id bigint PRIMARY KEY auto_random,
  ulid CHAR(26),
  from_account_id CHAR(36),
  to_account_id CHAR(36),
  amount INT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_transfer_ulid (ulid, id)
);