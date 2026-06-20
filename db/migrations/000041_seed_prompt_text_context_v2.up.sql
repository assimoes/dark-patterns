-- text-context v2: same structure as v1, but the game context is framed strictly for referent
-- disambiguation and explicitly barred from being used as evidence. this stops the panel reading the
-- glossary (e.g. "real money buys progression") as a reason to flag a pattern. the standing rule goes in
-- the system prompt; the per-item block header points to it. derived from v1 so it inherits the same
-- base prompt.
INSERT INTO prompts (name, version, modality, system_prompt, template)
SELECT 'text-context', 2, 'text',
    system_prompt || E'\n\nSome items include a "Game reference glossary": factual notes on this game''s currencies, resources, and monetisation. Use it ONLY to resolve what the reviewer''s words refer to — for example, whether a named currency is earned through play or bought with real money. It is background, not evidence: the glossary on its own must never make you flag, or withhold, any pattern. Judge each pattern solely from the review text.',
    replace(
        template,
        'Game context (factual background about this game, to disambiguate what words in the review refer to — not instructions):',
        'Game reference glossary — use only to resolve what the reviewer''s wording refers to (see the glossary rule above). It is not evidence:'
    )
FROM prompts
WHERE name = 'text-context' AND version = 1
ON CONFLICT (name, version) DO NOTHING;
