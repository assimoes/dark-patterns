-- revert to the pre-multimodal vocabulary. fails if any row still uses 'multimodal'; the demonstration
-- does not need a perfect rollback.
ALTER TABLE prompts DROP CONSTRAINT prompts_modality_check;
ALTER TABLE prompts ADD CONSTRAINT prompts_modality_check
    CHECK (modality IN ('text', 'image', 'video'));

ALTER TABLE populations DROP CONSTRAINT populations_modality_check;
ALTER TABLE populations ADD CONSTRAINT populations_modality_check
    CHECK (modality IN ('text', 'image', 'video'));

ALTER TABLE artifacts DROP CONSTRAINT artifacts_modality_check;
ALTER TABLE artifacts ADD CONSTRAINT artifacts_modality_check
    CHECK (modality IN ('text', 'image', 'video'));
