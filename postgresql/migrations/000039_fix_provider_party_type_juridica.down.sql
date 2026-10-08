ALTER TABLE deu.providers DROP CONSTRAINT IF EXISTS providers_party_type_check;
UPDATE deu.providers SET party_type = 'juridical' WHERE party_type = 'juridica';
ALTER TABLE deu.providers ADD CONSTRAINT providers_party_type_check
    CHECK (party_type IN ('natural', 'juridical'));
