-- An image prompt plus a multimodal panel. The prompt's modality is 'image' so a run pins it for image
-- populations; the models carry 'image' in their modalities array so they are valid raters for it.
 
INSERT INTO prompts (name, version, modality, system_prompt, template)
VALUES (
'image', 1, 'image',
$sys$You are an expert annotator labelling screenshots of video game user interfaces for dark patterns.
You are given a fixed codebook of pattern codes. Use ONLY those codes, never invent a code.
You are shown one screenshot. Judge, for every code in the codebook, whether the interface in the image
shows that pattern. If the screenshot shows no dark pattern, return an empty list.
Reply with JSON only, no prose, no markdown.$sys$,
$tmpl$Codebook (use only these codes):
{{.Taxonomy}}
 
Look at the attached screenshot. Any text extracted from it is provided below for reference, but trust
what you see in the image first:
{{.Content}}
 
Reply with exactly this JSON shape and nothing else:
{"patterns":[{"code":"<one codebook code>","evidence":"<what in the screenshot shows it>","explanation":"<why it fits>"}]}
Use an empty patterns list if none apply.$tmpl$
)
ON CONFLICT (name, version) DO NOTHING;
 
-- A small multimodal panel. modalities includes 'image', which is what makes them valid image raters.
INSERT INTO models (family, slug, name, modalities)
VALUES
('openai', 'openai/gpt-4o', 'GPT-4o', '{text,image}'),
('google', 'google/gemini-2.5-flash', 'Gemini 2.5 Flash (vision)', '{text,image}')
ON CONFLICT (slug) DO NOTHING;
 
-- One llm annotator per model, so a run can pin them as a panel.
INSERT INTO annotators (kind, model_id, label)
SELECT 'llm', m.id, m.name || ' rater'
FROM models m
WHERE m.slug IN ('openai/gpt-4o', 'google/gemini-2.5-flash')
ON CONFLICT (label) DO NOTHING;