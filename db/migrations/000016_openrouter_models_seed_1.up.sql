INSERT INTO models (family, slug, name, modalities)
VALUES 
('openai', 'openai/gpt-4o-mini', 'GPT-4o mini', '{text}'),
('anthropic', 'anthropic/claude-3.5-haiku', 'Claude 3.5 Haiku', '{text}'),
('google', 'google/gemini-2.5-flash-lite', 'Gemini 2.5 Flash', '{text}'),
('deepseek', 'deepseek/deepseek-chat-v3.1', 'DeepSeek Chat 3.1', '{text}')
ON CONFLICT (slug) DO NOTHING;


INSERT INTO annotators (kind, model_id, label)
SELECT 'llm', m.id, m.name || 'rater'
FROM models m
WHERE m.slug IN (
'openai/gpt-4o-mini', 'anthropic/claude-3.5-haiku',
'google/gemini-2.5-flash-lite', 'deepseek/deepseek-chat-v3.1'
)
ON CONFLICT (label) DO NOTHING;

INSERT INTO prompts (name, version, modality, system_prompt, template) VALUES
  ('dsr-meso', 1, 'text',
   'You are an expert annotator labelling Steam game reviews for dark patterns in video games. '
   'You are given a fixed codebook of pattern codes. Use ONLY those codes — never invent a code. '
   'If a review shows no dark pattern, return an empty list. Reply with JSON only, no prose, no markdown.',
   E'Codebook (use only these codes):\n{{.Taxonomy}}\n\n'
   'Reply with exactly this JSON shape and nothing else:\n'
   '{"patterns":[{"code":"<one codebook code>","evidence":"<verbatim quote from the review>","explanation":"<why it fits>"}]}\n'
   'Use an empty patterns list if none apply.\n\nReview:\n{{.Content}}')
ON CONFLICT (name, version) DO NOTHING;