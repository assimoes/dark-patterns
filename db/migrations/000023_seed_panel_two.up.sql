INSERT INTO models (family, slug, name, modalities)
VALUES
('qwen', 'qwen/qwen3-next-80b-a3b-instruct', 'Qwen3 Next 80B A3B', '{text}'),
('llama', 'meta-llama/llama-3.3-70b-instruct', 'Llama 3.3 70B', '{text}'),
('deepseek', 'deepseek/deepseek-v4-flash', 'DeepSeek V4 Flash', '{text}'),
('openai', 'openai/gpt-oss-120b', 'OpenAI OSS 120B', '{text}')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO annotators (kind, model_id, label)
SELECT 'llm', m.id, m.name || ' rater'
FROM models m
WHERE m.slug IN (
'qwen/qwen3-next-80b-a3b-instruct', 'meta-llama/llama-3.3-70b-instruct',
'deepseek/deepseek-v4-flash', 'openai/gpt-oss-120b'
)
ON CONFLICT (label) DO NOTHING;
