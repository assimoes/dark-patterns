-- multimodal is a third modality: an artifact that carries both a text body and an image. the modality
-- column is constrained on three tables, so widen each inline check to admit it. the inline checks are
-- auto-named <table>_modality_check.
ALTER TABLE artifacts DROP CONSTRAINT artifacts_modality_check;
ALTER TABLE artifacts ADD CONSTRAINT artifacts_modality_check
    CHECK (modality IN ('text', 'image', 'video', 'multimodal'));

ALTER TABLE populations DROP CONSTRAINT populations_modality_check;
ALTER TABLE populations ADD CONSTRAINT populations_modality_check
    CHECK (modality IN ('text', 'image', 'video', 'multimodal'));

ALTER TABLE prompts DROP CONSTRAINT prompts_modality_check;
ALTER TABLE prompts ADD CONSTRAINT prompts_modality_check
    CHECK (modality IN ('text', 'image', 'video', 'multimodal'));
