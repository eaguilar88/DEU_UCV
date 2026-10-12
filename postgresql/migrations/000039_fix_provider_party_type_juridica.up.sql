-- 'juridical' was a typo: the party type is 'juridica', as the diplomados UI sends it.
ALTER TABLE deu.providers DROP CONSTRAINT IF EXISTS providers_party_type_check;
UPDATE deu.providers SET party_type = 'juridica' WHERE party_type = 'juridical';
ALTER TABLE deu.providers ADD CONSTRAINT providers_party_type_check
    CHECK (party_type IN ('natural', 'juridica'));
