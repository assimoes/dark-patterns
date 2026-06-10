INSERT INTO models (family, slug, name, modalities)
VALUES
('qwen', 'qwen/qwen3-next-80b-a3b-instruct:free', 'Qwen3 Next 80B A3B', '{text}'),
('llama', 'meta-llama/llama-3.3-70b-instruct:free', 'Llama 3.3 70B', '{text}'),
('deepseek', 'deepseek/deepseek-v4-flash:free', 'DeepSeek V4 Flash', '{text}'),
('openai', 'openai/gpt-oss-120b:free', 'OpenAI OSS 120B', '{text}')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO annotators (kind, model_id, label)
SELECT 'llm', m.id, m.name || ' rater'
FROM models m
WHERE m.slug IN (
'qwen/qwen3-next-80b-a3b-instruct:free', 'meta-llama/llama-3.3-70b-instruct:free',
'deepseek/deepseek-v4-flash:free', 'openai/gpt-oss-120b:free'
)
ON CONFLICT (label) DO NOTHING;
