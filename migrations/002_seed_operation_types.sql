INSERT INTO operation_types (id, description, sign) VALUES
    (1, 'NORMAL PURCHASE', -1),
    (2, 'PURCHASE WITH INSTALLMENTS', -1),
    (3, 'WITHDRAWAL', -1),
    (4, 'CREDIT VOUCHER', 1)
ON CONFLICT (id) DO NOTHING;
